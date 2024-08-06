package main

import (
	"context"
	"strings"

	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

var cfg aws.Config

type AutoScalingIn struct {
	AutoScalingGroupName string                `json:"AutoScalingGroupName"`
	Instances            []Instances           `json:"Instances"`
	CapacityToTerminate  []CapacityToTerminate `json:"CapacityToTerminate"`
	Cause                string                `json:"Cause"`
}

type CapacityToTerminate struct {
	AvailabilityZone     string `json:"AvailabilityZone"`
	InstanceMarketOption string `json:"InstanceMarketOption"`
	Capacity             int32  `json:"Capacity"`
}

type Instances struct {
	AvailabilityZone     string `json:"AvailabilityZone"`
	InstanceId           string `json:"InstanceId"`
	InstanceMarketOption string `json:"InstanceMarketOption"`
	InstanceType         string `json:"InstanceType"`
}

type MyResponse struct {
	InstanceId []string `json:"InstanceIDs"`
}

func handler(ctx context.Context, event AutoScalingIn) (MyResponse, error) {
	log.Print("V4.2")
	log.Printf("Autoscaling group name: %s", event.AutoScalingGroupName)
	log.Printf("Cause : %s", event.Cause)
	log.Printf("Capacity : %v", event.CapacityToTerminate)
	log.Printf("Instances : %v", event.Instances)

	var number int32
	var numberofdeletableinstance int32
	var deletableInstance []string
	var response MyResponse
	for _, set := range event.CapacityToTerminate {
		number += set.Capacity
	}
	log.Printf("number: %d", number)
	//줄여도 되는 인스턴스 수
	numberofdeletableinstance = 6
	deletableInstance = []string{"i-00a4638------", "i-0b24c4b------", "i-05a8818------", "i-05fa57-------", "i-04bd------", "i-0f032c------"}
	////////////////////////////
	switch {
	case number > numberofdeletableinstance:
		log.Print("Not enough empty Instances to delete")
	case number <= numberofdeletableinstance:
		response = appendInstance(event, ctx, deletableInstance)
	}

	return response, nil
}

func getMetricOfInstance(instanceid string, ctx context.Context, client *cloudwatch.Client) string {
	metricName := aws.String("CPUUtilization")
	namespace := aws.String("AWS/EC2")
	dimensionName := aws.String("InstanceId")
	dimensionValue := aws.String(instanceid)
	id := aws.String(strings.Replace(instanceid, "i-", "i", 1))
	diffInMinutes := aws.Int(30)
	stat := aws.String("Maximum")
	period := aws.Int(300)

	if *metricName == "" || *namespace == "" || *dimensionName == "" || *dimensionValue == "" || *id == "" || *diffInMinutes == 0 || *stat == "" || *period == 0 {
		log.Println("You must supply a metricName, namespace, dimensionName, dimensionValue, id, diffInMinutes, stat, period")
		return ""
	}
	var demensionlist []types.Dimension = make([]types.Dimension, 0)
	var metricdataquerylist []types.MetricDataQuery = make([]types.MetricDataQuery, 0)
	var metricdataquery types.MetricDataQuery
	var demension types.Dimension

	demension = types.Dimension{
		Name:  aws.String(*dimensionName),
		Value: aws.String(*dimensionValue),
	}
	demensionlist = append(demensionlist, demension)

	metricdataquery = types.MetricDataQuery{
		Id: aws.String(*id),
		MetricStat: &types.MetricStat{
			Metric: &types.Metric{
				Namespace:  aws.String(*namespace),
				MetricName: aws.String(*metricName),
				Dimensions: demensionlist,
			},
			Period: aws.Int32(int32(*period)),
			Stat:   aws.String(*stat),
		},
		ReturnData: aws.Bool(true),
	}
	endtime := time.Unix(time.Now().Add(time.Duration(-5)*time.Minute).Unix(), 0)
	starttime := time.Unix(time.Now().Add(time.Duration(-*diffInMinutes)*time.Minute).Unix(), 0)
	metricdataquerylist = append(metricdataquerylist, metricdataquery)
	input := &cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(endtime),
		StartTime:         aws.Time(starttime),
		MetricDataQueries: metricdataquerylist,
		LabelOptions: &types.LabelOptions{
			Timezone: aws.String("+0900"),
		},
	}

	result, err := client.GetMetricData(ctx, input)
	if err != nil {
		log.Println("Could not fetch metric data")
		log.Printf("err : %s", err.Error())
	}

	for _, set := range result.MetricDataResults {
		log.Println("logdata")
		log.Printf("%s, %s, %s", *set.Id, *set.Label, set.StatusCode)
		for i := 0; i < len(set.Timestamps); i++ {
			log.Printf("%v , %f", set.Timestamps[i], set.Values[i])
			if set.Values[i] > 1.0 {
				return ""
			}
		}
	}
	return instanceid
}

func appendInstance(event AutoScalingIn, context context.Context, deletableInstance []string) MyResponse {
	var response MyResponse
	response.InstanceId = make([]string, 0)
	client := cloudwatch.NewFromConfig(cfg)
	log.Print("appendInstance")
	for _, set := range event.CapacityToTerminate { // 각 가용영역
		for i := 0; i < int(set.Capacity); i++ { // 해당 가용영역에서 축소해야하는 용량
			for _, set2 := range event.Instances { // 오토스케일링에 등록된 모든 인스턴스들
				if set.AvailabilityZone == set2.AvailabilityZone { //가용영역에서 축소되어야할 인스턴스와 인스턴스들 중 해당 가용영역에 할당되어 있는 인스턴스만
					for _, set3 := range deletableInstance { //비교된 인스턴스가 삭제 가능한지 여부 확인
						if set2.InstanceId == set3 {
							result := getMetricOfInstance(set2.InstanceId, context, client)
							if result != "" {
								response.InstanceId = append(response.InstanceId, set2.InstanceId)
							}
						}
					}
					if int(set.Capacity) != len(response.InstanceId) {
						log.Printf("Not enough empty Instace to delete in %s", set.AvailabilityZone)
					}
				}
			}
		}
	}
	return response
}

func init() {
	var cfgerr error
	cfg, cfgerr = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if cfgerr != nil {
		log.Fatalf("failed to load configuration, %v", cfgerr.Error())
	}

}

func main() {
	lambda.Start(handler)
}

/*

func (c *Client) UpdateAutoScalingGroup(ctx context.Context, params *UpdateAutoScalingGroupInput, optFns ...func(*Options)) (*UpdateAutoScalingGroupOutput, error) {
	if params == nil {
		params = &UpdateAutoScalingGroupInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateAutoScalingGroup", params, optFns, c.addOperationUpdateAutoScalingGroupMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateAutoScalingGroupOutput)
	out.ResultMetadata = metadata
	return out, nil
}

func (c *AutoScaling) TerminateInstanceInAutoScalingGroupRequest(input *TerminateInstanceInAutoScalingGroupInput) (req *request.Request, output *TerminateInstanceInAutoScalingGroupOutput) {
	op := &request.Operation{
		Name:       opTerminateInstanceInAutoScalingGroup,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &TerminateInstanceInAutoScalingGroupInput{}
	}

	output = &TerminateInstanceInAutoScalingGroupOutput{}
	req = c.newRequest(op, input, output)
	err := req.Send()
	if err == nil { // resp is now filled
		fmt.Println(resp)
	}
}
*/
