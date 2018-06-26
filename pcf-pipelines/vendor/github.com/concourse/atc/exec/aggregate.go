package exec

import (
	"context"
	"fmt"
	"strings"

	"github.com/concourse/atc/worker"
)

// Aggregate constructs a Step that will run each step in parallel.
type Aggregate []StepFactory

// Using delegates to each StepFactory and returns an AggregateStep.
func (a Aggregate) Using(repo *worker.ArtifactRepository) Step {
	sources := AggregateStep{}

	for _, step := range a {
		sources = append(sources, step.Using(repo))
	}

	return sources
}

// AggregateStep is a step of steps to run in parallel.
type AggregateStep []Step

// Run executes all steps in parallel. It will indicate that it's ready when
// all of its steps are ready, and propagate any signal received to all running
// steps.
//
// It will wait for all steps to exit, even if one step fails or errors. After
// all steps finish, their errors (if any) will be aggregated and returned as a
// single error.
func (step AggregateStep) Run(ctx context.Context) error {
	errs := make(chan error, len(step))

	for _, s := range step {
		s := s
		go func() {
			errs <- s.Run(ctx)
		}()
	}

	var errorMessages []string
	for i := 0; i < len(step); i++ {
		err := <-errs
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("one or more aggregated step errored:\n%s", strings.Join(errorMessages, "\n"))
	}

	return nil
}

// Succeeded is true if all of the steps' Succeeded is true
func (step AggregateStep) Succeeded() bool {
	succeeded := true

	for _, step := range step {
		if !step.Succeeded() {
			succeeded = false
		}
	}

	return succeeded
}
