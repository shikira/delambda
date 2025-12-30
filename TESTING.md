# Testing Guide for rm-lambda

This guide provides step-by-step instructions for testing the `rm-lambda` tool.

## Setup Test Environment

### 1. Deploy test infrastructure

```bash
cd test-infrastructure
npm install
npm run build
cdk deploy
```

Wait for deployment to complete (usually 3-5 minutes).

### 2. Verify deployment

```bash
cd ..
./rm-lambda list
```

You should see the three test functions:
- `test-vpc-lambda-ipv6` - VPC attached with IPv6 enabled
- `test-vpc-lambda-no-ipv6` - VPC attached without IPv6
- `test-no-vpc-lambda` - No VPC attachment

## Test Scenarios

### Scenario 1: List Functions

**Command:**
```bash
./rm-lambda list
```

**Expected Output:**
```
Found 3+ Lambda functions:

  - test-vpc-lambda-ipv6 [python3.11] VPC: vpc-xxxxx (IPv6 enabled)
  - test-vpc-lambda-no-ipv6 [nodejs18.x] VPC: vpc-xxxxx
  - test-no-vpc-lambda [python3.11] No VPC
```

### Scenario 2: Detach VPC with IPv6 Disabled First

**Command:**
```bash
./rm-lambda detach-vpc test-vpc-lambda-ipv6
```

**What happens:**
1. Disables IPv6 for the function (default behavior)
2. Waits for the IPv6 disable operation to complete
3. Detaches the VPC from the function
4. Waits for the VPC detachment to complete

**Expected Output:**
```
Disabling IPv6 for function test-vpc-lambda-ipv6...
IPv6 disabled successfully
Detaching VPC from function test-vpc-lambda-ipv6...
VPC detached successfully
```

**Verification:**
```bash
./rm-lambda list
```

The function should now show "No VPC".

### Scenario 3: Delete Function with VPC Detachment and Log Cleanup

**Command:**
```bash
./rm-lambda delete --detach-vpc --with-logs test-vpc-lambda-no-ipv6
```

**What happens:**
1. Detaches the VPC from the function
2. Waits for VPC detachment to complete
3. Deletes the Lambda function
4. Deletes the CloudWatch Logs log group

**Expected Output:**
```
Deleting Lambda function test-vpc-lambda-no-ipv6...
Lambda function deleted successfully
```

**Verification:**
```bash
./rm-lambda list
aws logs describe-log-groups --log-group-name-prefix /aws/lambda/test-vpc-lambda-no-ipv6
```

The function and its log group should both be deleted.

### Scenario 4: Delete Function without VPC

**Command:**
```bash
./rm-lambda delete --with-logs test-no-vpc-lambda
```

**What happens:**
1. Deletes the Lambda function
2. Deletes the CloudWatch Logs log group

**Expected Output:**
```
Deleting Lambda function test-no-vpc-lambda...
Lambda function deleted successfully
```

### Scenario 5: Delete Only Log Group

**Setup:** First, recreate a function or use an existing one

**Command:**
```bash
./rm-lambda delete-logs test-vpc-lambda-ipv6
```

**What happens:**
1. Deletes only the CloudWatch Logs log group
2. The Lambda function remains intact

**Expected Output:**
```
Deleting CloudWatch Logs log group for test-vpc-lambda-ipv6...
Log group deleted successfully
```

## Edge Cases to Test

### Test 1: Detach VPC from non-VPC function

```bash
# After detaching VPC from test-vpc-lambda-ipv6
./rm-lambda detach-vpc test-vpc-lambda-ipv6
```

**Expected:** Error message indicating the function is not attached to a VPC.

### Test 2: Delete already deleted function

```bash
# Try to delete the same function twice
./rm-lambda delete test-no-vpc-lambda
./rm-lambda delete test-no-vpc-lambda
```

**Expected:** Error message from AWS SDK.

### Test 3: Delete log group that doesn't exist

```bash
./rm-lambda delete-logs non-existent-function
```

**Expected:** Tool should handle gracefully (ResourceNotFoundException).

## Performance Testing

### Time VPC Detachment

```bash
time ./rm-lambda detach-vpc test-vpc-lambda-ipv6
```

**Expected time:** 10-30 seconds (depending on AWS)

### Time Complete Deletion with VPC

```bash
time ./rm-lambda delete --detach-vpc --with-logs test-vpc-lambda-no-ipv6
```

**Expected time:** 10-30 seconds

## Cleanup

After testing, destroy the test infrastructure:

```bash
cd test-infrastructure
cdk destroy
```

## Troubleshooting

### Issue: Function stuck in "Pending" state

**Solution:** Wait 2-3 minutes for AWS to complete the previous operation.

### Issue: VPC detachment timeout

**Solution:** Check if there are any network interfaces still attached. They should be automatically cleaned up.

### Issue: Permission errors

**Solution:** Ensure your AWS credentials have the required permissions:
- `lambda:*`
- `logs:*`
- `ec2:DescribeNetworkInterfaces`

## Advanced Testing

### Test with AWS Regions

```bash
# Test in different region
./rm-lambda --region us-west-2 list
./rm-lambda --region eu-west-1 list
```

### Test with AWS_REGION environment variable

```bash
export AWS_REGION=ap-northeast-1
./rm-lambda list
```

### Test error handling

```bash
# Non-existent function
./rm-lambda delete non-existent-function

# Invalid region
./rm-lambda --region invalid-region list
```
