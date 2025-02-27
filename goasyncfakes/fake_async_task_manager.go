// Code generated by counterfeiter. DO NOT EDIT.
package goasyncfakes

import (
	"context"
	"sync"

	goasync "github.com/DanLavine/goasync/v2"
)

type FakeAsyncTaskManager struct {
	AddExecuteTaskStub        func(string, goasync.ExecuteTask, goasync.EXECUTE_TASK_TYPE) error
	addExecuteTaskMutex       sync.RWMutex
	addExecuteTaskArgsForCall []struct {
		arg1 string
		arg2 goasync.ExecuteTask
		arg3 goasync.EXECUTE_TASK_TYPE
	}
	addExecuteTaskReturns struct {
		result1 error
	}
	addExecuteTaskReturnsOnCall map[int]struct {
		result1 error
	}
	AddTaskStub        func(string, goasync.Task, goasync.TASK_TYPE) error
	addTaskMutex       sync.RWMutex
	addTaskArgsForCall []struct {
		arg1 string
		arg2 goasync.Task
		arg3 goasync.TASK_TYPE
	}
	addTaskReturns struct {
		result1 error
	}
	addTaskReturnsOnCall map[int]struct {
		result1 error
	}
	RunStub        func(context.Context) []goasync.NamedError
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		arg1 context.Context
	}
	runReturns struct {
		result1 []goasync.NamedError
	}
	runReturnsOnCall map[int]struct {
		result1 []goasync.NamedError
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAsyncTaskManager) AddExecuteTask(arg1 string, arg2 goasync.ExecuteTask, arg3 goasync.EXECUTE_TASK_TYPE) error {
	fake.addExecuteTaskMutex.Lock()
	ret, specificReturn := fake.addExecuteTaskReturnsOnCall[len(fake.addExecuteTaskArgsForCall)]
	fake.addExecuteTaskArgsForCall = append(fake.addExecuteTaskArgsForCall, struct {
		arg1 string
		arg2 goasync.ExecuteTask
		arg3 goasync.EXECUTE_TASK_TYPE
	}{arg1, arg2, arg3})
	stub := fake.AddExecuteTaskStub
	fakeReturns := fake.addExecuteTaskReturns
	fake.recordInvocation("AddExecuteTask", []interface{}{arg1, arg2, arg3})
	fake.addExecuteTaskMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeAsyncTaskManager) AddExecuteTaskCallCount() int {
	fake.addExecuteTaskMutex.RLock()
	defer fake.addExecuteTaskMutex.RUnlock()
	return len(fake.addExecuteTaskArgsForCall)
}

func (fake *FakeAsyncTaskManager) AddExecuteTaskCalls(stub func(string, goasync.ExecuteTask, goasync.EXECUTE_TASK_TYPE) error) {
	fake.addExecuteTaskMutex.Lock()
	defer fake.addExecuteTaskMutex.Unlock()
	fake.AddExecuteTaskStub = stub
}

func (fake *FakeAsyncTaskManager) AddExecuteTaskArgsForCall(i int) (string, goasync.ExecuteTask, goasync.EXECUTE_TASK_TYPE) {
	fake.addExecuteTaskMutex.RLock()
	defer fake.addExecuteTaskMutex.RUnlock()
	argsForCall := fake.addExecuteTaskArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeAsyncTaskManager) AddExecuteTaskReturns(result1 error) {
	fake.addExecuteTaskMutex.Lock()
	defer fake.addExecuteTaskMutex.Unlock()
	fake.AddExecuteTaskStub = nil
	fake.addExecuteTaskReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAsyncTaskManager) AddExecuteTaskReturnsOnCall(i int, result1 error) {
	fake.addExecuteTaskMutex.Lock()
	defer fake.addExecuteTaskMutex.Unlock()
	fake.AddExecuteTaskStub = nil
	if fake.addExecuteTaskReturnsOnCall == nil {
		fake.addExecuteTaskReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addExecuteTaskReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeAsyncTaskManager) AddTask(arg1 string, arg2 goasync.Task, arg3 goasync.TASK_TYPE) error {
	fake.addTaskMutex.Lock()
	ret, specificReturn := fake.addTaskReturnsOnCall[len(fake.addTaskArgsForCall)]
	fake.addTaskArgsForCall = append(fake.addTaskArgsForCall, struct {
		arg1 string
		arg2 goasync.Task
		arg3 goasync.TASK_TYPE
	}{arg1, arg2, arg3})
	stub := fake.AddTaskStub
	fakeReturns := fake.addTaskReturns
	fake.recordInvocation("AddTask", []interface{}{arg1, arg2, arg3})
	fake.addTaskMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeAsyncTaskManager) AddTaskCallCount() int {
	fake.addTaskMutex.RLock()
	defer fake.addTaskMutex.RUnlock()
	return len(fake.addTaskArgsForCall)
}

func (fake *FakeAsyncTaskManager) AddTaskCalls(stub func(string, goasync.Task, goasync.TASK_TYPE) error) {
	fake.addTaskMutex.Lock()
	defer fake.addTaskMutex.Unlock()
	fake.AddTaskStub = stub
}

func (fake *FakeAsyncTaskManager) AddTaskArgsForCall(i int) (string, goasync.Task, goasync.TASK_TYPE) {
	fake.addTaskMutex.RLock()
	defer fake.addTaskMutex.RUnlock()
	argsForCall := fake.addTaskArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeAsyncTaskManager) AddTaskReturns(result1 error) {
	fake.addTaskMutex.Lock()
	defer fake.addTaskMutex.Unlock()
	fake.AddTaskStub = nil
	fake.addTaskReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeAsyncTaskManager) AddTaskReturnsOnCall(i int, result1 error) {
	fake.addTaskMutex.Lock()
	defer fake.addTaskMutex.Unlock()
	fake.AddTaskStub = nil
	if fake.addTaskReturnsOnCall == nil {
		fake.addTaskReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addTaskReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeAsyncTaskManager) Run(arg1 context.Context) []goasync.NamedError {
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

func (fake *FakeAsyncTaskManager) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *FakeAsyncTaskManager) RunCalls(stub func(context.Context) []goasync.NamedError) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = stub
}

func (fake *FakeAsyncTaskManager) RunArgsForCall(i int) context.Context {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	argsForCall := fake.runArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeAsyncTaskManager) RunReturns(result1 []goasync.NamedError) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 []goasync.NamedError
	}{result1}
}

func (fake *FakeAsyncTaskManager) RunReturnsOnCall(i int, result1 []goasync.NamedError) {
	fake.runMutex.Lock()
	defer fake.runMutex.Unlock()
	fake.RunStub = nil
	if fake.runReturnsOnCall == nil {
		fake.runReturnsOnCall = make(map[int]struct {
			result1 []goasync.NamedError
		})
	}
	fake.runReturnsOnCall[i] = struct {
		result1 []goasync.NamedError
	}{result1}
}

func (fake *FakeAsyncTaskManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addExecuteTaskMutex.RLock()
	defer fake.addExecuteTaskMutex.RUnlock()
	fake.addTaskMutex.RLock()
	defer fake.addTaskMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeAsyncTaskManager) recordInvocation(key string, args []interface{}) {
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

var _ goasync.AsyncTaskManager = new(FakeAsyncTaskManager)
