import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as logs from 'aws-cdk-lib/aws-logs';
import { Construct } from 'constructs';

export class TestInfrastructureStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Dual Stack VPC
    const vpc = new ec2.Vpc(this, 'TestVPC', {
      maxAzs: 2,
      ipProtocol: ec2.IpProtocol.DUAL_STACK,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: 'Private',
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      ],
    });

    // Security Group for Lambda
    const lambdaSecurityGroup = new ec2.SecurityGroup(this, 'LambdaSecurityGroup', {
      vpc,
      description: 'Security group for test Lambda functions',
      allowAllOutbound: true,
    });

    // Lambda function with VPC attachment and IPv6 enabled (dual stack)
    const vpcLambdaWithIPv6 = new lambda.Function(this, 'VpcLambdaWithIPv6', {
      runtime: lambda.Runtime.PYTHON_3_11,
      handler: 'index.handler',
      code: lambda.Code.fromInline(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from VPC Lambda with IPv6!'
    }
      `),
      functionName: 'test-vpc-lambda-ipv6',
      vpc: vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      securityGroups: [lambdaSecurityGroup],
      ipv6AllowedForDualStack: true,
      logRetention: logs.RetentionDays.ONE_DAY,
      timeout: cdk.Duration.seconds(30),
    });

    // Lambda function with VPC attachment but IPv6 disabled
    const vpcLambdaNoIPv6 = new lambda.Function(this, 'VpcLambdaNoIPv6', {
      runtime: lambda.Runtime.NODEJS_18_X,
      handler: 'index.handler',
      code: lambda.Code.fromInline(`
exports.handler = async (event) => {
    return {
        statusCode: 200,
        body: 'Hello from VPC Lambda without IPv6!'
    };
};
      `),
      functionName: 'test-vpc-lambda-no-ipv6',
      vpc: vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      securityGroups: [lambdaSecurityGroup],
      ipv6AllowedForDualStack: false,
      logRetention: logs.RetentionDays.ONE_DAY,
      timeout: cdk.Duration.seconds(30),
    });

    // Lambda function without VPC
    const noVpcLambda = new lambda.Function(this, 'NoVpcLambda', {
      runtime: lambda.Runtime.PYTHON_3_11,
      handler: 'index.handler',
      code: lambda.Code.fromInline(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from Lambda without VPC!'
    }
      `),
      functionName: 'test-no-vpc-lambda',
      logRetention: logs.RetentionDays.ONE_DAY,
      timeout: cdk.Duration.seconds(30),
    });

    // Outputs
    new cdk.CfnOutput(this, 'VpcId', {
      value: vpc.vpcId,
      description: 'VPC ID',
    });

    new cdk.CfnOutput(this, 'VpcLambdaWithIPv6Name', {
      value: vpcLambdaWithIPv6.functionName,
      description: 'Lambda function with VPC and IPv6 enabled',
    });

    new cdk.CfnOutput(this, 'VpcLambdaNoIPv6Name', {
      value: vpcLambdaNoIPv6.functionName,
      description: 'Lambda function with VPC but IPv6 disabled',
    });

    new cdk.CfnOutput(this, 'NoVpcLambdaName', {
      value: noVpcLambda.functionName,
      description: 'Lambda function without VPC',
    });
  }
}
