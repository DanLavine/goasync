package goasync_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DanLavine/goasync/v2"
	"github.com/DanLavine/goasync/v2/goasyncfakes"
	"github.com/DanLavine/goasync/v2/tasks"
	. "github.com/onsi/gomega"
)

func pointerOf[T any](value T) *T {
	return &value
}

func TestManager_AddTask(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the task in nil", func(t *testing.T) {
		director := goasync.NewTaskManager()

		err := director.AddTask("task1", nil, goasync.TASK_TYPE_STRICT)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("task cannot be nil"))
	})

	t.Run("It returns an error if the task type is invalid", func(t *testing.T) {
		director := goasync.NewTaskManager()

		err := director.AddTask("", &goasyncfakes.FakeTask{}, goasync.TASK_TYPE(32))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("taskType must be one of [TASK_TYPE_ERROR | TASK_TYPE_STRICT | TASK_TYPE_STOP_GROUP]"))
	})

	t.Run("It returns an error if the taskmanger is already running", func(t *testing.T) {
		fakeTask1 := &goasyncfakes.FakeTask{}
		fakeTask1Start := make(chan struct{})
		fakeTask1.InitializeStub = func(_ context.Context) error {
			fakeTask1Start <- struct{}{}

			<-fakeTask1Start
			return nil
		}
		fakeTask2 := &goasyncfakes.FakeTask{}

		director := goasync.NewTaskManager()
		director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)

		ctx, cancel := context.WithCancel(context.Background())
		errs := make(chan []goasync.NamedError)
		go func() {
			errs <- director.Run(ctx)
		}()

		g.Eventually(fakeTask1Start).Should(Receive())
		err := director.AddTask("task2", fakeTask2, goasync.TASK_TYPE_STRICT)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("task Manager is already running. Will not manage task"))
		cancel()

		g.Expect(fakeTask1.InitializeCallCount()).To(Equal(1))
		g.Expect(fakeTask2.InitializeCallCount()).To(Equal(0))
		close(fakeTask1Start)

		g.Eventually(errs).Should(Receive(BeNil()))
	})

	t.Run("It reurns an error if the taskmanager is already shut down", func(t *testing.T) {
		fakeTask1 := &goasyncfakes.FakeTask{}

		director := goasync.NewTaskManager()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// don't care about any of these errors
		_ = director.Run(ctx)

		err := director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("task Manager has already shutdown. Will not manage task"))
	})
}

func TestManager_AddExecuteTask(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It can be added before the TaskManager start processing", func(t *testing.T) {
		fakeTask1 := &goasyncfakes.FakeTask{}
		fakeTask1Start := make(chan struct{})
		fakeTask1.ExecuteStub = func(ctx context.Context) error {
			fakeTask1Start <- struct{}{}
			return nil
		}

		director := goasync.NewTaskManager()
		err := director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_STRICT)
		g.Expect(err).ToNot(HaveOccurred())

		errs := make(chan []goasync.NamedError)
		go func() {
			errs <- director.Run(context.Background())
		}()

		g.Eventually(fakeTask1Start).Should(Receive())
		g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("unexpected shutdown for task process"), ExecuteTaskType: pointerOf(goasync.EXECUTE_TASK_TYPE_STRICT)}})))
	})

	t.Run("It returns an error if the task type is invalid", func(t *testing.T) {
		director := goasync.NewTaskManager()

		err := director.AddExecuteTask("", &goasyncfakes.FakeTask{}, goasync.EXECUTE_TASK_TYPE(32))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("taskType must be one of [EXECUTE_TASK_TYPE_STRICT | EXECUTE_TASK_TYPE_ERROR]"))
	})

	t.Run("It returns an error if the task in nil", func(t *testing.T) {
		director := goasync.NewTaskManager()

		err := director.AddExecuteTask("task1", nil, goasync.EXECUTE_TASK_TYPE_ERROR)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("task cannot be nil"))
	})

	t.Run("It returns an error if TaskManager has already shutdown", func(t *testing.T) {
		director := goasync.NewTaskManager()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_ = director.Run(ctx) // don't care about any of these errors

		fakeTask1 := &goasyncfakes.FakeTask{}
		err := director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_STRICT)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(Equal("task Manager has already shutdown. Will not manage task"))
	})
}

func TestManager_Run(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It returns an error if the TaskManager is already shutdown", func(t *testing.T) {
		director := goasync.NewTaskManager()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_ = director.Run(ctx)

		errs := director.Run(ctx)
		g.Expect(len(errs)).To(Equal(1))
		g.Expect(errs[0].Err.Error()).To(Equal("task manager has already shutdown. Will not run again"))
	})

	t.Run("It returns an error if the TaskManager is already running", func(t *testing.T) {
		director := goasync.NewTaskManager()
		director.AddExecuteTask("test", tasks.Repeatable(&goasyncfakes.FakeTask{}), goasync.EXECUTE_TASK_TYPE_STRICT)

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
	})

	t.Run("Describe Initialize", func(t *testing.T) {
		t.Run("It runs tasks in the order they were added", func(t *testing.T) {
			fakeTask1 := &goasyncfakes.FakeTask{}
			fakeTask1Start := make(chan struct{})
			fakeTask1.InitializeStub = func(_ context.Context) error {
				<-fakeTask1Start
				return nil
			}
			fakeTask2 := &goasyncfakes.FakeTask{}

			director := goasync.NewTaskManager()
			director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
			director.AddTask("task2", fakeTask2, goasync.TASK_TYPE_STRICT)

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
		})

		t.Run("It calls Cleanup in reverse order if an Initialize returns an error", func(t *testing.T) {
			fakeTask2Done := make(chan struct{})
			fakeTask3Done := make(chan struct{})

			fakeTask1 := &goasyncfakes.FakeTask{}
			fakeTask2 := &goasyncfakes.FakeTask{}
			fakeTask2.CleanupStub = func(_ context.Context) error {
				<-fakeTask2Done
				return nil
			}
			fakeTask3 := &goasyncfakes.FakeTask{}
			fakeTask3.CleanupStub = func(_ context.Context) error {
				<-fakeTask3Done
				return fmt.Errorf("failed to cleanup")
			}
			fakeTask3.InitializeReturns(fmt.Errorf("failed to initialize"))

			director := goasync.NewTaskManager()
			director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
			director.AddTask("task2", fakeTask2, goasync.TASK_TYPE_STRICT)
			director.AddTask("task3", fakeTask3, goasync.TASK_TYPE_STRICT)

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

		})
	})

	t.Run("Describe Cleanup", func(t *testing.T) {
		t.Run("It runs in revers order of the added tasks", func(t *testing.T) {
			fakeTask1 := &goasyncfakes.FakeTask{}
			fakeTask1.CleanupReturns(fmt.Errorf("failed cleanup"))

			fakeTask2 := &goasyncfakes.FakeTask{}
			fakeTask2Cleanup := make(chan struct{})
			fakeTask2.CleanupStub = func(_ context.Context) error {
				<-fakeTask2Cleanup
				return nil
			}

			director := goasync.NewTaskManager()
			director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
			director.AddTask("task2", fakeTask2, goasync.TASK_TYPE_STRICT)

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
		})
	})

	t.Run("Describe Tasks", func(t *testing.T) {
		t.Run("Context TASK_TYPE_STRICT", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				failErr := fmt.Errorf("failed work")
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask1.ExecuteReturns(failErr)
				fakeTask2 := &goasyncfakes.FakeTask{}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
				director.AddTask("task2", tasks.Repeatable(fakeTask2), goasync.TASK_TYPE_STRICT)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: failErr, TaskType: pointerOf(goasync.TASK_TYPE_STRICT)}})))
			})

			t.Run("It cancels all other tasks on an unexpected return", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask2 := &goasyncfakes.FakeTask{}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STRICT)
				director.AddTask("task2", tasks.Repeatable(fakeTask2), goasync.TASK_TYPE_STRICT)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("unexpected shutdown for task process"), TaskType: pointerOf(goasync.TASK_TYPE_STRICT)}})))
			})
		})

		t.Run("Context TASK_TYPE_ERROR", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				failErr := fmt.Errorf("failed work")
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask1.ExecuteReturns(failErr)
				fakeTask2 := &goasyncfakes.FakeTask{}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_ERROR)
				director.AddTask("task2", tasks.Repeatable(fakeTask2), goasync.TASK_TYPE_ERROR)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: failErr, TaskType: pointerOf(goasync.TASK_TYPE_ERROR)}})))
			})

			t.Run("It does not cancel other tasks when returning", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask2 := &goasyncfakes.FakeTask{}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_ERROR)
				director.AddTask("task2", tasks.Repeatable(fakeTask2), goasync.TASK_TYPE_ERROR)

				cancelContext, cancel := context.WithCancel(context.Background())
				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(cancelContext)
				}()

				g.Consistently(errs).ShouldNot(Receive())
				cancel()

				g.Eventually(errs).Should(Receive(BeNil()))
			})
		})

		t.Run("Context TASK_TYPE_STOP_GROUP", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				failErr := fmt.Errorf("failed work")
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask1.ExecuteReturns(failErr)
				fakeTask2 := &goasyncfakes.FakeTask{}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STOP_GROUP)
				director.AddTask("task2", tasks.Repeatable(fakeTask2), goasync.TASK_TYPE_STOP_GROUP)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: failErr, TaskType: pointerOf(goasync.TASK_TYPE_STOP_GROUP)}})))
			})

			t.Run("It cancels all other tasks when each STOP_GROUP finishes", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeTask{}
				fakeTask2 := &goasyncfakes.FakeTask{}
				fakeTask3 := &goasyncfakes.FakeTask{}

				fakeTask3.ExecuteStub = func(ctx context.Context) error {
					select {
					case <-ctx.Done():
						panic("context was canceled on stop group")
					case <-time.After(time.Second):
					}

					return nil
				}

				director := goasync.NewTaskManager()
				director.AddTask("task1", fakeTask1, goasync.TASK_TYPE_STOP_GROUP)
				director.AddTask("task2", fakeTask2, goasync.TASK_TYPE_STOP_GROUP)
				director.AddTask("task3", fakeTask3, goasync.TASK_TYPE_STOP_GROUP)
				director.AddTask("task4", tasks.Repeatable(fakeTask1), goasync.TASK_TYPE_STRICT)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs, 2*time.Second).Should(Receive(BeNil()))
			})
		})
	})

	t.Run("Describe ExecuteTasks", func(t *testing.T) {
		t.Run("It can add a task to a running TaskMaanger", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				running := make(chan struct{})
				fakeTask1 := &goasyncfakes.FakeExecuteTask{}
				fakeTask1.ExecuteStub = func(ctx context.Context) error {
					running <- struct{}{}
					return nil
				}
				fakeTask2 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2.ExecuteStub = func(ctx context.Context) error {
					close(running)
					<-ctx.Done()
					return nil
				}

				director := goasync.NewTaskManager()
				g.Expect(director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_ERROR)).ToNot(HaveOccurred())

				ctx, cancel := context.WithCancel(context.Background())
				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(ctx)
				}()

				g.Eventually(running).Should(Receive(Equal(struct{}{})))
				g.Expect(director.AddExecuteTask("task2", fakeTask2, goasync.EXECUTE_TASK_TYPE_ERROR)).ToNot(HaveOccurred())

				cancel()
				g.Eventually(running).Should(BeClosed())
				g.Eventually(errs).Should(Receive(BeNil()))
			})
		})

		t.Run("Context EXECUTE_TASK_TYPE_STRICT", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				failErr := fmt.Errorf("failed work")
				fakeTask1 := &goasyncfakes.FakeExecuteTask{}
				fakeTask1.ExecuteReturns(failErr)
				fakeTask2 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2.ExecuteStub = func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				}

				director := goasync.NewTaskManager()
				director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_STRICT)
				director.AddExecuteTask("task2", fakeTask2, goasync.EXECUTE_TASK_TYPE_STRICT)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: failErr, ExecuteTaskType: pointerOf(goasync.EXECUTE_TASK_TYPE_STRICT)}})))
			})

			t.Run("It cancels all other tasks on an unexpected return", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2.ExecuteStub = func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				}

				director := goasync.NewTaskManager()
				director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_STRICT)
				director.AddExecuteTask("task2", fakeTask2, goasync.EXECUTE_TASK_TYPE_STRICT)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: fmt.Errorf("unexpected shutdown for task process"), ExecuteTaskType: pointerOf(goasync.EXECUTE_TASK_TYPE_STRICT)}})))
			})
		})

		t.Run("Context EXECUTE_TASK_TYPE_ERROR", func(t *testing.T) {
			t.Run("It cancels all other tasks on a failure", func(t *testing.T) {
				failErr := fmt.Errorf("failed work")
				fakeTask1 := &goasyncfakes.FakeExecuteTask{}
				fakeTask1.ExecuteReturns(failErr)
				fakeTask2 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2.ExecuteStub = func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				}

				director := goasync.NewTaskManager()
				director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_ERROR)
				director.AddExecuteTask("task2", fakeTask2, goasync.EXECUTE_TASK_TYPE_ERROR)

				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(context.Background())
				}()

				g.Eventually(errs).Should(Receive(Equal([]goasync.NamedError{{TaskName: "task1", Stage: goasync.Execute, Err: failErr, ExecuteTaskType: pointerOf(goasync.EXECUTE_TASK_TYPE_ERROR)}})))
			})

			t.Run("It does not cancel other tasks when returning", func(t *testing.T) {
				fakeTask1 := &goasyncfakes.FakeExecuteTask{}
				fakeTask2 := &goasyncfakes.FakeExecuteTask{}

				director := goasync.NewTaskManager()
				director.AddExecuteTask("task1", fakeTask1, goasync.EXECUTE_TASK_TYPE_ERROR)
				director.AddExecuteTask("task2", fakeTask2, goasync.EXECUTE_TASK_TYPE_ERROR)

				cancelContext, cancel := context.WithCancel(context.Background())
				errs := make(chan []goasync.NamedError)
				go func() {
					errs <- director.Run(cancelContext)
				}()

				g.Consistently(errs).ShouldNot(Receive())
				cancel()

				g.Eventually(errs).Should(Receive(BeNil()))
			})
		})
	})
}

func TestManager_parallel_tests(t *testing.T) {
	g := NewGomegaWithT(t)

	fakeTask := &goasyncfakes.FakeTask{}

	director := goasync.NewTaskManager()
	director.AddTask("task1", tasks.Repeatable(fakeTask), goasync.TASK_TYPE_STRICT)
	director.AddTask("task2", tasks.Repeatable(fakeTask), goasync.TASK_TYPE_STRICT)
	director.AddExecuteTask("task3", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
	director.AddExecuteTask("task4", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
	director.AddTask("task5", tasks.Repeatable(fakeTask), goasync.TASK_TYPE_STRICT)
	director.AddTask("task6", tasks.Repeatable(fakeTask), goasync.TASK_TYPE_STRICT)
	director.AddExecuteTask("task7", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan []goasync.NamedError)
	go func() {
		errs <- director.Run(ctx)
	}()

	director.AddExecuteTask("task7", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
	director.AddExecuteTask("task8", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)

	go func() {
		director.AddExecuteTask("task9", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
		director.AddExecuteTask("task10", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
		director.AddExecuteTask("task11", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
		director.AddExecuteTask("task12", tasks.Repeatable(fakeTask), goasync.EXECUTE_TASK_TYPE_STRICT)
	}()

	cancel()

	g.Eventually(errs).Should(Receive(BeNil()))
}
