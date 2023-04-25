package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/sts"
)

func main() {
	// Create a list of regions
	regionNames := []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"af-south-1",
		"ap-east-1",
		"ap-south-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-south-1",
		"eu-west-3",
		"eu-north-1",
		"me-south-1",
		"sa-east-1",
	}

	// Use the alec survey module to ask the user to select a region
	var selectedRegion string
	prompt := &survey.Select{
		Message: "Select a region:",
		Options: regionNames,
	}
	err := survey.AskOne(prompt, &selectedRegion)
	if err != nil {
		fmt.Println("Failed to get user input:", err)
		return
	}
	fmt.Println("Selected region:", selectedRegion)

	shouldRunEC2Checks := false
	ec2prompt := &survey.Confirm{
		Message: "Do you want to run EC2 checks?",
	}

	shouldRunServiceQuotaChecks := false
	serviceQuotaPrompt := &survey.Confirm{
		Message: "Do you want to run Service Quota checks?",
	}

	// Create a new session with the AWS SDK
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(selectedRegion),
	})
	if err != nil {
		fmt.Println("Failed to create session:", err)
		os.Exit(1)
	}
	S3svc := s3.New(sess)
	ec2Svc := ec2.New(sess)
	lambdaClient := lambda.New(sess)
	rdsClient := rds.New(sess)
	iamSvc := iam.New(sess)
	stsSvc := sts.New(sess)
	serviceQuotasSvc := servicequotas.New(sess)
	// Call printAccountInfo function
	printAccountInfo(iamSvc, stsSvc, selectedRegion)

	// Start of snapshot check. This can be made as a go routine

	// Create a list of snapshot ids
	snapshotIds := getSnapshotIds(selectedRegion)
	var foundPublicSnapshot bool
	counter := 0

	// loop over snapshot ids--- Checking only 100 snapshots as processing 1000s of snapshots array is creating a panic error-- to be fixed later
	for i := 0; i < len(snapshotIds) && i < 100; i++ {
		counter++
		fmt.Printf("\r #### Analyzed number of snapshots: %d ####", counter)
		foundPublicSnapshot = checkSnapshot(snapshotIds[i], selectedRegion)
	}
	if !foundPublicSnapshot {
		fmt.Println("\nNo snapshots were found that are publicly shared: ✅")
	}

	groups, err := getSecurityGroups(selectedRegion)
	if err != nil {
		fmt.Printf("Error describing security groups in %s: %v\n", selectedRegion, err)
		return
	}

	// Check security groups for range of open ports in inbound rules

	fmt.Printf("\n #### Analyzing %d Security groups ####", len(groups))
	foundSGWithOpenPorts := checkSecurityGroupHasPortRange(groups)
	if !foundSGWithOpenPorts {
		fmt.Println("\nNo security groups were found with open ports: ✅")
	}

	// get ec2 instances with default security groups
	defaultSecurityGroupInstances := GetDefaultSecurityGroupInstances(selectedRegion)
	if len(defaultSecurityGroupInstances) > 0 {
		fmt.Println("The following instances are using the default security group: ❌")
		for _, instance := range defaultSecurityGroupInstances {
			fmt.Println("Instance ID: ", *instance.InstanceId)
		}
	} else {
		fmt.Println("\nNo instances are using the default security group: ✅")
	}

	// Check security groups for broad private CIDR range as source
	foundSGWithBroadPrivateCidrRange := CheckSecurityGroupHasBroadPrivateCidrRange(groups)
	if !foundSGWithBroadPrivateCidrRange {
		fmt.Println("\nNo security groups were found with a broad private CIDR range as source: ✅")
	}
	// Check if a SG has a rule that is open to all

	foundSGWithOpenPort := CheckSecurityGroupHasOpenInboundRules(groups)
	if !foundSGWithOpenPort {
		fmt.Println("\nNo security groups were found with open to all sources on standard ports: ✅")
	}
	// Get a list of repository names
	repositories := getRepositoryNames(sess)
	if len(repositories) > 0 {
		foundPublicRepository := checkRepositoryPermissions(repositories, sess)
		if !foundPublicRepository {
			fmt.Println("\nNo repositories were found that are publicly shared: ✅")
		}
	} else {
		fmt.Println("\nNo repositories were found in the selected region: ✅")
	}
	foundOrphanedElasticIPs := checkElasticIPs(sess)
	if !foundOrphanedElasticIPs {
		fmt.Println("\nNo orphaned elastic IPs were found: ✅")
	}

	// Fetch list of subnets
	subnets := fetchSubnets(sess)
	// Extract subnet CIDRs and IDs
	subnetInfoList := extractSubnetInfo(subnets)
	checkSubnetOverlaps(subnetInfoList)

	// Get all the buckets in the region
	bucketnames, err := getAllBucketNames(S3svc)
	if err != nil {
		fmt.Println("Failed to get buckets:", err)
	}
	getPercentageStorageclasses(S3svc, bucketnames)
	//check for orphaned volumes
	orphanedVolumes, err := findOrphanedEBSVolumes(ec2Svc)
	if err != nil {
		fmt.Println("Failed to get orphaned volumes:", err)
	}
	if len(orphanedVolumes) > 0 {
		const costPerGBPerMonth float64 = 0.10 // Change this to the appropriate cost per GB per month
		// Sort the orphaned volumes by their size
		sort.Slice(orphanedVolumes, func(i, j int) bool {
			return aws.Int64Value(orphanedVolumes[i].Size) > aws.Int64Value(orphanedVolumes[j].Size)
		})
		fmt.Println("The following volumes are orphaned: ❌")
		var totalOrphanedVolumeSize int64
		var totalOrphanedVolumeCost float64
		for _, volume := range orphanedVolumes {
			volumeID := aws.StringValue(volume.VolumeId)
			volumeSize := aws.Int64Value(volume.Size)
			approxMonthlyCost := float64(volumeSize) * costPerGBPerMonth
			fmt.Printf("Volume ID: %s, Size: %d GB, Approximate monthly cost in USD: $%.2f\n", volumeID, volumeSize, approxMonthlyCost)
			totalOrphanedVolumeSize += volumeSize
			totalOrphanedVolumeCost += approxMonthlyCost
		}
		fmt.Printf("\n ####Total orphaned volume size: %d GB #### \n", totalOrphanedVolumeSize)
		fmt.Printf("\n ####Total approximate monthly cost of orphaned volumes: $%.2f ####\n", totalOrphanedVolumeCost)
	} else {
		fmt.Println("\nNo EBS volumes are orphaned: ✅")
	}

	//get lambda functions
	lambdaFunctions, err := listLambdaFunctions(lambdaClient)
	if err != nil {
		fmt.Println("Failed to get lambda functions:", err)
	}
	fmt.Printf("\n #### Analyzing %d Lambda functions for outdated runtimes ####", len(lambdaFunctions))
	for _, lambdaFunction := range lambdaFunctions {
		outdatedFunctionRuntimeCheck(lambdaClient, *lambdaFunction.FunctionArn)
	}

	// RDS checks
	rdsInstances, err := listRDSInstances(rdsClient)
	if err != nil {
		fmt.Println("Failed to get RDS instances:", err)
	}
	checkRDSInstanceAttributes(rdsInstances)

	// get Trusted Advisor checks-- available only on higher support plans. region must be us-east-1
	fmt.Println("Getting Trusted Advisor checks... (this may take a few minutes) \n")
	checkInfos, err := getTrustedAdvisorCheckIds()
	if err != nil {
		fmt.Println("Failed to get Trusted Advisor check IDs:", err)
	}
	fmt.Println("Printing Trusted Advisor check results... \n")
	getCheckResults(checkInfos)
	if err != nil {
		fmt.Println("Failed to get Trusted Advisor check results:", err)
	}

	// ask user if they want to run EC2 instance checks using survey
	err = survey.AskOne(ec2prompt, &shouldRunEC2Checks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}

	if shouldRunEC2Checks {
		performEC2Checks(ec2Svc)
	} else {
		os.Exit(0)
	}
	err = survey.AskOne(serviceQuotaPrompt, &shouldRunServiceQuotaChecks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}

	if shouldRunServiceQuotaChecks {
		performServiceQuotaChecks(serviceQuotasSvc)
	} else {
		os.Exit(0)
	}

}
