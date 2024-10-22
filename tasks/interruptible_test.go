package tasks_test

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/v2/goasyncfakes"
	"github.com/DanLavine/goasync/v2/tasks"
)

// NOTE these tests all use SIGINT since that should work on Windows as well

func TestInterruptible_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	g.Expect(interruptible.Initialize(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
}

func TestInterruptible_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	g.Expect(interruptible.Cleanup(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.CleanupCallCount()).To(Equal(1))
}

func TestInterruptible_Execute_ExitsIfContextIsCanceled(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Execute(ctx)
	}()

	g.Eventually(errChan).Should(Receive(BeNil()))
}

func TestInterruptible_Execute_RestartsTaskOnAProperIntercept(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteStub = func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("returning from task")
		return nil
	}
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	//ctx := context.Background()
	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Execute(ctx)
	}()

	g.Eventually(fakeTask1.ExecuteCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.ExecuteCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	g.Consistently(fakeTask1.InitializeCallCount).Should(Equal(0))

	// send interupt signal to the running application, triggering a reload in the task
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	g.Eventually(fakeTask1.ExecuteCallCount).Should(Equal(2))
	g.Consistently(fakeTask1.ExecuteCallCount).Should(Equal(2))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.InitializeCallCount).Should(Equal(1))

	cancel()

	g.Eventually(errChan).Should(Receive(BeNil()))
}

func TestInterruptible_Execute_ReturnsIfChildTaskStops(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Execute(context.Background())
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("Interruptible task unexpectedly stopped"))))
}

func TestInterruptible_Execute_ReturnsIfChildTaskErrors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("task error"))
	interruptible := tasks.Interruptible([]os.Signal{syscall.SIGINT}, fakeTask1)

	errChan := make(chan error)
	go func() {
		errChan <- interruptible.Execute(context.Background())
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("task error"))))
}
