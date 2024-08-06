// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX - License - Identifier: Apache - 2.0
package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
)

func getMetric() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := cloudwatch.NewFromConfig(cfg)
	name := "AWS/EC2"
	input := &cloudwatch.ListMetricsInput{
		Namespace: aws.String(name),
	}

	result, err := client.ListMetrics(context.TODO(), input)
	if err != nil {
		fmt.Println("Could not get metrics")
		return
	}

	fmt.Println("Metrics:")
	numMetrics := 0

	for _, m := range result.Metrics {
		fmt.Println("   Metric Name: " + *m.MetricName)
		fmt.Println("   Namespace:   " + *m.Namespace)
		fmt.Println("   Dimensions:")
		for _, d := range m.Dimensions {
			fmt.Println("      " + *d.Name + ": " + *d.Value)
		}

		fmt.Println("")
		numMetrics++
	}

	fmt.Println("Found " + strconv.Itoa(numMetrics) + " metrics")
}


func main() {
	lambda.Start(getMetric)
}
