package goasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/pkg/workers"
	. "github.com/onsi/gomega"
)

func TestDrector_Run_Initializes_Workers_InOrderTheyWereAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1Start := make(chan struct{})
	fakeWorker1.InitializeStub = func() error {
		<-fakeWorker1Start
		return nil
	}
	fakeWorker2 := &goasyncfakes.FakeWorker{}

	director := goasync.NewDirector()
	director.AddWorker("worker1", fakeWorker1)
	director.AddWorker("worker2", fakeWorker2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeWorker1.InitializeCallCount).Should(Equal(1))
	g.Expect(fakeWorker2.InitializeCallCount()).To(Equal(0))
	close(fakeWorker1Start)
	g.Eventually(fakeWorker2.InitializeCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestDrector_Run_Initalize_Error_CallsCleanup_InReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker2Done := make(chan struct{})
	fakeWorker3Done := make(chan struct{})

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker2 := &goasyncfakes.FakeWorker{}
	fakeWorker2.CleanupStub = func() error {
		<-fakeWorker2Done
		return nil
	}
	fakeWorker3 := &goasyncfakes.FakeWorker{}
	fakeWorker3.CleanupStub = func() error {
		<-fakeWorker3Done
		return fmt.Errorf("failed to cleanup")
	}
	fakeWorker3.InitializeReturns(fmt.Errorf("failed to initialize"))

	director := goasync.NewDirector()
	director.AddWorker("worker1", fakeWorker1)
	director.AddWorker("worker2", fakeWorker2)
	director.AddWorker("worker3", fakeWorker3)

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeWorker3.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeWorker2.CleanupCallCount).Should(Equal(0))
	g.Consistently(fakeWorker1.CleanupCallCount).Should(Equal(0))
	close(fakeWorker3Done)

	g.Eventually(fakeWorker2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeWorker1.CleanupCallCount).Should(Equal(0))
	close(fakeWorker2Done)

	g.Eventually(fakeWorker1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{WorkerName: "worker3", Stage: goasync.Initialize, Err: fmt.Errorf("failed to initialize")}, {WorkerName: "worker3", Stage: goasync.Cleanup, Err: fmt.Errorf("failed to cleanup")}})))
}

func TestDrector_Run_Workers_CancelsOnTheContext(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker2 := &goasyncfakes.FakeWorker{}

	director := goasync.NewDirector()
	director.AddWorker("worker1", workers.Repeatable(fakeWorker1))
	director.AddWorker("worker2", workers.Repeatable(fakeWorker2))

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Consistently(errs).ShouldNot(Receive())

	cancel()

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestDrector_Run_Workers_FailureWillCancelOtherWorkers(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.WorkReturns(fmt.Errorf("failed work"))
	fakeWorker2 := &goasyncfakes.FakeWorker{}

	director := goasync.NewDirector()
	director.AddWorker("worker1", fakeWorker1)
	director.AddWorker("worker2", workers.Repeatable(fakeWorker2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{WorkerName: "worker1", Stage: goasync.Work, Err: fmt.Errorf("failed work")}})))
}

func TestDrector_Run_Workers_UnexpectedReturnWillCancelOtherWorkers(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker2 := &goasyncfakes.FakeWorker{}

	director := goasync.NewDirector()
	director.AddWorker("worker1", fakeWorker1)
	director.AddWorker("worker2", workers.Repeatable(fakeWorker2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{WorkerName: "worker1", Stage: goasync.Work, Err: fmt.Errorf("unexpected shutdown for worker process")}})))
}

func TestDrector_Run_Cleanup_Workers_RunInReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.CleanupReturns(fmt.Errorf("failed cleanup"))

	fakeWorker2 := &goasyncfakes.FakeWorker{}
	fakeWorker2Cleanup := make(chan struct{})
	fakeWorker2.CleanupStub = func() error {
		<-fakeWorker2Cleanup
		return nil
	}

	director := goasync.NewDirector()
	director.AddWorker("worker1", fakeWorker1)
	director.AddWorker("worker2", fakeWorker2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeWorker2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeWorker1.CleanupCallCount).Should(Equal(0))
	close(fakeWorker2Cleanup)
	g.Eventually(fakeWorker1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{WorkerName: "worker1", Stage: goasync.Cleanup, Err: fmt.Errorf("failed cleanup")}})))
}
