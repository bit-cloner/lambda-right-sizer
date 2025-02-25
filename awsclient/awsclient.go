package awsclient

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaClient wraps the AWS Lambda client.
type LambdaClient struct {
	Client *lambda.Client
}

// NewLambdaClient creates a new Lambda client for the specified region.
func NewLambdaClient(ctx context.Context, region string) (*LambdaClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &LambdaClient{
		Client: lambda.NewFromConfig(cfg),
	}, nil
}

// FunctionMetadata holds minimal Lambda function configuration data.
type FunctionMetadata struct {
	MemorySize    int
	Architectures []string
}

// GetFunctionMetadata fetches the Lambda function configuration.
func (lc *LambdaClient) GetFunctionMetadata(ctx context.Context, functionARN string) (*FunctionMetadata, error) {
	out, err := lc.Client.GetFunctionConfiguration(ctx, &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionARN),
	})
	if err != nil {
		return nil, err
	}
	archs := []string{}
	// Some functions might have nil Architectures field; default to x86_64.
	if out.Architectures != nil && len(out.Architectures) > 0 {
		for _, a := range out.Architectures {
			archs = append(archs, string(a))
		}
	} else {
		archs = append(archs, "x86_64")
	}
	return &FunctionMetadata{
		MemorySize:    int(aws.ToInt32(out.MemorySize)),
		Architectures: archs,
	}, nil
}

// UpdateFunctionMemory updates the Lambda function's memory configuration.
func (lc *LambdaClient) UpdateFunctionMemory(ctx context.Context, functionARN string, memorySize int) error {
	_, err := lc.Client.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionARN),
		MemorySize:   aws.Int32(int32(memorySize)),
	})
	return err
}

// InvokeResponse wraps the response from invoking a Lambda function.
type InvokeResponse struct {
	LogResult string
	Payload   []byte
}

// InvokeFunction invokes the Lambda function with an optional payload.
func (lc *LambdaClient) InvokeFunction(ctx context.Context, functionARN string, payload []byte) (*InvokeResponse, error) {
	out, err := lc.Client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionARN),
		Payload:        payload,
		LogType:        lambdaTypes.LogTypeTail,
		InvocationType: lambdaTypes.InvocationTypeRequestResponse,
	})
	if err != nil {
		return nil, err
	}
	return &InvokeResponse{
		LogResult: aws.ToString(out.LogResult),
		Payload:   out.Payload,
	}, nil
}

// ParseDurationFromLogs decodes the base64-encoded logs and extracts the Duration in milliseconds.
// It expects a log line containing "Duration:" followed by the value in ms.
func ParseDurationFromLogs(logResult string) (float64, error) {
	if logResult == "" {
		return 0, errors.New("no log result available")
	}

	decoded, err := base64.StdEncoding.DecodeString(logResult)
	if err != nil {
		return 0, fmt.Errorf("failed to decode log result: %v", err)
	}

	// Example log line: "REPORT RequestId: ... Duration: 12.34 ms Billed Duration: ..."
	re := regexp.MustCompile(`Duration:\s+([\d\.]+)\s+ms`)
	matches := re.FindStringSubmatch(string(decoded))
	if len(matches) < 2 {
		return 0, errors.New("duration not found in logs")
	}
	return strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
}
