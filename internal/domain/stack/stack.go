package stack

// Stack represents a CloudFormation stack
type Stack struct {
	name string
}

// NewStack creates a new Stack entity
func NewStack(name string) *Stack {
	return &Stack{
		name: name,
	}
}

// Name returns the stack name
func (s *Stack) Name() string {
	return s.name
}
