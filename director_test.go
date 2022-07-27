package goasync_test

import (
	"context"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/goasyncfakes"
	. "github.com/onsi/gomega"
)

func TestDrector_Run_CallsWorkersStart_InOrderTheyWereAdded(t *testing.T) {
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
