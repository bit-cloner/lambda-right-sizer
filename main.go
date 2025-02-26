package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"lrs/awsclient"
	"lrs/pricing"
	"lrs/prompts"
	"lrs/visualization"
)

type TestResult struct {
	Memory     int
	DurationMs float64
	Cost       float64
	Logs       string
}

func main() {
	ctx := context.Background()

	// Step 1: Ask for the Lambda ARN using prompt-ui
	lambdaARN, err := prompts.PromptForLambdaARN()
	if err != nil {
		log.Fatalf("Error reading Lambda ARN: %v", err)
	}

	// Extract region from ARN (ARN format: arn:aws:lambda:{region}:{account}:function:{functionName})
	arnParts := strings.Split(lambdaARN, ":")
	if len(arnParts) < 4 {
		log.Fatalf("Invalid Lambda ARN format")
	}
	region := arnParts[3]
	fmt.Printf("Using region: %s\n", region)

	// Create an AWS Lambda client for the given region
	lambdaClient, err := awsclient.NewLambdaClient(ctx, region)
	if err != nil {
		log.Fatalf("Error creating Lambda client: %v", err)
	}

	// Get the Lambda function metadata to determine architecture & original memory configuration
	meta, err := lambdaClient.GetFunctionMetadata(ctx, lambdaARN)
	if err != nil {
		log.Fatalf("Error getting function metadata: %v", err)
	}
	originalMemory := meta.MemorySize
	arch := "x86_64"
	for _, a := range meta.Architectures {
		if a == "arm64" {
			arch = "arm64"
			break
		}
	}
	fmt.Printf("Detected architecture: %s\n", arch)

	// Ask the user if they want to use custom JSON test event data
	useCustomData, err := prompts.PromptYesNo("Do you want to use custom JSON test event data?")
	if err != nil {
		log.Fatalf("Error during prompt: %v", err)
	}

	var testEvent []byte
	if useCustomData {
		// Loop until valid JSON is provided or user opts out
		for {
			jsonStr, err := prompts.PromptForJSON("Please paste your JSON-formatted test event data:")
			if err != nil {
				log.Fatalf("Error reading JSON input: %v", err)
			}
			if json.Valid([]byte(jsonStr)) {
				testEvent = []byte(jsonStr)
				break
			} else {
				choice, err := prompts.PromptYesNo("Invalid JSON. Do you want to re-enter the data?")
				if err != nil {
					log.Fatalf("Error during prompt: %v", err)
				}
				if !choice {
					fmt.Println("Proceeding without custom test event data.")
					testEvent = nil
					break
				}
			}
		}
	}

	// Load pricing data for the given architecture.
	priceTable := pricing.GetPricingTable(arch)

	// Get a sorted list of memory configurations (keys) to iterate in order
	var memorySizes []int
	for mem := range priceTable {
		memorySizes = append(memorySizes, mem)
	}
	sort.Ints(memorySizes)

	var results []TestResult

	// Loop over each memory configuration
	for _, mem := range memorySizes {
		fmt.Printf("\nTesting configuration: %d MB\n", mem)

		// Update the Lambda configuration with the new memory size.
		err = lambdaClient.UpdateFunctionMemory(ctx, lambdaARN, mem)
		if err != nil {
			log.Printf("Error updating function memory to %d MB: %v", mem, err)
			continue
		}

		// Wait for the configuration update to propagate.
		fmt.Println("Waiting for configuration update...")
		time.Sleep(8 * time.Second) // In production, consider polling until the update is confirmed.

		// Invoke the Lambda function with the (optional) test event data.
		invokeResp, err := lambdaClient.InvokeFunction(ctx, lambdaARN, testEvent)
		if err != nil {
			log.Printf("Error invoking function at %d MB: %v", mem, err)
			continue
		}

		// Parse the logs from the invocation to extract duration.
		durationMs, err := awsclient.ParseDurationFromLogs(invokeResp.LogResult)
		if err != nil {
			log.Printf("Error parsing duration from logs at %d MB: %v", mem, err)
			durationMs = 0
		} else {
			fmt.Printf("Execution duration: %.2f ms\n", durationMs)
		}

		// Calculate cost for this configuration.
		costPerMs := priceTable[mem]
		cost := durationMs * costPerMs

		results = append(results, TestResult{
			Memory:     mem,
			DurationMs: durationMs,
			Cost:       cost,
			Logs:       invokeResp.LogResult,
		})
	}

	// Revert the function configuration to its original state.
	fmt.Printf("\nReverting function configuration to original memory size: %d MB\n", originalMemory)
	if err := lambdaClient.UpdateFunctionMemory(ctx, lambdaARN, originalMemory); err != nil {
		log.Printf("Error reverting function configuration: %v", err)
	}

	// Analyze results to find the performance sweet spot and cost-effective sweet spot.
	if len(results) == 0 {
		fmt.Println("No test results were collected.")
		return
	}

	perfBest := results[0]
	costBest := results[0]
	for _, res := range results {
		if res.DurationMs < perfBest.DurationMs {
			perfBest = res
		}
		if res.Cost < costBest.Cost {
			costBest = res
		}
	}

	fmt.Println("\n--- Test Summary ---")
	fmt.Printf("Performance Sweet Spot: %d MB with duration %.2f ms\n", perfBest.Memory, perfBest.DurationMs)
	fmt.Printf("Cost Effective Sweet Spot: %d MB with cost $%s\n",
		costBest.Memory, strconv.FormatFloat(costBest.Cost, 'f', 10, 64))

	// Ask the user if they want to see a visualization of the results.
	showVis, err := prompts.PromptYesNo("Would you like to see a visualization of the results?")
	if err != nil {
		log.Fatalf("Error during prompt: %v", err)
	}
	if showVis {
		var mems []int
		var durationsArr []float64
		var costArr []float64
		for _, res := range results {
			mems = append(mems, res.Memory)
			durationsArr = append(durationsArr, res.DurationMs)
			costArr = append(costArr, res.Cost)
		}
		err = visualization.GenerateVisualization(mems, durationsArr, costArr)
		if err != nil {
			log.Printf("Error generating visualization: %v", err)
		} else {
			fmt.Println("Visualization saved as visualization.html")
		}
	}
}
