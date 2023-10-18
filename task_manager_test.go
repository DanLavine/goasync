package goasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/goasyncfakes"
	"github.com/DanLavine/goasync/tasks"
	. "github.com/onsi/gomega"
)

func TestManager_AddTask_ReturnsAnErrorIfTheTaskIsNil(t *testing.T) {
	g := NewGomegaWithT(t)

	director := goasync.NewTaskManager(goasync.StrictConfig())

	err := director.AddTask("task1", nil)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("task cannot be nil"))
}

func TestManager_AddTask_ReturnsAnErrorIfAlreadyShutDown(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// don't care about any of these errors
	_ = director.Run(ctx)

	err := director.AddTask("task1", fakeTask1)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("Task Manager has already shutdown. Will not manage task"))
}

func TestManager_AddTask_ReturnsAnErrorIfAlreadyRunning(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1Start := make(chan struct{})
	fakeTask1.InitializeStub = func() error {
		fakeTask1Start <- struct{}{}

		<-fakeTask1Start
		return nil
	}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeTask1Start).Should(Receive())
	err := director.AddTask("task2", fakeTask2)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("Task Manager is already running. Will not manage task"))
	cancel()

	g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
	g.Expect(fakeTask2.InitializeCallCount()).To(Equal(0))
	close(fakeTask1Start)

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestManager_AddExecuteTask_ReturnsAnErrorIfTaskManagerHasAlreadyShutdown(t *testing.T) {
	g := NewGomegaWithT(t)

	director := goasync.NewTaskManager(goasync.StrictConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = director.Run(ctx) // don't care about any of these errors

	fakeTask1 := &goasyncfakes.FakeTask{}
	err := director.AddExecuteTask("task1", fakeTask1)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("Task Manager has already shutdown. Will not manage task"))
}

func TestManager_Run_ReturnsAnErrorIfTheManagerAlreadyFinished(t *testing.T) {
	g := NewGomegaWithT(t)

	director := goasync.NewTaskManager(goasync.StrictConfig())
	_ = director.Run(context.Background())

	errs := director.Run(context.Background())
	g.Expect(len(errs)).To(Equal(1))
	g.Expect(errs[0].Err.Error()).To(Equal("task manager has already shutdown. Will not run again"))
}

func TestManager_Run_ReturnsAnErrorIfTheManagerIsAlreadyRunning(t *testing.T) {
	g := NewGomegaWithT(t)

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddExecuteTask("test", tasks.Repeatable(&goasyncfakes.FakeTask{}))

	ctx, cancel := context.WithCancel(context.Background())
	errsChan := make(chan []goasync.NamedError)
	go func() {
		errsChan <- director.Run(ctx)
	}()
	go func() {
		errsChan <- director.Run(ctx)
	}()

	g.Eventually(errsChan).Should(Receive(Equal([]goasync.NamedError{{Err: fmt.Errorf("task manager is already running. Will not run again")}})))
	cancel()
}

func TestManager_Run_Initializes_Tasks_InOrderTheyWereAdded(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1Start := make(chan struct{})
	fakeTask1.InitializeStub = func() error {
		<-fakeTask1Start
		return nil
	}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeTask1.InitializeCallCount).Should(Equal(1))
	g.Expect(fakeTask2.InitializeCallCount()).To(Equal(0))
	close(fakeTask1Start)
	g.Eventually(fakeTask2.InitializeCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(BeNil()))
}

func TestManager_Run_Initalize_Error_CallsCleanup_InReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask2Done := make(chan struct{})
	fakeTask3Done := make(chan struct{})

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask2 := &goasyncfakes.FakeTask{}
	fakeTask2.CleanupStub = func() error {
		<-fakeTask2Done
		return nil
	}
	fakeTask3 := &goasyncfakes.FakeTask{}
	fakeTask3.CleanupStub = func() error {
		<-fakeTask3Done
		return fmt.Errorf("failed to cleanup")
	}
	fakeTask3.InitializeReturns(fmt.Errorf("failed to initialize"))

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)
	director.AddTask("task3", fakeTask3)

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeTask3.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask2.CleanupCallCount).Should(Equal(0))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask3Done)

	g.Eventually(fakeTask2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask2Done)

	g.Eventually(fakeTask1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task3", Stage: goasync.Initialize, Err: fmt.Errorf("failed to initialize")}, {TaskName: "task3", Stage: goasync.Cleanup, Err: fmt.Errorf("failed to cleanup")}})))
}

func TestManager_Run_Cleanup_Tasks_RunInReverseOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.CleanupReturns(fmt.Errorf("failed cleanup"))

	fakeTask2 := &goasyncfakes.FakeTask{}
	fakeTask2Cleanup := make(chan struct{})
	fakeTask2.CleanupStub = func() error {
		<-fakeTask2Cleanup
		return nil
	}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", fakeTask2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	g.Eventually(fakeTask2.CleanupCallCount).Should(Equal(1))
	g.Consistently(fakeTask1.CleanupCallCount).Should(Equal(0))
	close(fakeTask2Cleanup)
	g.Eventually(fakeTask1.CleanupCallCount).Should(Equal(1))

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Cleanup, Err: fmt.Errorf("failed cleanup")}})))
}

func TestManger_Configurations(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("Context when AllowNoManagedProcesses = true", func(t *testing.T) {
		t.Run("Context when AllowExecuteFailures = false", func(t *testing.T) {
			t.Run("It stops processing if a task fails", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}

				cfg := goasync.Config{
					AllowNoManagedProcesses: true,
					ReportErrors:            true,
					AllowExecuteFailures:    false,
				}

				director := goasync.NewTaskManager(cfg)
				director.AddTask("task1", fakeTask1)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
			})
		})

		t.Run("Context when AllowExecuteFailures = true", func(t *testing.T) {
			t.Run("It only stops processing when the context is canceled", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}

				cfg := goasync.Config{
					AllowNoManagedProcesses: true,
					ReportErrors:            true,
					AllowExecuteFailures:    true,
				}

				director := goasync.NewTaskManager(cfg)
				director.AddTask("task1", fakeTask1)

				ctx, cancel := context.WithCancel(context.Background())
				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(ctx)
				}()

				g.Consistently(errs).ShouldNot(Receive())

				cancel()
				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))

			})
		})
	})

	t.Run("Context when AllowNoManagedProcesses = false", func(t *testing.T) {
		t.Run("It imediately returns if there are no tasks to manage", func(t *testing.T) {
			cfg := goasync.Config{
				AllowNoManagedProcesses: false,
				ReportErrors:            true,
				AllowExecuteFailures:    false,
			}

			director := goasync.NewTaskManager(cfg)
			g.Expect(director.Run(context.Background())).ToNot(BeNil())
		})

		t.Run("Context when AllowExecuteFailures = false", func(t *testing.T) {
			t.Run("It stops processing if a task fails", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask1.ExecuteStub = func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				}

				fakeTask2 := &goasyncfakes.FakeTask{}

				cfg := goasync.Config{
					AllowNoManagedProcesses: false,
					ReportErrors:            true,
					AllowExecuteFailures:    false,
				}

				director := goasync.NewTaskManager(cfg)
				director.AddTask("task1", fakeTask1)
				director.AddTask("task2", fakeTask2)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task2", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
			})
		})

		t.Run("Context when AllowExecuteFailures = true", func(t *testing.T) {
			t.Run("It allows tasks to fail and only stops processing when the context is canceled", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask1.ExecuteStub = func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				}

				fakeTask2 := &goasyncfakes.FakeTask{}

				cfg := goasync.Config{
					AllowNoManagedProcesses: false,
					ReportErrors:            true,
					AllowExecuteFailures:    true,
				}

				director := goasync.NewTaskManager(cfg)
				director.AddTask("task1", fakeTask1)
				director.AddTask("task2", fakeTask2)

				ctx, cancel := context.WithCancel(context.Background())
				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(ctx)
				}()

				g.Consistently(errs).ShouldNot(Receive())

				cancel()
				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task2", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
			})
		})
	})
}

func TestManager_Run_Tasks_FailureWillCancelOtherTasks(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1.ExecuteReturns(fmt.Errorf("failed work"))
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", tasks.Repeatable(fakeTask2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("failed work")}})))
}

func TestManager_Run_Tasks_UnexpectedReturnWillCancelOtherTasks(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)
	director.AddTask("task2", tasks.Repeatable(fakeTask2))

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
}

func TestManager_AddExecuteTask_CanBeAddedBeforeStaringTheTaskManager(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1Start := make(chan struct{})
	fakeTask1.ExecuteStub = func(ctx context.Context) error {
		fakeTask1Start <- struct{}{}
		return nil
	}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	err := director.AddExecuteTask("task1", fakeTask1)
	g.Expect(err).ToNot(HaveOccurred())

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeTask1Start).Should(Receive())
	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
}

func TestManager_AddExecuteTask_CanFailExecuteTaskmanagerCorrectly(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask1 := &goasyncfakes.FakeTask{}
	fakeTask1Start := make(chan struct{})
	fakeTask1.ExecuteStub = func(ctx context.Context) error {
		fakeTask1Start <- struct{}{}

		<-ctx.Done()
		return nil
	}
	fakeTask2 := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", fakeTask1)

	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(context.Background())
	}()

	g.Eventually(fakeTask1Start).Should(Receive())
	err := director.AddExecuteTask("task2", fakeTask2)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(fakeTask1.ExecuteCallCount()).To(Equal(1))
	g.Eventually(fakeTask2.ExecuteCallCount).Should(Equal(1))
	close(fakeTask1Start)

	g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task2", Stage: goasync.Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}})))
}

func TestManager_Run_ShutsDownAnyNumberOfProperlyRunningTasks(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager(goasync.StrictConfig())
	director.AddTask("task1", tasks.Repeatable(fakeTask))
	director.AddTask("task2", tasks.Repeatable(fakeTask))
	director.AddExecuteTask("task3", tasks.Repeatable(fakeTask))
	director.AddExecuteTask("task4", tasks.Repeatable(fakeTask))
	director.AddTask("task5", tasks.Repeatable(fakeTask))
	director.AddTask("task6", tasks.Repeatable(fakeTask))
	director.AddExecuteTask("task7", tasks.Repeatable(fakeTask))

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	director.AddExecuteTask("task7", tasks.Repeatable(fakeTask))
	director.AddExecuteTask("task8", tasks.Repeatable(fakeTask))

	go func() {
		director.AddExecuteTask("task9", tasks.Repeatable(fakeTask))
		director.AddExecuteTask("task10", tasks.Repeatable(fakeTask))
		director.AddExecuteTask("task11", tasks.Repeatable(fakeTask))
		director.AddExecuteTask("task12", tasks.Repeatable(fakeTask))
	}()

	cancel()

	g.Eventually(errs).Should(Receive(BeNil()))
}
