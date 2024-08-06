package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	costextype "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

func GetRecommendRI(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	client := costexplorer.NewFromConfig(cfg)
	input := costexplorer.GetReservationPurchaseRecommendationInput{
		Service:      aws.String("Amazon Elastic Compute Cloud - Compute"),
		AccountId:    aws.String("123456789012"),
		AccountScope: costextype.AccountScopeLinked.Values()[1],
	}
	output, err := client.GetReservationPurchaseRecommendation(ctx, &input)
	check(err)
	fmt.Println(len(output.Recommendations))
	for _, data := range output.Recommendations {
		fmt.Println("-----------------------------------------------------")
		fmt.Println(data.AccountScope)
		fmt.Println(data.PaymentOption.Values())
		fmt.Println(len(data.RecommendationDetails))
		for _, data1 := range data.RecommendationDetails {
			fmt.Println(".................................................")
			fmt.Println(*data1.AccountId)
			fmt.Println(*data1.AverageUtilization)
			fmt.Println(*data1.AverageNormalizedUnitsUsedPerHour)
			fmt.Println(*data1.AverageNumberOfInstancesUsedPerHour)
			fmt.Println(*data1.CurrencyCode)
			fmt.Println(*data1.EstimatedMonthlyOnDemandCost)
			fmt.Println(*data1.EstimatedMonthlySavingsAmount)
			fmt.Println(*data1.EstimatedMonthlySavingsPercentage)
			fmt.Println("current")
			fmt.Println(data1.InstanceDetails.EC2InstanceDetails.CurrentGeneration)
			fmt.Println(*data1.InstanceDetails.EC2InstanceDetails.InstanceType)
			fmt.Println(*data1.InstanceDetails.EC2InstanceDetails.Family)
			fmt.Println("recommended")
			fmt.Println(*data1.RecommendedNormalizedUnitsToPurchase)
			fmt.Println(*data1.RecommendedNumberOfInstancesToPurchase)
			fmt.Println(*data1.RecurringStandardMonthlyCost)
		}

	}

}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	GetRecommendRI(context.TODO())
}
