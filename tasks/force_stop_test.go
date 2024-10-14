package tasks_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/tasks"
)

func TestForceStop_Initialize_Calls_Callback_Initialize(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	forceStop := tasks.ForceStop(time.Second, fakeTask1)

	g.Expect(forceStop.Initialize(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
}

func TestForceStop_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	forceStop := tasks.ForceStop(time.Second, fakeTask1)

	g.Expect(forceStop.Cleanup(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.CleanupCallCount()).To(Equal(1))
}

func TestForceStop_Execute_ExitsIfContextIsCanceled(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	forceStop := tasks.ForceStop(time.Second, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := make(chan error)
	go func() {
		errChan <- forceStop.Execute(ctx)
	}()

	g.Eventually(errChan).Should(Receive(BeNil()))
}

func TestForceStop_Execute_WillTimeOutIfSubprocessDoesNotExit(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteStub = func(ctx context.Context) error {
		<-ctx.Done()
		time.Sleep(5 * time.Second)
		return nil
	}
	forceStop := tasks.ForceStop(time.Nanosecond, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := make(chan error)
	go func() {
		errChan <- forceStop.Execute(ctx)
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("Failed to stop sub task in 1ns"))))
}

func TestForceStop_Execute_ReturnsIfChildTaskStops(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("Failed sub task"))
	forceStop := tasks.ForceStop(time.Second, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errChan := make(chan error)
	go func() {
		errChan <- forceStop.Execute(ctx)
	}()

	g.Eventually(errChan).Should(Receive(Equal(fmt.Errorf("Failed sub task"))))
}
