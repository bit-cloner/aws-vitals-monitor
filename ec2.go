package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

	// Counting On-Demand and Spot instances
	err := svc.DescribeInstancesPages(input,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
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
