package prompts

import (
	"errors"

	"github.com/manifoldco/promptui"
)

// PromptForLambdaARN asks the user to enter the Lambda ARN.
func PromptForLambdaARN() (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("Lambda ARN cannot be empty")
		}
		// Further ARN format validation can be added here.
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter the Lambda function ARN",
		Validate: validate,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

// PromptYesNo asks a yes/no question and returns true for yes, false for no.
func PromptYesNo(label string) (bool, error) {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}

// PromptForJSON asks the user to paste JSON data.
func PromptForJSON(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}
