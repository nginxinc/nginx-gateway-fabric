// Code generated by counterfeiter. DO NOT EDIT.
package validationfakes

import (
	"sync"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

type FakeGenericValidator struct {
	ValidateEscapedStringNoVarExpansionStub        func(string) error
	validateEscapedStringNoVarExpansionMutex       sync.RWMutex
	validateEscapedStringNoVarExpansionArgsForCall []struct {
		arg1 string
	}
	validateEscapedStringNoVarExpansionReturns struct {
		result1 error
	}
	validateEscapedStringNoVarExpansionReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansion(arg1 string) error {
	fake.validateEscapedStringNoVarExpansionMutex.Lock()
	ret, specificReturn := fake.validateEscapedStringNoVarExpansionReturnsOnCall[len(fake.validateEscapedStringNoVarExpansionArgsForCall)]
	fake.validateEscapedStringNoVarExpansionArgsForCall = append(fake.validateEscapedStringNoVarExpansionArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ValidateEscapedStringNoVarExpansionStub
	fakeReturns := fake.validateEscapedStringNoVarExpansionReturns
	fake.recordInvocation("ValidateEscapedStringNoVarExpansion", []interface{}{arg1})
	fake.validateEscapedStringNoVarExpansionMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansionCallCount() int {
	fake.validateEscapedStringNoVarExpansionMutex.RLock()
	defer fake.validateEscapedStringNoVarExpansionMutex.RUnlock()
	return len(fake.validateEscapedStringNoVarExpansionArgsForCall)
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansionCalls(stub func(string) error) {
	fake.validateEscapedStringNoVarExpansionMutex.Lock()
	defer fake.validateEscapedStringNoVarExpansionMutex.Unlock()
	fake.ValidateEscapedStringNoVarExpansionStub = stub
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansionArgsForCall(i int) string {
	fake.validateEscapedStringNoVarExpansionMutex.RLock()
	defer fake.validateEscapedStringNoVarExpansionMutex.RUnlock()
	argsForCall := fake.validateEscapedStringNoVarExpansionArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansionReturns(result1 error) {
	fake.validateEscapedStringNoVarExpansionMutex.Lock()
	defer fake.validateEscapedStringNoVarExpansionMutex.Unlock()
	fake.ValidateEscapedStringNoVarExpansionStub = nil
	fake.validateEscapedStringNoVarExpansionReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeGenericValidator) ValidateEscapedStringNoVarExpansionReturnsOnCall(i int, result1 error) {
	fake.validateEscapedStringNoVarExpansionMutex.Lock()
	defer fake.validateEscapedStringNoVarExpansionMutex.Unlock()
	fake.ValidateEscapedStringNoVarExpansionStub = nil
	if fake.validateEscapedStringNoVarExpansionReturnsOnCall == nil {
		fake.validateEscapedStringNoVarExpansionReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.validateEscapedStringNoVarExpansionReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeGenericValidator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.validateEscapedStringNoVarExpansionMutex.RLock()
	defer fake.validateEscapedStringNoVarExpansionMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGenericValidator) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ validation.GenericValidator = new(FakeGenericValidator)
