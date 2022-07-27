package workers_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/pkg/workers"
)

func TestRepeatable_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	repeatable := workers.NewRepeatableWorker(fakeWorker1)

	g.Expect(repeatable.Initialize()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.InitializeCallCount()).To(Equal(1))
}

func TestRepeatable_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	repeatable := workers.NewRepeatableWorker(fakeWorker1)

	g.Expect(repeatable.Cleanup()).ToNot(HaveOccurred())
	g.Expect(fakeWorker1.CleanupCallCount()).To(Equal(1))
}

func TestRepetable_Work_Swallows_Errors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeWorker1 := &goasyncfakes.FakeWorker{}
	fakeWorker1.WorkReturns(fmt.Errorf("error that should be swallowed"))
	repeatable := workers.NewRepeatableWorker(fakeWorker1)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	go func() {
		errChan <- repeatable.Work(ctx)
	}()

	// check that we keep restarting the process
	g.Eventually(fakeWorker1.WorkCallCount).Should(BeNumerically(">=", 5))

	// cancel the background process
	cancel()
	g.Eventually(errChan).Should(Receive(BeNil()))
}
