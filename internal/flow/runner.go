package flow

import (
	"fmt"
	"gograph/internal/log"
	"io"
)

type FlowRunner struct {
	flow    *FlowDefinition
	step    uint
	results []*StepResult
}

type RunOption struct {
	BreakAfterStepNamed string
}

func (f *FlowRunner) Run(options *RunOption) {
	f.results = make([]*StepResult, len(f.flow.Steps))
	for i, step := range f.flow.Steps {

		result := f.RunStep(uint(i))

		if len(options.BreakAfterStepNamed) > 0 {

			if step.Name == options.BreakAfterStepNamed {
				break
			}
		}
		log.Println("")

		if result.HasError() && !step.Result.ContinueOnFailure {
			log.Println("Aborting")
			log.Println("")

			return
		}

	}
}

func (f *FlowRunner) RunStep(step uint) *StepResult {
	if f.step >= uint(len(f.flow.Steps)) {
		log.Fatalln("Step overflow", step, ">=", len(f.flow.Steps))
	}
	stepDef := f.flow.Steps[step]

	result := stepDef.Run(f.flow)
	// Store result
	f.results = append(f.results, result)
	f.DumpStepResult(result)
	return result
}

func (f *FlowRunner) Summarize(w io.Writer) {

	w.Write([]byte(fmt.Sprintf("[%v] Summary\n", f.flow.Name)))
	for _, result := range f.results {
		if result == nil {
			continue
		}
		var status string
		if result.HasError() {
			status = "ER"
		} else {
			status = "OK"
		}

		w.Write([]byte(fmt.Sprintf("- [%v] %v\n", status, result.Name)))
		if result.HasError() {
			for _, err := range result.Errors {
				w.Write([]byte(fmt.Sprintf("  - %v\n", err)))
			}
		}
	}
}

func (f *FlowRunner) DumpStepResult(result *StepResult) {
	if f.step >= uint(len(f.results)) {
		return
	}

	rstring := "OK"
	if result.HasError() {
		rstring = "KO"
	}

	log.Printf("[%v] results: %v", result.Name, rstring)
	for _, err := range result.Errors {
		log.Println("    - ", err)
	}
	if len(result.State) > 0 {
		log.Println("  state:", len(result.State))
		for k, v := range result.State {
			log.Printf("    - %v: %v", k, v)
		}
	}
}

func NewFlowRunner(flow *FlowDefinition) *FlowRunner {

	return &FlowRunner{
		flow: flow,
		step: 0,
	}
}
