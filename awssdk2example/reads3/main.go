package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ReadCredential() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client := s3.NewFromConfig(cfg)
	input := s3.GetObjectInput{
		Bucket:              aws.String("buckentname"),
		Key:                 aws.String("data/data2/data3.csv"),
		ExpectedBucketOwner: aws.String("123456789012"),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	value, readerr := ReadCsv(output.Body)
	if readerr != nil {
		log.Fatalf("fail to read csv : %v", readerr)
	}
	log.Println(value)
}

func main() {
	lambda.Start(ReadCredential)
}

func ReadCsv(f io.ReadCloser) ([][]string, error) {

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}
