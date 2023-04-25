package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

func regionFullName(regionCode string) string {
	regionNames := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"af-south-1":     "Africa (Cape Town)",
		"ap-east-1":      "Asia Pacific (Hong Kong)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-3": "Asia Pacific (Osaka-Local)",
		"ca-central-1":   "Canada (Central)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-west-2":      "Europe (London)",
		"eu-south-1":     "Europe (Milan)",
		"eu-west-3":      "Europe (Paris)",
		"eu-north-1":     "Europe (Stockholm)",
		"me-south-1":     "Middle East (Bahrain)",
		"sa-east-1":      "South America (São Paulo)",
		"us-gov-east-1":  "AWS GovCloud (US-East)",
		"us-gov-west-1":  "AWS GovCloud (US-West)",
		"cn-north-1":     "China (Beijing)",
		"cn-northwest-1": "China (Ningxia)",
	}

	regionName, ok := regionNames[regionCode]
	if !ok {
		regionName = "Unknown"
	}

	return regionName
}

func printAccountInfo(iamSvc *iam.IAM, stsSvc *sts.STS, region string) {
	// Get caller identity
	callerIdentityOutput, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatalf("Failed to get caller identity: %v", err)
	}

	// Get Account ID
	accountID := *callerIdentityOutput.Account

	// List account aliases
	accountAliasesOutput, err := iamSvc.ListAccountAliases(&iam.ListAccountAliasesInput{})
	if err != nil {
		log.Fatalf("Failed to list account aliases: %v", err)
	}

	// Get Account Alias
	accountAlias := "N/A"
	if len(accountAliasesOutput.AccountAliases) > 0 {
		accountAlias = *accountAliasesOutput.AccountAliases[0]
	}

	// Get region full name
	regionFullName := regionFullName(region)

	// Print banner
	banner := strings.Builder{}
	banner.WriteString("╔════════════════════════════════════════════════════════════════════════════════════╗\n")
	banner.WriteString("║\t\t\t\tAccount Information\t\t\t\t\n")
	banner.WriteString("╠════════════════════════════════════════════════════════════════════════════════════╣\n")
	banner.WriteString(fmt.Sprintf("║\tAccount ID:\t%-40s\t\n", accountID))
	banner.WriteString(fmt.Sprintf("║\tAccount Alias:\t%-40s\t\n", accountAlias))
	banner.WriteString(fmt.Sprintf("║\tRegion Code:\t%-40s\t\n", region))
	banner.WriteString(fmt.Sprintf("║\tRegion Name:\t%-40s\t\n", regionFullName))
	banner.WriteString("╚════════════════════════════════════════════════════════════════════════════════════╝\n")

	fmt.Print(banner.String())
}
