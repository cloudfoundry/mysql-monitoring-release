// Code generated by counterfeiter. DO NOT EDIT.
package metricsfakes

import (
	"github.com/cloudfoundry-incubator/mysql-monitoring-release/src/mysql-metrics/metrics"
	"sync"
)

type FakeSender struct {
	SendValueStub        func(string, float64, string) error
	sendValueMutex       sync.RWMutex
	sendValueArgsForCall []struct {
		arg1 string
		arg2 float64
		arg3 string
	}
	sendValueReturns struct {
		result1 error
	}
	sendValueReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSender) SendValue(arg1 string, arg2 float64, arg3 string) error {
	fake.sendValueMutex.Lock()
	ret, specificReturn := fake.sendValueReturnsOnCall[len(fake.sendValueArgsForCall)]
	fake.sendValueArgsForCall = append(fake.sendValueArgsForCall, struct {
		arg1 string
		arg2 float64
		arg3 string
	}{arg1, arg2, arg3})
	fake.recordInvocation("SendValue", []interface{}{arg1, arg2, arg3})
	fake.sendValueMutex.Unlock()
	if fake.SendValueStub != nil {
		return fake.SendValueStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.sendValueReturns
	return fakeReturns.result1
}

func (fake *FakeSender) SendValueCallCount() int {
	fake.sendValueMutex.RLock()
	defer fake.sendValueMutex.RUnlock()
	return len(fake.sendValueArgsForCall)
}

func (fake *FakeSender) SendValueCalls(stub func(string, float64, string) error) {
	fake.sendValueMutex.Lock()
	defer fake.sendValueMutex.Unlock()
	fake.SendValueStub = stub
}

func (fake *FakeSender) SendValueArgsForCall(i int) (string, float64, string) {
	fake.sendValueMutex.RLock()
	defer fake.sendValueMutex.RUnlock()
	argsForCall := fake.sendValueArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeSender) SendValueReturns(result1 error) {
	fake.sendValueMutex.Lock()
	defer fake.sendValueMutex.Unlock()
	fake.SendValueStub = nil
	fake.sendValueReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSender) SendValueReturnsOnCall(i int, result1 error) {
	fake.sendValueMutex.Lock()
	defer fake.sendValueMutex.Unlock()
	fake.SendValueStub = nil
	if fake.sendValueReturnsOnCall == nil {
		fake.sendValueReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.sendValueReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSender) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.sendValueMutex.RLock()
	defer fake.sendValueMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSender) recordInvocation(key string, args []interface{}) {
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

var _ metrics.Sender = new(FakeSender)
