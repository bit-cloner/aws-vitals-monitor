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
	fmt.Printf("\n#### Analyzing %d RDS Instances ####\n", len(dbInstances))
	for _, instance := range dbInstances {
		fmt.Printf("\n--- Instance ID: %s ---\n", *instance.DBInstanceIdentifier)
		hasNegativeFindings := false

		// Publicly Accessible
		if instance.PubliclyAccessible != nil && *instance.PubliclyAccessible {
			fmt.Printf("  ❌ Publicly Accessible\n")
			hasNegativeFindings = true
		}

		// Storage Encryption
		if instance.StorageEncrypted != nil && !*instance.StorageEncrypted {
			fmt.Printf("  ❌ Encryption Not Enabled\n")
			hasNegativeFindings = true
		}

		// Disk Type
		if instance.StorageType != nil && *instance.StorageType == "gp2" {
			fmt.Printf("  ⚠️ Using gp2 disk type (Consider upgrading to GP3)\n")
			hasNegativeFindings = true
		}

		// MultiAZ
		if instance.MultiAZ != nil && *instance.MultiAZ {
			fmt.Printf("  ⚠️  MultiAZ Enabled\n")
		} else {
			fmt.Printf("  ⚠️ MultiAZ Not Enabled\n")
			hasNegativeFindings = true
		}

		// Backup Retention
		if instance.BackupRetentionPeriod != nil && *instance.BackupRetentionPeriod > 0 {
			fmt.Printf(" ❓ Backup Retention: %d Days\n", *instance.BackupRetentionPeriod)
		} else {
			fmt.Printf("  ❌ Backup Retention: Not Enabled\n")
			hasNegativeFindings = true
		}

		if !hasNegativeFindings {
			fmt.Printf("  ✅ No negative findings\n")
		}

		fmt.Println("-------------------------------") // Separator for the next instance
	}
}
