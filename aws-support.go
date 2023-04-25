package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/support"
)

type CheckInfo struct {
	CheckId   string
	CheckName string
}

func getTrustedAdvisorCheckIds() ([]CheckInfo, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		return nil, err
	}

	svc := support.New(sess)
	input := &support.DescribeTrustedAdvisorChecksInput{
		Language: aws.String("en"),
	}

	result, err := svc.DescribeTrustedAdvisorChecks(input)

	if err != nil {
		if strings.Contains(err.Error(), "SubscriptionRequiredException") {
			fmt.Println("Please run this check from an account with the correct support plan")
		} else {
			return nil, err
		}
	}
	checkInfos := []CheckInfo{}
	for _, check := range result.Checks {
		checkInfo := CheckInfo{
			CheckId:   aws.StringValue(check.Id),
			CheckName: aws.StringValue(check.Name),
		}
		checkInfos = append(checkInfos, checkInfo)
	}

	//fmt.Printf("checkIds: %v", checkIds)
	return checkInfos, nil
}

func getCheckResults(checkInfos []CheckInfo) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		return err
	}

	svc := support.New(sess)

	for _, checkId := range checkInfos {
		input := &support.DescribeTrustedAdvisorCheckResultInput{
			CheckId:  aws.String(checkId.CheckId),
			Language: aws.String("en"),
		}

		result, err := svc.DescribeTrustedAdvisorCheckResult(input)
		if err != nil {
			return err
		}
		if *result.Result.Status != "ok" && *result.Result.Status != "not_available" {
			fmt.Println("\n-------------------------")
			fmt.Printf("Check Name: %s\n", checkId.CheckName)
			fmt.Printf("Status: %s\n", *result.Result.Status)
			fmt.Printf("Resources Summary: %+v\n", result.Result.ResourcesSummary)
			fmt.Printf("flgged resources are below\n")
			for _, resource := range result.Result.FlaggedResources {
				fmt.Printf("\tResource ID: %s\n", *resource.ResourceId)
				for _, metadata := range resource.Metadata {
					if metadata != nil {
						fmt.Printf("\t\tMetadata: %s\n", *metadata)
					}

				}
			}
			fmt.Println("\n-------------------------")
		}

	}

	return nil
}
