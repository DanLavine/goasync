package goasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/tasks"
	. "github.com/onsi/gomega"
)

func TestDrector_Run_Initializes_Tasks_InOrderTheyWereAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1Start := make(chan struct{})
	fakeTask1.InitializeStub = func() error {
		<-fakeTask1Start
		return nil
	}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager()
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeTask1.InitializeCallCount).Should(Equal(1))
	g.Expect(fakeTask2.InitializeCallCount()).To(Equal(0))
	close(fakeTask1Start)
	g.Eventually(fakeTask2.InitializeCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestDrector_Run_Initalize_Error_CallsCleanup_InReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask2Done := make(chan struct{})
	fakeTask3Done := make(chan struct{})

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask2 := &goasyncfakes.FakeTask{}
	fakeTask2.CleanupStub = func() error {
		<-fakeTask2Done
		return nil
	}
	fakeTask3 := &goasyncfakes.FakeTask{}
	fakeTask3.CleanupStub = func() error {
		<-fakeTask3Done
		return fmt.Errorf("failed to cleanup")
	}
	fakeTask3.InitializeReturns(fmt.Errorf("failed to initialize"))

	director := goasync.NewTaskManager()
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)
	director.AddTask("task3", fakeTask3)

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeTask3.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask2.CleanupCallCount).Should(Equal(0))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask3Done)

	g.Eventually(fakeTask2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask2Done)

	g.Eventually(fakeTask1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task3", Stage: goasync.Initialize, Err: fmt.Errorf("failed to initialize")}, {TaskName: "task3", Stage: goasync.Cleanup, Err: fmt.Errorf("failed to cleanup")}})))
}

func TestDrector_Run_Tasks_CancelsOnTheContext(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager()
	director.AddTask("task1", tasks.Repeatable(fakeTask1))
	director.AddTask("task2", tasks.Repeatable(fakeTask2))

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Consistently(errs).ShouldNot(Receive())

	cancel()

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestDrector_Run_Tasks_FailureWillCancelOtherTasks(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("failed work"))
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager()
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", tasks.Repeatable(fakeTask2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("failed work")}})))
}

func TestDrector_Run_Tasks_UnexpectedReturnWillCancelOtherTasks(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager()
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", tasks.Repeatable(fakeTask2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
}

func TestDrector_Run_Cleanup_Tasks_RunInReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.CleanupReturns(fmt.Errorf("failed cleanup"))

	fakeTask2 := &goasyncfakes.FakeTask{}
	fakeTask2Cleanup := make(chan struct{})
	fakeTask2.CleanupStub = func() error {
		<-fakeTask2Cleanup
		return nil
	}

	director := goasync.NewTaskManager()
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeTask2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask2Cleanup)
	g.Eventually(fakeTask1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Cleanup, Err: fmt.Errorf("failed cleanup")}})))
}
