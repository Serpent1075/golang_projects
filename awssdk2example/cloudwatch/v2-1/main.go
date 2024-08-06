package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func getMetric(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	client := cloudwatch.NewFromConfig(cfg)

	metricName := flag.String("mN", "CPUUtilization", "The name of the metric")
	namespace := flag.String("n", "AWS/EC2", "The namespace for the metric")
	dimensionName := flag.String("dn", "InstanceId", "The name of the dimension")
	dimensionValue := flag.String("dv", "i-00c4da667a6169f43", "The value of the dimension")
	id := flag.String("id", "i00c4da667a6169f43_cpumetric", "A short name used to tie this object to the results in the response")
	diffInMinutes := flag.Int("dM", 35, "The difference in minutes for which the metrics are required")
	stat := flag.String("s", "Maximum", "The statistic to to return")
	period := flag.Int("p", 300, "The granularity, in seconds, of the returned data points")
	flag.Parse()

	if *metricName == "" || *namespace == "" || *dimensionName == "" || *dimensionValue == "" || *id == "" || *diffInMinutes == 0 || *stat == "" || *period == 0 {
		log.Println("You must supply a metricName, namespace, dimensionName, dimensionValue, id, diffInMinutes, stat, period")
		return
	}
	var demensionlist []types.Dimension = make([]types.Dimension, 0)
	var metricdataquerylist []types.MetricDataQuery = make([]types.MetricDataQuery, 0)
	var metricdataquery types.MetricDataQuery
	var demension types.Dimension

	demension = types.Dimension{
		Name:  aws.String(*dimensionName),
		Value: aws.String(*dimensionValue),
	}
	log.Printf("dimensionName: %s", *dimensionName)
	log.Printf("dimensionValue: %s", *dimensionValue)
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
	log.Printf("namespace: %s", *namespace)
	log.Printf("metricname: %s", *metricName)
	endtime := time.Unix(time.Now().Add(time.Duration(-5)*time.Minute).Unix(), 0)
	starttime := time.Unix(time.Now().Add(time.Duration(-*diffInMinutes)*time.Minute).Unix(), 0)
	log.Print(endtime)
	log.Print(starttime)
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

	log.Printf("Metric Data: %v", result)
	log.Print(len(result.MetricDataResults))
	for _, set := range result.MetricDataResults {
		log.Println("logdata")
		log.Printf("%s, %s, %s", *set.Id, *set.Label, set.StatusCode)
		for i := 0; i < len(set.Timestamps); i++ {
			log.Printf("%v , %f", set.Timestamps[i], set.Values[i])
		}
	}
}

func main() {
	lambda.Start(getMetric)
}
