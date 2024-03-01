package main

import (
	"fmt"
	"os"
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

	checkInfos := make([]CheckInfo, 0)
	for _, check := range result.Checks {
		checkInfos = append(checkInfos, CheckInfo{
			CheckId:   aws.StringValue(check.Id),
			CheckName: aws.StringValue(check.Name),
		})
	}

	return checkInfos, nil
}

func getCheckResults(checkInfos []CheckInfo) error {
	file, err := os.Create("trusted-advisor-findings.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := support.New(sess)

	for _, checkInfo := range checkInfos {
		input := &support.DescribeTrustedAdvisorCheckResultInput{
			CheckId:  aws.String(checkInfo.CheckId),
			Language: aws.String("en"),
		}

		result, err := svc.DescribeTrustedAdvisorCheckResult(input)
		if err != nil {
			return err
		}

		if *result.Result.Status != "ok" && *result.Result.Status != "not_available" {
			fmt.Fprintf(file, "\n-------------------------\n")
			fmt.Fprintf(file, "Check Name: %s\n", checkInfo.CheckName)
			fmt.Fprintf(file, "Status: %s\n", *result.Result.Status)
			fmt.Fprintf(file, "Resources Summary: %+v\n", result.Result.ResourcesSummary)
			fmt.Fprintf(file, "Flagged resources are below:\n")
			for _, resource := range result.Result.FlaggedResources {
				fmt.Fprintf(file, "\tResource ID: %s\n", *resource.ResourceId)
				metadataStrings := make([]string, len(resource.Metadata))
				for i, metadata := range resource.Metadata {
					if metadata != nil {
						metadataStrings[i] = *metadata
					} else {
						metadataStrings[i] = "" // Or some placeholder for nil metadata
					}
				}
				fmt.Fprintf(file, "\t\tMetadata: %s\n", strings.Join(metadataStrings, ", "))
			}
			fmt.Fprintf(file, "\n-------------------------\n")
		}
	}

	return nil
}
