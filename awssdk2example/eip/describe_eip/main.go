package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func DescribeRecord(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("AKID", "SecretKey", ""),
		),
	)

	client := ec2.NewFromConfig(cfg)

	result, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		fmt.Println("get ip address error")
		fmt.Println(err.Error())
	}
	for _, data := range result.Addresses {
		fmt.Println(*data.PublicIp)
		if data.AssociationId != nil {
			fmt.Println(*data.AssociationId)
		} else {
			fmt.Println("not using")
		}

	}

}

func main() {
	DescribeRecord(context.TODO())
}
