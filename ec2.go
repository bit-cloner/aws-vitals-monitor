package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mikioh/ipaddr"
)

type InstanceTypePercentage struct {
	InstanceType string
	Count        int
	Percentage   float64
}

// function to get a list of snapshot ids that are owned by the account
func getSnapshotIds(region string) []string {
	// Create a new session with the AWS SDK
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		fmt.Println("Failed to create session:", err)
		return nil
	}

	svc := ec2.New(sess)

	// Get list of all snapshot ids that are owned by the account
	var snpshotIds []string
	err = svc.DescribeSnapshotsPages(&ec2.DescribeSnapshotsInput{
		OwnerIds:   []*string{aws.String("self")},
		MaxResults: aws.Int64(100),
	}, func(page *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		for _, snapshot := range page.Snapshots {
			snpshotIds = append(snpshotIds, *snapshot.SnapshotId)
		}
		return !lastPage

	})
	if err != nil {
		fmt.Println("Failed to describe snapshots:", err)
		return nil
	}
	return snpshotIds
}

//a function to check if a snapshot has public permission to create volume returns boolean value
func checkSnapshot(snapshotId string, region string) bool {
	// Create a new session with the AWS SDK
	sess, err := session.NewSession(&aws.Config{ // can enhance this by makin it an input to function
		Region: aws.String(region),
	})
	if err != nil {
		fmt.Println("Failed to create session:", err)
		os.Exit(1)
	}
	svc := ec2.New(sess)
	// describe snapshot attribute
	snapshotAttributes, err := svc.DescribeSnapshotAttribute(&ec2.DescribeSnapshotAttributeInput{
		Attribute:  aws.String("createVolumePermission"),
		SnapshotId: aws.String(snapshotId),
	})
	if err != nil {
		fmt.Println("Failed to describe snapshot:", err)
		return false
	}
	// loop over snapshot attribute and print them
	if len(snapshotAttributes.CreateVolumePermissions) > 0 {
		for _, snapshot := range snapshotAttributes.CreateVolumePermissions {
			if *snapshot.Group == "all" {
				fmt.Println("\nA snapshot has public permission to create volume. Please investigate snapshot: ❌", snapshotId)
				return true
			}
		}

	}

	return false
}

// Function that takes session as input and describes addresses and checks for assosciationid is empty
func checkElasticIPs(sess *session.Session) bool {
	svc := ec2.New(sess)
	filters := []*ec2.Filter{
		{
			Name:   aws.String("association-id"),
			Values: []*string{aws.String("")},
		},
	}
	addresses, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: filters,
	})
	if err != nil {
		fmt.Println("Failed to describe addresses:", err)
		return false
	}
	for _, address := range addresses.Addresses {
		if address.AssociationId == nil {
			fmt.Println("An elastic IP is not associated with any instance. Please investigate  and release elastic IP: ❌", *address.PublicIp)
			return true
		}
	}
	return false
}

func fetchSubnets(sess *session.Session) []*ec2.Subnet {
	ec2Client := ec2.New(sess)
	input := &ec2.DescribeSubnetsInput{}
	result, err := ec2Client.DescribeSubnets(input)
	if err != nil {
		panic(err)
	}
	return result.Subnets
}

type SubnetInfo struct {
	Cidr string
	Id   string
}

func extractSubnetInfo(subnets []*ec2.Subnet) []SubnetInfo {
	var subnetInfoList []SubnetInfo
	for _, subnet := range subnets {
		info := SubnetInfo{
			Cidr: *subnet.CidrBlock,
			Id:   *subnet.SubnetId,
		}
		subnetInfoList = append(subnetInfoList, info)
	}
	return subnetInfoList
}
func checkSubnetOverlaps(arr []SubnetInfo) {
	prefixList := make([]*ipaddr.Prefix, len(arr))
	for i, cidr := range arr {
		_, ipNet, err := net.ParseCIDR(cidr.Cidr)
		if err != nil {
			fmt.Printf("Error parsing CIDR: %s, error: %v\n", cidr, err)
			return
		}
		prefixList[i] = ipaddr.NewPrefix(ipNet)
	}
	//fmt.Println("subnets are ", prefixList)
	foundOverlapping := false
	for i := 0; i < len(prefixList); i++ {
		for j := i + 1; j < len(prefixList); j++ {
			if prefixList[i].Overlaps(prefixList[j]) {
				fmt.Printf("Subnets %s and %s overlap ⚠️ \n", arr[i].Id, arr[j].Id)
				foundOverlapping = true
			}
		}
	}
	if !foundOverlapping {
		fmt.Println("\n No overlapping subnets found ✅")
	}
}

func findOrphanedEBSVolumes(svc *ec2.EC2) ([]*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		MaxResults: aws.Int64(200),
	}

	result, err := svc.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}

	orphanedVolumes := make([]*ec2.Volume, 0)

	for _, volume := range result.Volumes {
		if aws.StringValue(volume.State) == "available" && len(volume.Attachments) == 0 {
			orphanedVolumes = append(orphanedVolumes, volume)
		}
	}

	return orphanedVolumes, nil
}

func performEC2Checks(svc *ec2.EC2) {
	input := &ec2.DescribeInstancesInput{}

	onDemandCount := 0
	spotCount := 0
	instanceTypeCounts := make(map[string]int)
	instances := make([]*ec2.Instance, 0)
	runningInstances := make([]*ec2.Instance, 0)
	// Counting On-Demand and Spot instances
	err := svc.DescribeInstancesPages(input,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
					instances = append(instances, instance)
					if *instance.State.Name == "running" {
						runningInstances = append(runningInstances, instance)
					}
					if instance.InstanceLifecycle != nil && *instance.InstanceLifecycle == "spot" {
						spotCount++
					} else {
						onDemandCount++
					}
					instanceTypeCounts[*instance.InstanceType]++

				}
			}
			return !lastPage
		})

	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}
	// check for imdv1 instances
	checkForIMDv1Instances(instances)
	// check for underutilized instances
	cpuThreshold := 20.0
	timeframe := 72 * time.Hour
	displayUnderutilizedInstances(runningInstances, cpuThreshold, timeframe)
	// Counting Reserved instances
	reservedInput := &ec2.DescribeReservedInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("active")},
			},
		},
	}

	reservedInstances, err := svc.DescribeReservedInstances(reservedInput)
	if err != nil {
		fmt.Println("Error describing reserved instances:", err)
		return
	}

	fmt.Printf("\n Found %d purchases of reserved instances\n", len(reservedInstances.ReservedInstances))
	for _, reservedInstance := range reservedInstances.ReservedInstances {
		years := float64(*reservedInstance.Duration) / (60 * 60 * 24 * 365)
		fmt.Printf("Instance Type: %s, Availability Zone: %s, Instance Count: %d, Duration: %.2f years\n",
			*reservedInstance.InstanceType,
			*reservedInstance.AvailabilityZone,
			*reservedInstance.InstanceCount,
			years,
		)
	}
	// Calculate On-Demand and Spot instances percentages
	onDemandPercentage := float64(onDemandCount) / float64(onDemandCount+spotCount) * 100
	spotPercentage := float64(spotCount) / float64(onDemandCount+spotCount) * 100

	// Print On-Demand and Spot instances percentages
	fmt.Printf("\nOn-Demand Instances: %d (%.2f%%)\n", onDemandCount, onDemandPercentage)
	fmt.Printf("Spot Instances: %d (%.2f%%)\n", spotCount, spotPercentage)

	totalInstances := onDemandCount + spotCount
	instancePercentages := make([]InstanceTypePercentage, 0, len(instanceTypeCounts))

	for instanceType, count := range instanceTypeCounts {
		percentage := float64(count) / float64(totalInstances) * 100
		instancePercentages = append(instancePercentages, InstanceTypePercentage{
			InstanceType: instanceType,
			Count:        count,
			Percentage:   percentage,
		})
	}
	sort.Slice(instancePercentages, func(i, j int) bool {
		return instancePercentages[i].Percentage > instancePercentages[j].Percentage
	})
	fmt.Println("\n Instance type distribution")
	fmt.Println("=========================================")
	// Define the maximum bar width
	const maxBarWidth = 15

	for _, instancePercentage := range instancePercentages {
		// Calculate the bar width based on the percentage
		barWidth := int(math.Round(float64(maxBarWidth) * instancePercentage.Percentage / 100))

		// Create the bar using the Unicode character █
		bar := strings.Repeat("█", barWidth)

		// Print the result
		fmt.Printf("Instace Type: %-20s: %s (%7.1f%%)\n\n", instancePercentage.InstanceType, bar, instancePercentage.Percentage)
	}
	fmt.Println("=========================================")
}

func checkForIMDv1Instances(instances []*ec2.Instance) {
	fmt.Printf("\nChecking %d instances for (IMDv1)\n", len(instances))
	imdv1Instances := make([]*ec2.Instance, 0)
	for _, instance := range instances {
		imdv1Instance := false
		if instance.MetadataOptions != nil {
			httpEndpointEnabled := instance.MetadataOptions.HttpEndpoint != nil && *instance.MetadataOptions.HttpEndpoint == "enabled"
			httpTokensOptional := instance.MetadataOptions.HttpTokens != nil && *instance.MetadataOptions.HttpTokens == "optional"

			imdv1Instance = httpEndpointEnabled && httpTokensOptional
		}
		if imdv1Instance {
			imdv1Instances = append(imdv1Instances, instance)
		}
	}

	if len(imdv1Instances) > 0 {
		fmt.Printf("Found %d instances using instance metadata version 1 (IMDv1):\n", len(imdv1Instances))
		for _, instance := range imdv1Instances {
			fmt.Printf("- Instance ID: %s\n", *instance.InstanceId)
		}
	} else {
		fmt.Println("No instances found using instance metadata version 1 (IMDv1).")
	}
}

func getInstanceAverageCPU(instance *ec2.Instance, timeframe time.Duration) (float64, error) {
	cwSess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := cloudwatch.New(cwSess)
	endTime := time.Now()
	startTime := endTime.Add(-timeframe)

	// Check if the instance was started within the specified time frame
	if instance.LaunchTime.After(startTime) {
		return 0, fmt.Errorf("instance %s was started within the specified time period", *instance.InstanceId)
	}

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/EC2"),
		MetricName: aws.String("CPUUtilization"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: instance.InstanceId,
			},
		},
		StartTime: &startTime,
		EndTime:   &endTime,
		Period:    aws.Int64(3600),
		Statistics: []*string{
			aws.String(cloudwatch.StatisticAverage),
		},
	}

	output, err := svc.GetMetricStatistics(input)
	if err != nil {
		return 0, err
	}

	total := 0.0
	count := 0.0
	for _, datapoint := range output.Datapoints {
		total += *datapoint.Average
		count++
	}

	if count == 0 {
		return 0, fmt.Errorf("no datapoints found for instance %s", *instance.InstanceId)
	}

	average := total / count
	return average, nil
}

func displayUnderutilizedInstances(runningInstances []*ec2.Instance, cpuThreshold float64, timeframe time.Duration) {
	results := make(chan *ec2.Instance, len(runningInstances))
	cpuUsages := make(chan float64, len(runningInstances))
	concurrencyLimit := 10
	sem := make(chan bool, concurrencyLimit)

	for _, instance := range runningInstances {
		sem <- true
		go func(instance *ec2.Instance) {
			defer func() { <-sem }()
			averageCPU, err := getInstanceAverageCPU(instance, timeframe)
			if err != nil {
				if strings.Contains(err.Error(), "instance was started within the specified time period") {
					fmt.Printf("Skipping instance %s: %s\n", *instance.InstanceId, err)
				} else {
					fmt.Printf("Error getting average CPU for instance %s: %s\n", *instance.InstanceId, err)
				}
				return
			}

			if averageCPU < cpuThreshold {
				results <- instance
				cpuUsages <- averageCPU
			}
		}(instance)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	close(results)
	close(cpuUsages)

	for instance := range results {
		averageCPU := <-cpuUsages
		fmt.Printf("Instance %s is underutilized (Average CPU usage: %.2f%%, Threshold: %.2f%%)\n", *instance.InstanceId, averageCPU, cpuThreshold)
	}
}
