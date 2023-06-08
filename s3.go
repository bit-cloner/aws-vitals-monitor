package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func getAllBucketNames(svc *s3.S3) ([]string, error) {
	buckets, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %v", err)
	}

	bucketNames := make([]string, len(buckets.Buckets))
	for i, bucket := range buckets.Buckets {
		bucketNames[i] = *bucket.Name
	}
	fmt.Printf("\n Found %d buckets\n", len(bucketNames))
	// If there are more than 100 buckets, randomly select 100 of them
	if len(bucketNames) > 100 {
		fmt.Println("More than 100 buckets found, randomly selecting 100 buckets")
		rand.Seed(time.Now().UnixNano())
		selectedBucketNames := make([]string, 100)
		for i := range selectedBucketNames {
			randomIndex := rand.Intn(len(bucketNames))
			selectedBucketNames[i] = bucketNames[randomIndex]
			bucketNames[randomIndex] = bucketNames[len(bucketNames)-1]
			bucketNames = bucketNames[:len(bucketNames)-1]
		}
		bucketNames = selectedBucketNames
	}

	return bucketNames, nil
}

func getBucketLifecyclePolicy(svc *s3.S3, bucketName string) (bool, error) {
	input := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	}

	_, err := svc.GetBucketLifecycleConfiguration(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NoSuchLifecycleConfiguration" {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func getPercentageStorageclasses(svc *s3.S3, bucketNames []string) {
	bucketsWithoutLifecycle := 0
	skippedBuckets := 0

	var wg sync.WaitGroup
	// Semaphore to limit concurrency
	concurrencyLimit := make(chan struct{}, 20) // Adjust this value based on your rate limit

	for _, bucketName := range bucketNames {
		wg.Add(1)
		go func(bucketName string) {
			defer wg.Done()
			concurrencyLimit <- struct{}{}        // Acquire
			defer func() { <-concurrencyLimit }() // Release

			input := &s3.ListObjectsV2Input{
				Bucket:  aws.String(bucketName),
				MaxKeys: aws.Int64(100),
			}

			result, err := svc.ListObjectsV2(input)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "BucketRegionError" {
					fmt.Printf("Bucket %s is in another region, skipping\n", bucketName)
					skippedBuckets++
					return
				} else {
					fmt.Printf("Failed to list objects in bucket %s: %v\n", bucketName, err)
					return
				}
			}

			// Calculate the storage class counts
			storageClassCounts := make(map[string]int)
			totalObjects := len(result.Contents)
			for _, object := range result.Contents {
				storageClass := aws.StringValue(object.StorageClass)
				storageClassCounts[storageClass]++
			}

			// Create a new tabwriter.Writer
			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			// Print the bucket name and storage class percentages
			fmt.Fprintf(writer, "Bucket: %s\n", bucketName)
			fmt.Fprintf(writer, "%-20s\t|\t%-10s\n", "StorageClass", "Percentage")
			fmt.Fprintln(writer, "--------------------\t|\t-----------")

			for storageClass, count := range storageClassCounts {
				percentage := float64(count) / float64(totalObjects) * 100
				fmt.Fprintf(writer, "%-20s\t|\t%9.2f%%\n", storageClass, percentage)
			}
			// Flush the tabwriter to print the formatted output
			writer.Flush()
			fmt.Println()

			// Call `getBucketLifecyclePolicy` for each bucket
			hasLifecycle, err := getBucketLifecyclePolicy(svc, bucketName)
			if err != nil {
				fmt.Printf("Failed to get lifecycle policy for bucket %s: %v\n", bucketName, err)
				skippedBuckets++
				return
			}
			if !hasLifecycle {
				bucketsWithoutLifecycle++
			}
		}(bucketName)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Calculate and print the percentage of buckets without a lifecycle policy
	// considering only the processed buckets (excluding skipped buckets)
	processedBuckets := len(bucketNames) - skippedBuckets
	percentageWithoutLifecycle := float64(bucketsWithoutLifecycle) / float64(processedBuckets) * 100
	bannerText := fmt.Sprintf("Percentage of buckets without a lifecycle policy: %.2f%%", percentageWithoutLifecycle)
	printBanner(bannerText)
}

func printBanner(text string) {
	border := strings.Repeat("=", len(text)+6)
	fmt.Printf("%s\n", border)
	fmt.Printf("== %s ==\n", text)
	fmt.Printf("%s\n", border)
}
