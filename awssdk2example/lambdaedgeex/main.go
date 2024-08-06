package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func handler(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		log.Println("get ip address error")
		log.Println(err.Error())
	}
	for _, data := range result.Addresses {
		log.Println(*data.PublicIp)
		if data.AssociationId != nil {
			log.Println(*data.AssociationId)
		} else {
			log.Println("not using")
		}

	}
}

func main() {
	lambda.Start(handler)
}
