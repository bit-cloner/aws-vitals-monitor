package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
)

func performServiceQuotaChecks(svc *servicequotas.ServiceQuotas) {
	// List service quotas for all services
	listServicesInput := &servicequotas.ListServicesInput{}
	for {
		listServicesOutput, err := svc.ListServices(listServicesInput)
		if err != nil {
			log.Fatalf("Failed to list services: %v", err)
		}

		for _, service := range listServicesOutput.Services {
			serviceCode := *service.ServiceCode

			// List service quotas for the current service
			listQuotasInput := &servicequotas.ListServiceQuotasInput{
				ServiceCode: aws.String(serviceCode),
			}
			listQuotasOutput, err := svc.ListServiceQuotas(listQuotasInput)
			if err != nil {
				log.Fatalf("Failed to list quotas for service '%s': %v", serviceCode, err)
			}

			// Display service quotas as a percentage compared to default values
			for _, quota := range listQuotasOutput.Quotas {
				if !*quota.Adjustable {
					percentage := (*quota.Value / *quota.Value) * 100
					fmt.Printf("%s - %s: %.2f%%\n", serviceCode, *quota.QuotaName, percentage)
				} else {
					fmt.Printf("%s - %s: Adjustable quota, default value unknown\n", serviceCode, *quota.QuotaName)
				}
			}
		}

		if listServicesOutput.NextToken == nil {
			break
		}
		listServicesInput.NextToken = listServicesOutput.NextToken
	}
}
