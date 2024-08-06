package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	cfg := GetConfig()
	client := ec2.NewFromConfig(*cfg)
	input := ec2.DescribeVolumesInput{}
	output, err := client.DescribeVolumes(context.TODO(), &input)
	check(err)
	fmt.Println(len(output.Volumes))
	for _, data := range output.Volumes {
		fmt.Println("////////////////////////////////////")
		fmt.Println(data.State) // 이거 써야함
		for _, data := range data.State.Values() {
			fmt.Println(data)
		}
		if data.State == data.State.Values()[1] {
			fmt.Println("instance is available")
		}
		fmt.Println(len(data.Attachments))
		for _, data1 := range data.Attachments {
			fmt.Println("-------------")
			fmt.Println(*data1.InstanceId)
			fmt.Println(data1.State.Values())
			fmt.Println(data1.State.Values()[0])
			fmt.Println(data1.State.Values()[1])
			fmt.Println(data1.State.Values()[2])
			fmt.Println(data1.State.Values()[3])
			fmt.Println(data1.State.Values()[4])
		}
		fmt.Println(*data.Encrypted)
	}
}

func GetConfig() *aws.Config {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("ACCESSKEY", "SECRETKEY", ""),
		),
	)
	check(err)
	return &cfg
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
