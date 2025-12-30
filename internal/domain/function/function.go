package function

import "github.com/aws/aws-sdk-go-v2/service/lambda/types"

// Function represents a Lambda function domain entity
type Function struct {
	name      string
	runtime   types.Runtime
	state     types.State
	vpcConfig *VPCConfig
}

// VPCConfig represents the VPC configuration of a Lambda function
type VPCConfig struct {
	VPCId                   string
	SubnetIds               []string
	SecurityGroupIds        []string
	IPv6AllowedForDualStack bool
}

// NewFunction creates a new Function entity
func NewFunction(name string, runtime types.Runtime, state types.State, vpcConfig *VPCConfig) *Function {
	return &Function{
		name:      name,
		runtime:   runtime,
		state:     state,
		vpcConfig: vpcConfig,
	}
}

// Name returns the function name
func (f *Function) Name() string {
	return f.name
}

// Runtime returns the function runtime
func (f *Function) Runtime() types.Runtime {
	return f.runtime
}

// State returns the function state
func (f *Function) State() types.State {
	return f.state
}

// VPCConfig returns the VPC configuration
func (f *Function) VPCConfig() *VPCConfig {
	return f.vpcConfig
}

// IsAttachedToVPC checks if the function is attached to a VPC
func (f *Function) IsAttachedToVPC() bool {
	return f.vpcConfig != nil && len(f.vpcConfig.SubnetIds) > 0
}

// HasIPv6Enabled checks if IPv6 is enabled for dual stack
func (f *Function) HasIPv6Enabled() bool {
	return f.vpcConfig != nil && f.vpcConfig.IPv6AllowedForDualStack
}
