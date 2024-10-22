package tasks_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/DanLavine/goasync/v2/goasyncfakes"
	"github.com/DanLavine/goasync/v2/tasks"
)

func TestRepeatable_Initialize_Calls_Callback_Initialize(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	repeatable := tasks.Repeatable(fakeTask1)

	g.Expect(repeatable.Initialize(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
}

func TestRepeatable_Cleanup_Calls_Callback_Cleanup(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	repeatable := tasks.Repeatable(fakeTask1)

	g.Expect(repeatable.Cleanup(context.Background())).ToNot(HaveOccurred())
	g.Expect(fakeTask1.CleanupCallCount()).To(Equal(1))
}

func TestRepetable_Execute_Swallows_Errors(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("error that should be swallowed"))
	repeatable := tasks.Repeatable(fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	go func() {
		errChan <- repeatable.Execute(ctx)
	}()

	// check that we keep restarting the process
	g.Eventually(fakeTask1.ExecuteCallCount).Should(BeNumerically(">=", 5))

	// cancel the background process
	cancel()
	g.Eventually(errChan).Should(Receive(BeNil()))
}
