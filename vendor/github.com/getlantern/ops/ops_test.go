package ops_test

import (
	"sync"
	"testing"

	"github.com/getlantern/errors"
	"github.com/getlantern/ops"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	var reportedFailure error
	var reportedCtx map[string]interface{}
	report := func(failure error, ctx map[string]interface{}) {
		reportedFailure = failure
		reportedCtx = ctx
	}

	ops.RegisterReporter(report)
	ops.SetGlobal("g", "g1")
	op := ops.Begin("test_success").Set("a", 1).SetDynamic("b", func() interface{} { return 2 })
	defer op.End()
	innerOp := op.Begin("inside")
	innerOp.FailIf(nil)
	innerOp.End()

	assert.Nil(t, reportedFailure)
	expectedCtx := map[string]interface{}{
		"op":      "inside",
		"root_op": "test_success",
		"g":       "g1",
		"a":       1,
		"b":       2,
	}
	assert.Equal(t, expectedCtx, reportedCtx)
}

func TestFailure(t *testing.T) {
	doTestFailure(t, false)
}

func TestCancel(t *testing.T) {
	doTestFailure(t, true)
}

func doTestFailure(t *testing.T, cancel bool) {
	var reportedFailure error
	var reportedCtx map[string]interface{}
	report := func(failure error, ctx map[string]interface{}) {
		reportedFailure = failure
		reportedCtx = ctx
	}

	ops.RegisterReporter(report)
	op := ops.Begin("test_failure")
	var wg sync.WaitGroup
	wg.Add(1)
	op.Go(func() {
		op.FailIf(errors.New("I failed").With("errorcontext", 5))
		wg.Done()
	})
	wg.Wait()
	if cancel {
		op.Cancel()
	}
	op.End()

	if cancel {
		assert.Nil(t, reportedFailure)
		assert.Nil(t, reportedCtx)
	} else {
		assert.Contains(t, reportedFailure.Error(), "I failed")
		assert.Equal(t, 5, reportedCtx["errorcontext"])
	}
}
