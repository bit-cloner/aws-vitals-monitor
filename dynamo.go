package main

import (
	"fmt"
	"log"
	"math"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func printdynamoTableStats(svc *dynamodb.DynamoDB) error {

	// Get the list of table names
	input := &dynamodb.ListTablesInput{}
	tables := make([]string, 0)

	for len(tables) < 500 {
		result, err := svc.ListTables(input)
		if err != nil {
			log.Fatalf("Failed to list DynamoDB tables: %v", err)
		}

		tables = append(tables, aws.StringValueSlice(result.TableNames)...)

		if result.LastEvaluatedTableName == nil {
			break
		}
		input.ExclusiveStartTableName = result.LastEvaluatedTableName
	}

	// Set counters
	counter := 0
	provisionedTables := 0
	ondemandTables := 0

	for _, tableName := range tables {

		// Get table description
		descParams := &dynamodb.DescribeTableInput{TableName: aws.String(tableName)}
		tableDescription, err := svc.DescribeTable(descParams)
		if err != nil {
			return err
		}
		if tableDescription.Table.ProvisionedThroughput != nil &&
			tableDescription.Table.ProvisionedThroughput.WriteCapacityUnits != nil {
			// Get billing mode summary

			if *tableDescription.Table.ProvisionedThroughput.WriteCapacityUnits != 0 {

				provisionedTables++

			} else {
				ondemandTables++
			}

		}

		// Increment counter
		counter++
	}

	// Calculate percentages
	totalTables := float64(counter)
	provisionedPercentage := math.Round(float64(provisionedTables) * 100 / totalTables)
	ondemandPercentage := math.Round(float64(ondemandTables) * 100 / totalTables)

	fmt.Printf("Total tables: %d\n", counter)
	fmt.Printf("Provisioned tables: %d (%.1f%%)\n", provisionedTables, provisionedPercentage)
	fmt.Printf("On-Demand tables: %d (%.1f%%)\n", ondemandTables, ondemandPercentage)

	return nil
}
