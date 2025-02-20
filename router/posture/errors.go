package posture

import (
	"fmt"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/pkg/errors"
)

type CheckError struct {
	Id    string
	Name  string
	Cause error
}

func (p *CheckError) Error() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("posture check %s [%s] failed due to error(s): %s", p.Name, p.Id, p.Cause.Error())
}

type NoPoliciesError struct {
}

func (e *NoPoliciesError) Error() string {
	logtrace.LogWithFunctionName()
	return "no policies provide access"
}

type PolicyAccessError struct {
	Id     string
	Name   string
	Errors []error
}

func (p *PolicyAccessError) Error() string {
	logtrace.LogWithFunctionName()
	if len(p.Errors) == 0 {
		return fmt.Sprintf("policy %s [%s] failed", p.Id, p.Name)
	}

	subErrStr := ""
	for i, err := range p.Errors {
		if i == 0 {
			subErrStr = err.Error()
		} else {
			subErrStr = subErrStr + ", " + err.Error()
		}
	}
	return fmt.Sprintf("policy %s [%s] failed due to %d error(s): %s", p.Name, p.Id, len(p.Errors), subErrStr)
}

type PolicyAccessErrors []*PolicyAccessError

func (pae *PolicyAccessErrors) Error() string {
	logtrace.LogWithFunctionName()
	if pae == nil || len(*pae) == 0 {
		return "unknown policy failure"
	}

	subErr := ""

	for i, err := range *pae {
		if i == 0 {
			subErr = err.Error()
		} else {
			subErr = subErr + ", " + err.Error()
		}
	}

	return fmt.Sprintf("%d policies failed: %s", len(*pae), subErr)
}

func EvaluatePostureCheck(postureCheck *edge_ctrl_pb.DataState_PostureCheck, cache *Cache) *CheckError {
	logtrace.LogWithFunctionName()
	check := CtrlCheckToLogic(postureCheck)
	return check.Evaluate(cache)
}

// FailedValueError represents a complex object comparison that failed. If a simple comparison failure is needed
// (i.e. bool != bool, string != string) use an `error` instead.
type FailedValueError[V fmt.Stringer] struct {
	ExpectedValue V
	GivenValue    V
	Reason        error
}

func (v *FailedValueError[V]) String() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("the state did not match because %v, expected: %s, given: %s", v.Reason, v.ExpectedValue, v.GivenValue)
}

func (v *FailedValueError[V]) Error() string {
	logtrace.LogWithFunctionName()
	return v.String()
}

// AllInListError indicates that a given array of expected values had one or more values that did not match/pass.
// GivenValues represents all values supplied to match the expected values. FailedValues represents all the expected
// values that did not pass.
type AllInListError[V fmt.Stringer] struct {
	FailedValues []FailedValueError[V]
	GivenValues  []V
}

func (e *AllInListError[V]) Error() string {
	logtrace.LogWithFunctionName()
	var failureStrings []string

	for _, failedValue := range e.FailedValues {
		failureStrings = append(failureStrings, failedValue.String())
	}

	valueStr := ""

	for i, v := range e.GivenValues {
		if i == 0 {
			valueStr = fmt.Sprintf("%v", v)
		} else {
			valueStr = valueStr + ", " + fmt.Sprintf("%v", v)
		}
	}

	return fmt.Sprintf("all values must be valid have at least one failure, have: %s, failed for: %v", valueStr, failureStrings)
}

// AnyInListError represents the fact that zero expected values did not match/pass where at least one was required.
// GivenValues represents all values supplied to match the expected values. FailedValues represents all the expected
// values that did not pass.
type AnyInListError[V fmt.Stringer] struct {
	FailedValues []FailedValueError[V]
	GivenValues  []V
}

func (e *AnyInListError[V]) Error() string {
	logtrace.LogWithFunctionName()
	var failureStrings []string

	for _, failedValue := range e.FailedValues {
		failureStrings = append(failureStrings, failedValue.String())
	}

	valueStr := ""

	for i, v := range e.GivenValues {
		if i == 0 {
			valueStr = fmt.Sprintf("%v", v)
		} else {
			valueStr = valueStr + ", " + fmt.Sprintf("%v", v)
		}
	}

	return fmt.Sprintf("one valid values is required, got 0, have: %s, failed for: %v", valueStr, failureStrings)
}

// OneInListError represents two arrays of values where one of the supplied GivenValues must be in the ValidValues.
// Used when a large cross join of values (i.e. mac address approve/deny lists) would be reported for every comparison.
type OneInListError[V fmt.Stringer] struct {
	ValidValues []V
	GivenValues []V
}

func (e *OneInListError[V]) Error() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("none of the given values were in the valid values, given: %v, valid: %v", e.GivenValues, e.ValidValues)
}

type Str string

func (s Str) String() string {
	logtrace.LogWithFunctionName()
	return string(s)
}

var NilStateError = errors.New("posture state was nil")

var NotEqualError = errors.New("the values were not equal")
