package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// a function that accepts region as input and returns a list of security groups
func getSecurityGroups(region string) ([]*ec2.SecurityGroup, error) {
	// Create a new session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return nil, err
	}
	// Create a new EC2 client
	svc := ec2.New(sess)

	// Iterate over each page of results, adding the security groups to the result slice
	var groups []*ec2.SecurityGroup
	// Use the DescribeSecurityGroupsPages method to paginate results
	err = svc.DescribeSecurityGroupsPages(&ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int64(100),
	}, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, group := range page.SecurityGroups {
			groups = append(groups, group)
		}
		return true
	})

	// Check for any errors encountered during pagination
	if err != nil {
		// Check if there are no security groups found in the region
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "InvalidGroup.NotFound" {
				return nil, fmt.Errorf("no security groups found in region %s", region)
			}
		}
		return nil, err
	}

	return groups, nil
}

// a function that accepts groups and region as input and checks if any of the security group rules have a range of ports defined
func checkSecurityGroupHasPortRange(groups []*ec2.SecurityGroup) bool {
	var found bool
	// loop over security groups
	for _, group := range groups {
		// loop over security group rules

		for _, ipPermission := range group.IpPermissions {
			if ipPermission.FromPort != nil && ipPermission.ToPort != nil && *ipPermission.FromPort != *ipPermission.ToPort {
				fmt.Println("\n\nA security group has a range of ports defined. Please investigate security group: ❌", *group.GroupId)
				fmt.Println("The range of ports is: ", *ipPermission.FromPort, " to ", *ipPermission.ToPort)
				found = true
			}
		}

	}
	if found {
		return true
	}
	return false
}

// Function to check SG has broad private CIDR range as source
func CheckSecurityGroupHasBroadPrivateCidrRange(groups []*ec2.SecurityGroup) bool {
	var found bool
	// loop over security groups
	for _, group := range groups {
		// loop over security group rules
		for _, ipPermission := range group.IpPermissions {
			for _, ipRange := range ipPermission.IpRanges {
				if *ipRange.CidrIp == "10.0.0.0/8" || *ipRange.CidrIp == "172.16.0.0/12" || *ipRange.CidrIp == "192.168.0.0/16" {
					fmt.Println("\nA security group has a broad private CIDR range as source. Please investigate security group: ❌", *group.GroupId)
					fmt.Println("The CIDR range is: ", *ipRange.CidrIp)
					found = true
				}
			}
		}
	}
	if found {
		return true
	}
	return false
}

// Function to get Ec2 instances that are using default security group

func GetDefaultSecurityGroupInstances(region string) []*ec2.Instance {
	// Initialize a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region), // Replace with the AWS region you are using
	})

	if err != nil {
		panic(err)
	}

	// Create a new EC2 client
	ec2Svc := ec2.New(sess)

	// Initialize input parameters for DescribeInstances API call
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance.group-name"),
				Values: []*string{
					aws.String("default"),
				},
			},
		},
	}

	// Initialize a variable to hold the paginated output
	var result []*ec2.Instance

	// Paginate through the DescribeInstances results
	err = ec2Svc.DescribeInstancesPages(input,
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			// Append each instance to the result variable
			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
					result = append(result, instance)
				}
			}
			// Return true to continue pagination
			return !lastPage
		})

	if err != nil {
		panic(err)
	}

	// Return the list of instances
	return result
}

// function to check if any SG has inbound rules that are open to all IPs
func CheckSecurityGroupHasOpenInboundRules(groups []*ec2.SecurityGroup) bool {
	var found bool
	// loop over security groups
	for _, group := range groups {
		// loop over security group rules
		for _, ipPermission := range group.IpPermissions {
			for _, ipRange := range ipPermission.IpRanges {
				// Check for nil values before dereferencing pointers
				if ipRange.CidrIp != nil && ipPermission.FromPort != nil && group.GroupId != nil {
					if *ipRange.CidrIp == "0.0.0.0/0" || *ipRange.CidrIp == "::/0" {
						fmt.Println("\nA security group has an excessively open inbound rule on port", *ipPermission.FromPort, ". Please investigate security group: ❌", *group.GroupId)
						found = true
					}
				}
			}
		}
	}
	return found
}
