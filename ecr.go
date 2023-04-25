package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// function to get an array of repository names
func getRepositoryNames(sess *session.Session) []string {
	svc := ecr.New(sess)

	// Get list of all repository names
	var repositoryNames []string
	err := svc.DescribeRepositoriesPages(&ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}, func(page *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		for _, repository := range page.Repositories {
			repositoryNames = append(repositoryNames, *repository.RepositoryName)
		}
		return !lastPage
	})
	if err != nil {
		fmt.Println("Failed to describe repositories:", err)
		return nil
	}
	return repositoryNames
}

//function that takes an array of repository names and checks if they have public permissions
func checkRepositoryPermissions(repositoryNames []string, sess *session.Session) bool {
	svc := ecr.New(sess)
	for _, repositoryName := range repositoryNames {
		policyInput := &ecr.GetRepositoryPolicyInput{
			RepositoryName: aws.String(repositoryName),
		}
		policyOutput, err := svc.GetRepositoryPolicy(policyInput)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "RepositoryPolicyNotFoundException" {
				fmt.Printf("\n Repository policy does not exist for %s, skipping\n", repositoryName)
				continue
			} else {
				fmt.Println("Error getting repository policy:", err)
				return false
			}
		}
		// Check if the policy allows public access
		if *policyOutput.PolicyText == "{\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":[\"ecr:GetDownloadUrlForLayer\",\"ecr:BatchGetImage\",\"ecr:BatchCheckLayerAvailability\"],\"Resource\":\"*\"}]}" {
			fmt.Println("The repository", repositoryName, "has public access enabled. ❌")
			return true
		}
	}
	return false
}
