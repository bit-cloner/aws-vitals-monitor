package main

import (
	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/http"

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
	outdatedRuntimes, err := GetDeprecatedRuntimes()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, outdatedRuntime := range outdatedRuntimes {
		if function.Runtime != nil && *function.Runtime == outdatedRuntime {
			fmt.Printf("\nOutdated runtime detected for function %s: %s\n", *function.FunctionName, *function.Runtime)
		}
	}
}

// DeprecatedRuntimesResponse struct to map the JSON response
type DeprecatedRuntimesResponse struct {
	DeprecatedRuntimes []string `json:"deprecated_runtimes"`
}

// GetDeprecatedRuntimes fetches a list of deprecated runtimes from the specified URL
func GetDeprecatedRuntimes() ([]string, error) {
	url := "https://lambda-deprecated-runtimes-atzlvbq4rq-uc.a.run.app"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching deprecated runtimes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var response DeprecatedRuntimesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return response.DeprecatedRuntimes, nil
}
