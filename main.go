package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
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
	shouldRunTrustedAdvisorChecks := false
	trustedAdvisorPrompt := &survey.Confirm{
		Message: "Do you want to run Trusted Advisor checks?",
	}

	shouldRunEC2Checks := false
	ec2prompt := &survey.Confirm{
		Message: "Do you want to run EC2 checks?",
	}
	cpuThresholdPrompt := &survey.Select{
		Message: "Enter a CPU threshold (default 20%):",
		Options: []string{"20", "30", "40", "50"},
		Default: "20",
	}

	timeframePrompt := &survey.Select{
		Message: "Enter a timeframe (default 3 days):",
		Options: []string{"1", "3", "7", "14"},
		Default: "3",
	}

	shouldRunDynamoDBChecks := false
	dynamoDBPrompt := &survey.Confirm{
		Message: "Do you want to run DynamoDB checks?",
	}

	// Ask the user if they want to run s3 checks
	shouldRunS3Checks := false
	s3prompt := &survey.Confirm{
		Message: "Do you want to run S3 checks?",
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
	dynamoDBSvc := dynamodb.New(sess)
	// Call printAccountInfo function
	accountInfo := printAccountInfo(iamSvc, stsSvc, selectedRegion)

	// Start of snapshot check. This can be made as a go routine

	// Create a list of snapshot ids
	snapshotIds := getSnapshotIds(selectedRegion)
	var foundPublicSnapshot bool
	counter := 0

	if len(snapshotIds) > 100 {
		fmt.Println("There are more than 100 snapshots in this region. 100 random snapshots will be checked.")

		// Seed the random number generator
		rand.Seed(time.Now().UnixNano())

		// Shuffle the snapshotIds slice
		rand.Shuffle(len(snapshotIds), func(i, j int) {
			snapshotIds[i], snapshotIds[j] = snapshotIds[j], snapshotIds[i]
		})

		// Now just take the first 100
		snapshotIds = snapshotIds[:100]
	}

	for _, snapshotId := range snapshotIds {
		counter++
		fmt.Printf("\r #### Analyzed number of snapshots: %d ####", counter)
		foundPublicSnapshot = checkSnapshot(snapshotId, selectedRegion)
	}

	if !foundPublicSnapshot {
		fmt.Println("\nNo snapshots were found that are publicly shared: ✅")
	}
	// populate Snapshots struct
	snapshotsData := Snapshots{
		TotalAnalyzed:  counter,
		PubliclyShared: foundPublicSnapshot,
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

	// Ask user if they want to run Trusted Advisor checks using survey

	err = survey.AskOne(trustedAdvisorPrompt, &shouldRunTrustedAdvisorChecks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}
	if shouldRunTrustedAdvisorChecks {
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

	}

	// ask user if they want to run EC2 instance checks using survey
	err = survey.AskOne(ec2prompt, &shouldRunEC2Checks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}

	if shouldRunEC2Checks {
		cpuThresholdStr := ""
		err := survey.AskOne(cpuThresholdPrompt, &cpuThresholdStr)
		if err != nil {
			fmt.Println("Error with survey:", err)
			return
		}
		cpuThreshold, err := strconv.Atoi(cpuThresholdStr)
		if err != nil {
			fmt.Println("Error converting CPU threshold to integer:", err)
			return
		}
		//fmt.Println("CPU threshold set to:", cpuThreshold)

		// ask for timeframe
		timeframeStr := ""
		err = survey.AskOne(timeframePrompt, &timeframeStr)
		if err != nil {
			fmt.Println("Error with survey:", err)
			return
		}
		timeframeDays, err := strconv.Atoi(timeframeStr)
		if err != nil {
			fmt.Println("Error converting timeframe to integer:", err)
			return
		}
		timeframe := time.Duration(timeframeDays) * 24 * time.Hour
		//fmt.Println("Timeframe set to:", timeframe)
		performEC2Checks(ec2Svc, cpuThreshold, timeframe)
	}
	// ask user if they want to run s3 checks using survey
	err = survey.AskOne(s3prompt, &shouldRunS3Checks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}
	if shouldRunS3Checks {
		// Get all the buckets in the region
		bucketnames, err := getAllBucketNames(S3svc)
		if err != nil {
			fmt.Println("Failed to get buckets:", err)
		}
		getPercentageStorageclasses(S3svc, bucketnames)
	}

	err = survey.AskOne(dynamoDBPrompt, &shouldRunDynamoDBChecks)
	if err != nil {
		fmt.Println("Error with survey:", err)
		return
	}
	if shouldRunDynamoDBChecks {
		err = printdynamoTableStats(dynamoDBSvc)
		if err != nil {
			fmt.Println("Error with DynamoDB checks:", err)
			return
		}
	}

	// Ask user to display in JSON format
	displayJSON := false
	displayprompt := &survey.Confirm{
		Message: "Do you want to display this information in JSON format?",
	}
	err = survey.AskOne(displayprompt, &displayJSON)
	if err != nil {
		fmt.Println("Error with survey:", err)
	}

	if displayJSON {
		findings := Findings{
			Snapshots: snapshotsData,
		}
		masterStruct := MasterStructure{
			AccountInformation: accountInfo,
			Findings:           findings,
		}

		jsonData, err := json.MarshalIndent(masterStruct, "", "    ")
		if err != nil {
			fmt.Println("Error with parsing json data:", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}

}
