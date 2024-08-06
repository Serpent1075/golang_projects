package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
	Capacity             int    `json:"Capacity"`
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
	log.Print("V2.6")
	log.Printf("Autoscaling group name: %s", event.AutoScalingGroupName)
	log.Printf("Cause : %s", event.Cause)
	log.Printf("Capacity : %v", event.CapacityToTerminate)
	log.Printf("Instances : %v", event.Instances)
	log.Printf("len capcacity: %d", len(event.CapacityToTerminate))
	log.Printf("len instance: %d", len(event.Instances))
	var response MyResponse
	response.InstanceId = make([]string, 0)
	var number int
	for _, set := range event.CapacityToTerminate {
		log.Printf("Availability Zone : %v", set.AvailabilityZone)
		log.Printf("Capacity : %v", set.Capacity) // 종료해야하는 인스턴스 수
		log.Printf("InstanceMarketOption : %v", set.InstanceMarketOption)
		number += set.Capacity
	}
	log.Printf("number: %d", number)
	for i, set := range event.Instances {
		if i == number {
			break
		}
		response.InstanceId = append(response.InstanceId, set.InstanceId)
		log.Printf("Availability Zone : %v", set.AvailabilityZone)
		log.Printf("Instance Id : %v", set.InstanceId)
		log.Printf("Marget Option : %v", set.InstanceMarketOption)
		log.Printf("Instance Type : %v", set.InstanceType)
		i++
	}
	log.Printf("response : %v", response)
	return response, nil
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
