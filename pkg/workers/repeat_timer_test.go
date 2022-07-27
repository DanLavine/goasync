package workers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/pkg/workers"
)

func TestRepeatTimer_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	repeatTimer := workers.RepeatTimer(time.Second, fakeWorker1)

	g.Expect(repeatTimer.Initialize()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.InitializeCallCount()).To(Equal(1))
}

func TestRepeatTimer_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	repeatTimer := workers.RepeatTimer(time.Second, fakeWorker1)

	g.Expect(repeatTimer.Cleanup()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.CleanupCallCount()).To(Equal(1))
}

func TestRepetTimer_Work_Swallows_Errors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.WorkReturns(fmt.Errorf("error that should be swallowed"))
	repeatTimer := workers.RepeatTimer(time.Nanosecond, fakeWorker1)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	go func() {
		errChan <- repeatTimer.Work(ctx)
	}()

	// check that we keep restarting the process
	g.Eventually(fakeWorker1.WorkCallCount).Should(BeNumerically(">=", 5))

	// cancel the background process
	cancel()
	g.Eventually(errChan).Should(Receive(BeNil()))
}
