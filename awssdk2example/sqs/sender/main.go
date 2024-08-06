package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func handler(ctx context.Context) error {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	// Get URL of queue
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String("jhoh-tf-queue.fifo"),
	}

	result, err := GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return err
	}

	queueURL := result.QueueUrl
	for i := 0; i < 20; i++ {

		sMInput := &sqs.SendMessageInput{
			//DelaySeconds: 90, FIFO에선 설정 안함
			MessageAttributes: map[string]types.MessageAttributeValue{
				"Title": {
					DataType:    aws.String("String"),
					StringValue: aws.String("The Whistler" + strconv.Itoa(i)),
				},
				"Author": {
					DataType:    aws.String("String"),
					StringValue: aws.String("John Grisham" + strconv.Itoa(i)),
				},
				"WeeksOn": {
					DataType:    aws.String("Number"),
					StringValue: aws.String(strconv.Itoa(i)),
				},
			},
			MessageBody:    aws.String("Information about the NY Times fiction bestseller for the week of 12/11/" + strconv.Itoa(i) + "."),
			QueueUrl:       queueURL,
			MessageGroupId: aws.String("test1234"),
		}

		resp, err := SendMsg(context.TODO(), client, sMInput)
		if err != nil {
			fmt.Println("Got an error sending the message:")
			fmt.Println(err)

		}

		fmt.Println("Sent message with ID: " + *resp.MessageId)

	}
	return nil
}

func main() {
	lambda.Start(handler)
}

// SQSSendMessageAPI defines the interface for the GetQueueUrl and SendMessage functions.
// We use this interface to test the functions using a mocked service.
type SQSSendMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// GetQueueURL gets the URL of an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a GetQueueUrlOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to GetQueueUrl.
func GetQueueURL(c context.Context, api SQSSendMessageAPI, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}

// SendMsg sends a message to an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a SendMessageOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to SendMessage.
func SendMsg(c context.Context, api SQSSendMessageAPI, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}
