package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

func listRDSInstances(rdsClient *rds.RDS) ([]*rds.DBInstance, error) {
	input := &rds.DescribeDBInstancesInput{
		MaxRecords: aws.Int64(100),
	}
	result, err := rdsClient.DescribeDBInstances(input)

	if err != nil {
		return nil, err
	}

	return result.DBInstances, nil
}

func checkRDSInstanceAttributes(dbInstances []*rds.DBInstance) {
	fmt.Printf("\n #### Analyzing %d RDS Instances ####\n", len(dbInstances))
	for _, instance := range dbInstances {
		publiclyAccessible := *instance.PubliclyAccessible
		storageEncrypted := *instance.StorageEncrypted

		if publiclyAccessible {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  Publicly Accessible: %t❌\n", publiclyAccessible)
		}
		if !storageEncrypted {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  Encryption Enabled: %t❌\n", storageEncrypted)
		}
	}
}
