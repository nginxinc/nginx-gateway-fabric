// Code generated by counterfeiter. DO NOT EDIT.
package exchangerfakes

import (
	"context"
	"sync"

	"github.com/nginx/agent/sdk/v2/proto"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/grpc/commander/exchanger"
)

type FakeCommandExchanger struct {
	InStub        func() chan<- *proto.Command
	inMutex       sync.RWMutex
	inArgsForCall []struct {
	}
	inReturns struct {
		result1 chan<- *proto.Command
	}
	inReturnsOnCall map[int]struct {
		result1 chan<- *proto.Command
	}
	OutStub        func() <-chan *proto.Command
	outMutex       sync.RWMutex
	outArgsForCall []struct {
	}
	outReturns struct {
		result1 <-chan *proto.Command
	}
	outReturnsOnCall map[int]struct {
		result1 <-chan *proto.Command
	}
	RunStub        func(context.Context) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		arg1 context.Context
	}
	runReturns struct {
		result1 error
	}
	runReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCommandExchanger) In() chan<- *proto.Command {
	fake.inMutex.Lock()
	ret, specificReturn := fake.inReturnsOnCall[len(fake.inArgsForCall)]
	fake.inArgsForCall = append(fake.inArgsForCall, struct {
	}{})
	stub := fake.InStub
	fakeReturns := fake.inReturns
	fake.recordInvocation("In", []interface{}{})
	fake.inMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCommandExchanger) InCallCount() int {
	fake.inMutex.RLock()
	defer fake.inMutex.RUnlock()
	return len(fake.inArgsForCall)
}

func (fake *FakeCommandExchanger) InCalls(stub func() chan<- *proto.Command) {
	fake.inMutex.Lock()
	defer fake.inMutex.Unlock()
	fake.InStub = stub
}

func (fake *FakeCommandExchanger) InReturns(result1 chan<- *proto.Command) {
	fake.inMutex.Lock()
	defer fake.inMutex.Unlock()
	fake.InStub = nil
	fake.inReturns = struct {
		result1 chan<- *proto.Command
	}{result1}
}

func (fake *FakeCommandExchanger) InReturnsOnCall(i int, result1 chan<- *proto.Command) {
	fake.inMutex.Lock()
	defer fake.inMutex.Unlock()
	fake.InStub = nil
	if fake.inReturnsOnCall == nil {
		fake.inReturnsOnCall = make(map[int]struct {
			result1 chan<- *proto.Command
		})
	}
	fake.inReturnsOnCall[i] = struct {
		result1 chan<- *proto.Command
	}{result1}
}

func (fake *FakeCommandExchanger) Out() <-chan *proto.Command {
	fake.outMutex.Lock()
	ret, specificReturn := fake.outReturnsOnCall[len(fake.outArgsForCall)]
	fake.outArgsForCall = append(fake.outArgsForCall, struct {
	}{})
	stub := fake.OutStub
	fakeReturns := fake.outReturns
	fake.recordInvocation("Out", []interface{}{})
	fake.outMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCommandExchanger) OutCallCount() int {
	fake.outMutex.RLock()
	defer fake.outMutex.RUnlock()
	return len(fake.outArgsForCall)
}

func (fake *FakeCommandExchanger) OutCalls(stub func() <-chan *proto.Command) {
	fake.outMutex.Lock()
	defer fake.outMutex.Unlock()
	fake.OutStub = stub
}

func (fake *FakeCommandExchanger) OutReturns(result1 <-chan *proto.Command) {
	fake.outMutex.Lock()
	defer fake.outMutex.Unlock()
	fake.OutStub = nil
	fake.outReturns = struct {
		result1 <-chan *proto.Command
	}{result1}
}

func (fake *FakeCommandExchanger) OutReturnsOnCall(i int, result1 <-chan *proto.Command) {
	fake.outMutex.Lock()
	defer fake.outMutex.Unlock()
	fake.OutStub = nil
	if fake.outReturnsOnCall == nil {
		fake.outReturnsOnCall = make(map[int]struct {
			result1 <-chan *proto.Command
		})
	}
	fake.outReturnsOnCall[i] = struct {
		result1 <-chan *proto.Command
	}{result1}
}

func (fake *FakeCommandExchanger) Run(arg1 context.Context) error {
	fake.runMutex.Lock()
	ret, specificReturn := fake.runReturnsOnCall[len(fake.runArgsForCall)]
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.RunStub
	fakeReturns := fake.runReturns
	fake.recordInvocation("Run", []interface{}{arg1})
	fake.runMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeCommandExchanger) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeCommandExchanger) RunCalls(stub func(context.Context) error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = stub
}

func (fake *FakeCommandExchanger) RunArgsForCall(i int) context.Context {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	argsForCall := fake.runArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeCommandExchanger) RunReturns(result1 error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandExchanger) RunReturnsOnCall(i int, result1 error) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	if fake.runReturnsOnCall == nil {
		fake.runReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.runReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommandExchanger) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.inMutex.RLock()
	defer fake.inMutex.RUnlock()
	fake.outMutex.RLock()
	defer fake.outMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCommandExchanger) recordInvocation(key string, args []interface{}) {
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

var _ exchanger.CommandExchanger = new(FakeCommandExchanger)
