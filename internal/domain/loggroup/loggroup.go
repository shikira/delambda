package loggroup

import "fmt"

// LogGroup represents a CloudWatch Logs log group domain entity
type LogGroup struct {
	name string
}

// NewLogGroup creates a new LogGroup entity
func NewLogGroup(name string) *LogGroup {
	return &LogGroup{
		name: name,
	}
}

// NewLogGroupForFunction creates a LogGroup for a Lambda function
func NewLogGroupForFunction(functionName string) *LogGroup {
	return &LogGroup{
		name: fmt.Sprintf("/aws/lambda/%s", functionName),
	}
}

// Name returns the log group name
func (lg *LogGroup) Name() string {
	return lg.name
}
