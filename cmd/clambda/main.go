package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/shirasu/clambda/internal/application/usecase"
	"github.com/shirasu/clambda/internal/infrastructure/repository"
	"github.com/shirasu/clambda/pkg/client"
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
	case "detach-vpc":
		handleDetachVPC(region, profile)
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
	fs.Parse(os.Args[2:])

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
	listUseCase := usecase.NewListFunctionsUseCase(functionRepo)

	functions, err := listUseCase.Execute(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list functions: %v\n", err)
		os.Exit(1)
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

func handleDetachVPC(region, profile *string) {
	fs := flag.NewFlagSet("detach-vpc", flag.ExitOnError)
	disableIPv6 := fs.Bool("disable-ipv6", true, "Disable IPv6 before detaching VPC")
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Function name is required")
		fmt.Fprintln(os.Stderr, "Usage: clambda detach-vpc [--disable-ipv6=true] <function-name>")
		os.Exit(1)
	}

	functionName := args[0]

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
	detachVPCUseCase := usecase.NewDetachVPCUseCase(functionRepo)

	input := &usecase.DetachVPCInput{
		FunctionName: functionName,
		DisableIPv6:  *disableIPv6,
	}

	if err := detachVPCUseCase.Execute(ctx, input); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to detach VPC: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully detached VPC from %s\n", functionName)
}

func handleDelete(region, profile *string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	detachVPC := fs.Bool("detach-vpc", false, "Detach VPC before deleting")
	disableIPv6 := fs.Bool("disable-ipv6", false, "Disable IPv6 before detaching VPC")
	withLogs := fs.Bool("with-logs", false, "Delete associated CloudWatch logs")
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Function name is required")
		fmt.Fprintln(os.Stderr, "Usage: clambda delete [--detach-vpc] [--disable-ipv6] [--with-logs] <function-name>")
		os.Exit(1)
	}

	functionName := args[0]

	ctx := context.Background()
	awsClient, err := client.NewAWSClient(ctx, *regionFlag, *profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create AWS client: %v\n", err)
		os.Exit(1)
	}

	functionRepo := repository.NewFunctionRepository(awsClient.Lambda)
	logGroupRepo := repository.NewLogGroupRepository(awsClient.Logs)
	deleteUseCase := usecase.NewDeleteFunctionUseCase(functionRepo, logGroupRepo)

	input := &usecase.DeleteFunctionInput{
		FunctionName: functionName,
		DetachVPC:    *detachVPC,
		DisableIPv6:  *disableIPv6,
		DeleteLogs:   *withLogs,
	}

	if err := deleteUseCase.Execute(ctx, input); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete function: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted function %s\n", functionName)
}

func handleDeleteLogs(region, profile *string) {
	fs := flag.NewFlagSet("delete-logs", flag.ExitOnError)
	regionFlag := fs.String("region", *region, "AWS region")
	profileFlag := fs.String("profile", *profile, "AWS profile")
	fs.Parse(os.Args[2:])

	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Log group name is required")
		fmt.Fprintln(os.Stderr, "Usage: clambda delete-logs <log-group-name>")
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
	usage := `clambda - A powerful CLI tool to safely delete AWS Lambda functions with VPC attachments

Usage:
  clambda <command> [options]

Commands:
  list                 List all Lambda functions with VPC status
  detach-vpc           Detach VPC from a Lambda function
  delete               Delete a Lambda function
  delete-logs          Delete a CloudWatch Logs log group
  help                 Show this help message

Global Options:
  --region string      AWS region (optional, uses default or AWS_REGION env var)
  --profile string     AWS profile (optional, uses default or AWS_PROFILE env var)

Examples:
  # List all Lambda functions
  clambda list

  # Detach VPC from a function
  clambda detach-vpc --disable-ipv6=true my-function

  # Delete a function with VPC detachment and log cleanup
  clambda delete --detach-vpc --disable-ipv6 --with-logs my-function

  # Delete CloudWatch Logs log group
  clambda delete-logs /aws/lambda/my-function
`
	fmt.Print(usage)
}
