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

func TestRepeatTimer_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	repeatTimer := tasks.RepeatTimer(time.Second, fakeTask1)

	g.Expect(repeatTimer.Initialize()).ToNot(HaveOccurred())
	g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
}

func TestRepeatTimer_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	repeatTimer := tasks.RepeatTimer(time.Second, fakeTask1)

	g.Expect(repeatTimer.Cleanup()).ToNot(HaveOccurred())
	g.Expect(fakeTask1.CleanupCallCount()).To(Equal(1))
}

func TestRepetTimer_Execute_Swallows_Errors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("error that should be swallowed"))
	repeatTimer := tasks.RepeatTimer(time.Nanosecond, fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	go func() {
		errChan <- repeatTimer.Execute(ctx)
	}()

	// check that we keep restarting the process
	g.Eventually(fakeTask1.ExecuteCallCount).Should(BeNumerically(">=", 5))

	// cancel the background process
	cancel()
	g.Eventually(errChan).Should(Receive(BeNil()))
}
