// Code generated by counterfeiter. DO NOT EDIT.
package eventsfakes

import (
	"context"
	"sync"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
)

type FakeFirstEventBatchPreparer struct {
	PrepareStub        func(context.Context) (events.EventBatch, error)
	prepareMutex       sync.RWMutex
	prepareArgsForCall []struct {
		arg1 context.Context
	}
	prepareReturns struct {
		result1 events.EventBatch
		result2 error
	}
	prepareReturnsOnCall map[int]struct {
		result1 events.EventBatch
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeFirstEventBatchPreparer) Prepare(arg1 context.Context) (events.EventBatch, error) {
	fake.prepareMutex.Lock()
	ret, specificReturn := fake.prepareReturnsOnCall[len(fake.prepareArgsForCall)]
	fake.prepareArgsForCall = append(fake.prepareArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.PrepareStub
	fakeReturns := fake.prepareReturns
	fake.recordInvocation("Prepare", []interface{}{arg1})
	fake.prepareMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeFirstEventBatchPreparer) PrepareCallCount() int {
	fake.prepareMutex.RLock()
	defer fake.prepareMutex.RUnlock()
	return len(fake.prepareArgsForCall)
}

func (fake *FakeFirstEventBatchPreparer) PrepareCalls(stub func(context.Context) (events.EventBatch, error)) {
	fake.prepareMutex.Lock()
	defer fake.prepareMutex.Unlock()
	fake.PrepareStub = stub
}

func (fake *FakeFirstEventBatchPreparer) PrepareArgsForCall(i int) context.Context {
	fake.prepareMutex.RLock()
	defer fake.prepareMutex.RUnlock()
	argsForCall := fake.prepareArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeFirstEventBatchPreparer) PrepareReturns(result1 events.EventBatch, result2 error) {
	fake.prepareMutex.Lock()
	defer fake.prepareMutex.Unlock()
	fake.PrepareStub = nil
	fake.prepareReturns = struct {
		result1 events.EventBatch
		result2 error
	}{result1, result2}
}

func (fake *FakeFirstEventBatchPreparer) PrepareReturnsOnCall(i int, result1 events.EventBatch, result2 error) {
	fake.prepareMutex.Lock()
	defer fake.prepareMutex.Unlock()
	fake.PrepareStub = nil
	if fake.prepareReturnsOnCall == nil {
		fake.prepareReturnsOnCall = make(map[int]struct {
			result1 events.EventBatch
			result2 error
		})
	}
	fake.prepareReturnsOnCall[i] = struct {
		result1 events.EventBatch
		result2 error
	}{result1, result2}
}

func (fake *FakeFirstEventBatchPreparer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.prepareMutex.RLock()
	defer fake.prepareMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeFirstEventBatchPreparer) recordInvocation(key string, args []interface{}) {
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

var _ events.FirstEventBatchPreparer = new(FakeFirstEventBatchPreparer)
