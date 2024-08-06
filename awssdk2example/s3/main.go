package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"

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

	file, err := os.Open("./input/test.csv")
	check(err)
	defer file.Close()
	putinput := s3.PutObjectInput{
		Bucket: aws.String("testbucket"),
		Key:    aws.String("s3test/test.csv"),
		Body:   file,
	}

	putoutput, err := client.PutObject(context.Background(), &putinput)
	check(err)
	log.Println(*putoutput.VersionId)

	getinput := s3.GetObjectInput{
		Bucket:              aws.String("testbucket"),
		Key:                 aws.String("s3test/test.csv"),
		ExpectedBucketOwner: aws.String("123456789012"),
	}
	getoutput, err := client.GetObject(context.Background(), &getinput)
	check(err)
	file2 := getoutput.Body
	lines, err := csv.NewReader(file2).ReadAll()
	if err != nil {
		log.Println(err.Error())
	}
	for _, data := range lines {
		log.Println(data[0])
		log.Println(data[1])
	}
}

func main() {
	ReadCredential()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ReadCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}
