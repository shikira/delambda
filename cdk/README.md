# Test Infrastructure for rm-lambda

This CDK project creates a test environment for the `rm-lambda` tool.

## What it creates

1. **Dual Stack VPC** - VPC with both IPv4 and IPv6 support
2. **Three Lambda functions** for testing:
   - `test-vpc-lambda-ipv6` - Lambda with VPC attachment and IPv6 enabled
   - `test-vpc-lambda-no-ipv6` - Lambda with VPC attachment but IPv6 disabled
   - `test-no-vpc-lambda` - Lambda without VPC attachment

## Prerequisites

- AWS CLI configured with credentials
- Node.js 18+ installed
- pnpm installed (`npm install -g pnpm`)
- CDK CLI installed (`pnpm add -g aws-cdk`)

## Deployment

### 1. Install dependencies

```bash
cd test-infrastructure
pnpm install
```

### 2. Bootstrap CDK (first time only)

```bash
cdk bootstrap
```

### 3. Build and deploy

```bash
pnpm run build
cdk deploy
```

This will create all resources in your default AWS account/region.

## Testing rm-lambda

After deployment, you can test the `rm-lambda` tool:

### List all Lambda functions

```bash
cd ..
./rm-lambda list
```

Expected output:
```
Found 3+ Lambda functions:

  - test-vpc-lambda-ipv6 [python3.11] VPC: vpc-xxxxx (IPv6 enabled)
  - test-vpc-lambda-no-ipv6 [nodejs18.x] VPC: vpc-xxxxx
  - test-no-vpc-lambda [python3.11] No VPC
```

### Test VPC detachment

```bash
# Detach VPC from the IPv6-enabled function
./rm-lambda detach-vpc test-vpc-lambda-ipv6
```

This will:
1. Disable IPv6 (since it's enabled by default with --disable-ipv6 flag)
2. Detach the VPC from the function

### Test deletion with VPC detachment

```bash
# Delete function with VPC detachment and log cleanup
./rm-lambda delete --detach-vpc --with-logs test-vpc-lambda-no-ipv6
```

This will:
1. Detach the VPC
2. Delete the Lambda function
3. Delete the CloudWatch Logs log group

### Test simple deletion (no VPC)

```bash
# Delete the non-VPC function
./rm-lambda delete --with-logs test-no-vpc-lambda
```

## Cleanup

To destroy all resources:

```bash
cd test-infrastructure
cdk destroy
```

## Cost

This infrastructure should cost minimal amounts (mostly for VPC NAT Gateway if used). Remember to destroy the stack when not in use.
