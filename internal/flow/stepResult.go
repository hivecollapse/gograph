package flow

import (
	"fmt"
	"gograph/internal/log"
)

type StepResult struct {
	Name   string
	Result any
	State  map[string]interface{}
	Errors []error
}

func (step *StepResult) HasError() bool {
	return step.Errors != nil && len(step.Errors) > 0
}

func (step *StepResult) Printf(format string, args ...any) {
	log.Printf("Step[%v] %v", step.Name, fmt.Sprintf(format, args...))
}

func (step *StepResult) Verbosef(format string, args ...any) {
	log.Verbosef("Step[%v] %v", step.Name, fmt.Sprintf(format, args...))
}

func (step *StepResult) Debugf(format string, args ...any) {
	log.Debugf("Step[%v] %v", step.Name, fmt.Sprintf(format, args...))
}

func (step *StepResult) Errorf(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	log.Printf("Step[%v] ERROR %v", step.Name, err)
	step.Errors = append(step.Errors, err)
}
