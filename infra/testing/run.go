package testing

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime/debug"

	"github.com/wojciech-sif/localnet/infra"
	"github.com/wojciech-sif/localnet/lib/logger"
	"go.uber.org/zap"
)

type errPanic struct {
	value interface{}
	stack []byte
}

func (err errPanic) Error() string {
	return fmt.Sprintf("test panicked: %s\n\n%s", err.value, err.stack)
}

// Run deploys testing environment and runs tests there
func Run(ctx context.Context, target infra.Target, env infra.Set, tests []*T, filters []*regexp.Regexp) error {
	toRun := make([]*T, 0, len(tests))
	for _, test := range tests {
		if !matchesAny(test.name, filters) {
			continue
		}
		if err := test.prepare(ctx); err != nil {
			return err
		}
		toRun = append(toRun, test)
	}

	if len(toRun) == 0 {
		return errors.New("there are no tests to run")
	}

	if err := target.Deploy(ctx, env); err != nil {
		return err
	}

	failed := false
	for _, t := range toRun {
		runTest(logger.With(ctx, zap.String("test", t.name)), t)
		failed = failed || t.failed
	}
	if failed {
		return errors.New("tests failed")
	}
	logger.Get(ctx).Info("All tests succeeded")
	return nil
}

func matchesAny(val string, regs []*regexp.Regexp) bool {
	if len(regs) == 0 {
		return true
	}
	for _, reg := range regs {
		if reg.MatchString(val) {
			return true
		}
	}
	return false
}

func runTest(ctx context.Context, t *T) {
	log := logger.Get(ctx)
	log.Info("Test started")
	defer func() {
		log.Info("Test finished")

		r := recover()
		switch {
		// Panic in tested code causes failure of test.
		// Panic caused by T.FailNow is ignored (r != t) as it is used only to exit the test after first failure.
		case r != nil && r != t:
			t.failed = true
			t.errors = append(t.errors, errPanic{value: r, stack: debug.Stack()})
			log.Error("Test panicked", zap.Any("panic", r))
		case t.failed:
			for _, e := range t.errors {
				log.Error("Test failed", zap.Error(e))
			}
		default:
			log.Info("Test succeeded")
		}
	}()
	t.run(ctx, t)
}
