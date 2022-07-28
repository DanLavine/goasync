package workers_test

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/pkg/workers"
)

// NOTE these tests all use SIGINT since that should work on Windows as well

func TestInterruptible_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	g.Expect(interruptible.Initialize()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.InitializeCallCount()).To(Equal(1))
}

func TestInterruptible_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	g.Expect(interruptible.Cleanup()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.CleanupCallCount()).To(Equal(1))
}

func TestInterruptible_Work_ExitsIfContextIsCanceled(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Work(ctx)
	}()

	g.Eventually(errChan).Should(Receive(BeNil()))
}

func TestInterruptible_Work_RestartsWorkerOnAProperIntercept(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.WorkStub = func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("returning from worker")
		return nil
	}
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	ctx, cancel := context.WithCancel(context.Background())
	//ctx := context.Background()
	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Work(ctx)
	}()

	g.Eventually(fakeWorker1.WorkCallCount).Should(Equal(1))
	g.Consistently(fakeWorker1.WorkCallCount).Should(Equal(1))
	g.Consistently(fakeWorker1.CleanupCallCount).Should(Equal(0))
	g.Consistently(fakeWorker1.InitializeCallCount).Should(Equal(0))

	// send interupt signal to the running application, triggering a reload in the worker
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	g.Eventually(fakeWorker1.WorkCallCount).Should(Equal(2))
	g.Consistently(fakeWorker1.WorkCallCount).Should(Equal(2))
	g.Consistently(fakeWorker1.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeWorker1.InitializeCallCount).Should(Equal(1))

	cancel()

	g.Eventually(errChan).Should(Receive(BeNil()))
}

func TestInterruptible_Work_ReturnsIfChildWorkerStops(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Work(context.Background())
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("Interruptible worker unexpectedly stopped"))))
}

func TestInterruptible_Work_ReturnsIfChildWorkerErrors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.WorkReturns(fmt.Errorf("worker error"))
	interruptible := workers.Interruptible([]os.Signal{syscall.SIGINT}, fakeWorker1)

	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Work(context.Background())
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("worker error"))))
}
