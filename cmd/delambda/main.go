package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/shirasu/delambda/internal/application/usecase"
	"github.com/shirasu/delambda/internal/domain/function"
	"github.com/shirasu/delambda/internal/infrastructure/repository"
	"github.com/shirasu/delambda/pkg/client"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Global flags
	region := flag.String("region", "", "AWS region")
	profile := flag.String("profile", "", "AWS profile")

	command := os.Args[1]

	switch command {
	case "list":
		handleList(region, profile)
	case "detach":
		handleDetach(region, profile)
	case "delete":
		handleDelete(region, profile)
	case "delete-logs":
		handleDeleteLogs(region, profile)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleList(region, profile *string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	stackFlag := fs.String("stack", "", "CloudFormation stack name")
	fs.Parse(os.Args[2:])

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	var functions []*function.Function

	if *stackFlag != "" {
		// List functions in a specific stack
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		stackRepo := repository.NewStackRepository(awsClient.CloudFormation)
		listStackUseCase := usecase.NewListStackFunctionsUseCase(functionRepo, stackRepo)

		functions, err = listStackUseCase.Execute(ctx, *stackFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list functions in stack: %v\n", err)
			os.Exit(1)
		}
	} else {
		// List all functions
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		listUseCase := usecase.NewListFunctionsUseCase(functionRepo)

		functions, err = listUseCase.Execute(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list functions: %v\n", err)
			os.Exit(1)
		}
	}

	if len(functions) == 0 {
		fmt.Println("No Lambda functions found")
		return
	}

	fmt.Printf("Found %d Lambda function(s):\n\n", len(functions))
	for _, fn := range functions {
		vpcInfo := "No VPC"
		if fn.VPCConfig() != nil && len(fn.VPCConfig().SubnetIds) > 0 {
			vpcInfo = fmt.Sprintf("VPC: %s", fn.VPCConfig().VPCId)
			if fn.HasIPv6Enabled() {
				vpcInfo += " (IPv6 enabled)"
			}
		}
		fmt.Printf("  - %s [%v] %s\n", fn.Name(), fn.Runtime(), vpcInfo)
	}
}

func handleDetach(region, profile *string) {
	fs := flag.NewFlagSet("detach", flag.ExitOnError)
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	lambdaFlag := fs.String("lambda", "", "Lambda function name")
	stackFlag := fs.String("stack", "", "CloudFormation stack name")
	fs.Parse(os.Args[2:])

	// Validate flags
	if *lambdaFlag == "" && *stackFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: Either --lambda or --stack must be specified")
		fmt.Fprintln(os.Stderr, "Usage: delambda detach --lambda <function-name>")
		fmt.Fprintln(os.Stderr, "       delambda detach --stack <stack-name>")
		os.Exit(1)
	}

	if *lambdaFlag != "" && *stackFlag != "" {
		fmt.Fprintln(os.Stderr, "Error: Cannot specify both --lambda and --stack")
		os.Exit(1)
	}

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	if *lambdaFlag != "" {
		// Detach VPC from a single function
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		detachVPCUseCase := usecase.NewDetachVPCUseCase(functionRepo)

		input := &usecase.DetachVPCInput{
			FunctionName: *lambdaFlag,
			DisableIPv6:  true,
		}

		if err := detachVPCUseCase.Execute(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to detach VPC: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully detached VPC from %s\n", *lambdaFlag)
	} else {
		// Detach VPC from all functions in a stack
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		stackRepo := repository.NewStackRepository(awsClient.CloudFormation)
		detachVPCStackUseCase := usecase.NewDetachVPCStackUseCase(functionRepo, stackRepo)

		input := &usecase.DetachVPCStackInput{
			StackName:   *stackFlag,
			DisableIPv6: true,
		}

		if err := detachVPCStackUseCase.Execute(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to detach VPC from stack: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully detached VPC from all functions in stack %s\n", *stackFlag)
	}
}

func handleDelete(region, profile *string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	lambdaFlag := fs.String("lambda", "", "Lambda function name")
	stackFlag := fs.String("stack", "", "CloudFormation stack name")
	withoutLogs := fs.Bool("without-logs", false, "Don't delete CloudWatch logs (logs are deleted by default)")
	fs.Parse(os.Args[2:])

	// Validate flags
	if *lambdaFlag == "" && *stackFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: Either --lambda or --stack must be specified")
		fmt.Fprintln(os.Stderr, "Usage: delambda delete --lambda <function-name>")
		fmt.Fprintln(os.Stderr, "       delambda delete --stack <stack-name>")
		os.Exit(1)
	}

	if *lambdaFlag != "" && *stackFlag != "" {
		fmt.Fprintln(os.Stderr, "Error: Cannot specify both --lambda and --stack")
		os.Exit(1)
	}

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	// Delete logs by default (unless --without-logs is specified)
	deleteLogs := !*withoutLogs

	if *lambdaFlag != "" {
		// Delete a single function
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		logGroupRepo := repository.NewLogGroupRepository(awsClient.Logs)
		deleteUseCase := usecase.NewDeleteFunctionUseCase(functionRepo, logGroupRepo, os.Stdout)

		input := &usecase.DeleteFunctionInput{
			FunctionName: *lambdaFlag,
			DetachVPC:    true,
			DisableIPv6:  true,
			DeleteLogs:   deleteLogs,
		}

		if err := deleteUseCase.Execute(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete function: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSuccessfully deleted function %s\n", *lambdaFlag)
	} else {
		// Delete all functions in a stack
		functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
		logGroupRepo := repository.NewLogGroupRepository(awsClient.Logs)
		stackRepo := repository.NewStackRepository(awsClient.CloudFormation)
		deleteStackUseCase := usecase.NewDeleteStackFunctionsUseCase(functionRepo, logGroupRepo, stackRepo, os.Stdout)

		input := &usecase.DeleteStackFunctionsInput{
			StackName:   *stackFlag,
			DetachVPC:   true,
			DisableIPv6: true,
			DeleteLogs:  deleteLogs,
		}

		if err := deleteStackUseCase.Execute(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete stack functions: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSuccessfully deleted all functions in stack %s\n", *stackFlag)
	}
}

func handleDeleteLogs(region, profile *string) {
	fs := flag.NewFlagSet("delete-logs", flag.ExitOnError)
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Log group name is required")
		fmt.Fprintln(os.Stderr, "Usage: delambda delete-logs <log-group-name>")
		os.Exit(1)
	}

	logGroupName := args[0]

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	logGroupRepo := repository.NewLogGroupRepository(awsClient.Logs)
	deleteLogsUseCase := usecase.NewDeleteLogGroupUseCase(logGroupRepo)

	if err := deleteLogsUseCase.Execute(ctx, logGroupName); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete log group: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted log group %s\n", logGroupName)
}

func printUsage() {
	usage := `delambda - A powerful CLI tool to safely delete AWS Lambda functions with VPC attachments

Usage:
  delambda <command> [options]

Commands:
  list                 List all Lambda functions with VPC status
  detach               Detach VPC from a Lambda function
  delete               Delete a Lambda function
  delete-logs          Delete a CloudWatch Logs log group
  help                 Show this help message

Global Options:
  --region string      AWS region (optional, uses default or AWS_REGION env var)
  --profile string     AWS profile (optional, uses default or AWS_PROFILE env var)

Examples:
  # List all Lambda functions in the account and region
  delambda list

  # List Lambda functions in a specific CloudFormation stack
  delambda list --stack my-stack

  # Detach VPC from a single Lambda function
  delambda detach --lambda my-function

  # Detach VPC from all Lambda functions in a CloudFormation stack
  delambda detach --stack my-stack

  # Delete a Lambda function and its log group (VPC will be automatically detached if attached)
  delambda delete --lambda my-function

  # Delete a Lambda function without deleting its log group
  delambda delete --lambda my-function --without-logs

  # Delete all Lambda functions in a CloudFormation stack (including log groups)
  delambda delete --stack my-stack

  # Delete CloudWatch Logs log group
  delambda delete-logs /aws/lambda/my-function
`
	fmt.Print(usage)
}
