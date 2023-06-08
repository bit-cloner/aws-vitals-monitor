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
		publiclyAccessible := false
		if instance.PubliclyAccessible != nil {
			publiclyAccessible = *instance.PubliclyAccessible
		}

		storageEncrypted := false
		if instance.StorageEncrypted != nil {
			storageEncrypted = *instance.StorageEncrypted
		}

		iops := int64(0)
		if instance.Iops != nil {
			iops = *instance.Iops
		}

		multiAZ := false
		if instance.MultiAZ != nil {
			multiAZ = *instance.MultiAZ
		}

		if instance.BackupRetentionPeriod != nil && *instance.BackupRetentionPeriod > 0 {
			fmt.Printf("Automated backups are enabled for DB instance: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("Backup Retention period is: %d Days\n", instance.BackupRetentionPeriod)
		} else {
			fmt.Printf("Automated backups are NOT enabled for DB instance: %s ❓\n", *instance.DBInstanceIdentifier)
		}

		if publiclyAccessible {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  Publicly Accessible: %t❌\n", publiclyAccessible)
		}
		if !storageEncrypted {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  Encryption Enabled: %t❌\n", storageEncrypted)
		}
		if iops > 0 {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  Provisioned IOPS: %d❓\n", iops)
		}
		if !multiAZ {
			fmt.Printf("Instance ID: %s\n", *instance.DBInstanceIdentifier)
			fmt.Printf("  MultiAZ: %t❌\n", multiAZ)
		}
	}
}
