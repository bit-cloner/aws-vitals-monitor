package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func listLambdaFunctions(lambdaClient *lambda.Lambda) ([]*lambda.FunctionConfiguration, error) {
	var functions []*lambda.FunctionConfiguration
	input := &lambda.ListFunctionsInput{}

	for {
		result, err := lambdaClient.ListFunctions(input)
		if err != nil {
			return nil, err
		}

		functions = append(functions, result.Functions...)

		// If there are no more functions, break the loop
		if result.NextMarker == nil {
			break
		}

		// Set the Marker for the next iteration
		input.Marker = result.NextMarker
	}

	return functions, nil
}

func outdatedFunctionRuntimeCheck(lambdaClient *lambda.Lambda, functionArn string) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionArn),
	}

	result, err := lambdaClient.GetFunction(input)

	if err != nil {
		fmt.Printf("Error getting function configuration for %s: %s\n", functionArn, err)
		return
	}

	function := result.Configuration

	// Check for outdated runtime environments
	outdatedRuntimes := []string{
		"python3.6",
		"python2.7",
		"dotnetcore2.1",
		"ruby2.5",
		"nodejs10.x",
		"nodejs8.10",
		"nodejs4.3",
		"nodejs6.10",
		"dotnetcore1.0",
		"dotnetcore2.0",
		"nodejs4.3-edge",
		"nodejs",
	}

	for _, outdatedRuntime := range outdatedRuntimes {
		if *function.Runtime == outdatedRuntime {
			fmt.Printf("\nOutdated runtime detected for function %s: %s\n", *function.FunctionName, *function.Runtime)
		}
	}
}
