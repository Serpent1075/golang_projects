package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

func main() {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
	)
	check(err)
	client := wafv2.NewFromConfig(cfg)
	input := wafv2.GetManagedRuleSetInput{
		Id:    aws.String(""),
		Name:  aws.String("waf-tf_WAF_ACL"),
		Scope: types.ScopeRegional,
	}

	result, resulterr := client.GetManagedRuleSet(context.Background(), &input)
	check(resulterr)
	fmt.Println(*result.ManagedRuleSet)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
