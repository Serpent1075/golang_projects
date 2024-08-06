package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type ClientInstanceType struct {
	instancetype       string
	instancegeneration string
}

func DescribeRecord(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
	)
	check(err)

	client := ec2.NewFromConfig(cfg)
	input := ec2.DescribeInstancesInput{}
	output, err := client.DescribeInstances(context.TODO(), &input)
	check(err)
	list := output.Reservations[0].Instances[0].InstanceType.Values()
	for i := 0; i < len(list); i++ {
		temp := strings.Fields(string(list[i]))
		for j := 0; j < 10; j++{
			if temp[j] == "." {
				break
			} else if 
		}
	}
	fmt.Println(list)
	for _, data := range output.Reservations {
		for _, data1 := range data.Instances {
			//fmt.Println(data1.InstanceType.Values())
		}
	}
}

func main() {
	DescribeRecord(context.TODO())
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
