package goasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/goasyncfakes"
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

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeWorker1.InitializeCallCount).Should(Equal(1))
	g.Expect(fakeWorker2.InitializeCallCount()).To(Equal(0))
	close(fakeWorker1Start)
	g.Eventually(fakeWorker2.InitializeCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(nil))
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
