package ops

import (
	"fmt"
	"log"
	"time"
)

type Operation interface {
	Retries() int
	Run(input interface{}) (output interface{}, err error)
}
type Pipeline struct {
	steps   []Operation
	results map[int]interface{}
}

func NewPipeline(steps ...Operation) *Pipeline {
	return &Pipeline{steps: steps, results: make(map[int]interface{})}
}

func (p *Pipeline) Execute(initialInput interface{}) (interface{}, error) {
	var err error
	output := initialInput

	now := time.Now()
	for index, step := range p.steps {
		log.Println("Execute operation", index)
		output, err = p.runWithRetries(step, output, step.Retries())
		if err != nil {
			log.Println("Operation", index, "failed in", time.Since(now))
			return nil, err // Terminate pipeline if a step fails after retries
		}
		// Store the successful output for this operation
		p.results[index] = output
		log.Println("Operation", index, "completed in", time.Since(now))
	}
	return output, nil
}

// GetResult returns the output of a specific operation by its index
func (p *Pipeline) GetResult(stepIndex int) (interface{}, bool) {
	result, exists := p.results[stepIndex]
	return result, exists
}

func (p *Pipeline) runWithRetries(op Operation, input interface{}, retries int) (interface{}, error) {
	var output interface{}
	var err error

	for attempt := 1; attempt <= retries; attempt++ {
		output, err = op.Run(input)
		if err == nil {
			return output, nil
		}
		log.Printf("Attempt %d/%d for operation %T failed: %v", attempt, retries, op, err)
	}

	return nil, fmt.Errorf("operation %T failed after %d retries: %w", op, retries, err)
}
