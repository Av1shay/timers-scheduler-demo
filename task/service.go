package task

import (
	"context"
	"fmt"
	"github.com/Av1shay/timers-scheduler-demo/ent"
	"github.com/Av1shay/timers-scheduler-demo/ent/task"
	"github.com/Av1shay/timers-scheduler-demo/log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Service struct {
	dbClient   *ent.Client
	queue      Queue
	httpClient *http.Client
}

type Queue interface {
	Publish(ctx context.Context, task *Task) error
}

func NewService(dbClient *ent.Client, queue Queue, httpClient *http.Client) *Service {
	return &Service{
		dbClient,
		queue,
		httpClient,
	}
}

// ProcessCurrentTasks process task with dueDate in current second
func (s *Service) ProcessCurrentTasks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	dueDate := time.Now().UTC().Truncate(time.Second)
	tasks, err := s.dbClient.Task.
		Query().
		Where(task.DueDateEQ(dueDate), task.StatusEQ(task.StatusPending)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tasks for proccesing: %s", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	return s.processTasks(ctx, tasks)
}

// ProcessOldTasks finds and process any task with status PENDING that was not processed in time for some reason
func (s *Service) ProcessOldTasks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	tasks, err := s.dbClient.Task.
		Query().
		Where(task.DueDateLTE(time.Now().UTC()), task.StatusEQ(task.StatusPending)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to old tasks for proccesing: %s", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	return s.processTasks(ctx, tasks)
}

func (s *Service) processTasks(ctx context.Context, tasks []*ent.Task) error {
	workers := 5
	wg := sync.WaitGroup{}
	wg.Add(workers)
	tasksChan := make(chan *ent.Task, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for t := range tasksChan {
				log.Info(ctx, "processing task", t.ID)

				if err := s.processTask(ctx, t); err != nil {
					log.Errorf(ctx, "failed to process task %d: %v\n", t.ID, err)
				}
			}
		}()
	}

	for _, t := range tasks {
		tasksChan <- t
	}

	close(tasksChan)
	wg.Wait()

	return nil
}

// processTask insert task to queue, wrap the insertion and status update in a transaction
func (s *Service) processTask(ctx context.Context, t *ent.Task) error {
	tx, err := s.dbClient.Tx(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Task.UpdateOne(t).SetStatus(task.StatusRunning).Save(ctx)
	if err != nil {
		return rollback(tx, err)
	}
	if err := s.queue.Publish(ctx, parseTask(t)); err != nil {
		return rollback(tx, err)
	}
	return tx.Commit()
}

func (s *Service) SaveTask(ctx context.Context, dueDate time.Time, webhookURL string) (*Task, error) {
	taskEnt, err := s.dbClient.Task.Create().SetDueDate(dueDate.Truncate(time.Second)).SetWebhookUrl(webhookURL).Save(ctx)
	if err != nil {
		return nil, err
	}
	return parseTask(taskEnt), nil
}

func (s *Service) GetTask(ctx context.Context, id int) (*Task, error) {
	taskEnt, err := s.dbClient.Task.Get(ctx, id)
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			return nil, &ApiError{404, err.Error(), fmt.Sprintf("task with id %d not found", id)}
		}
		return nil, &ApiError{500, err.Error(), "something went wrong"}
	}
	return parseTask(taskEnt), nil
}

// EmitTask send POST request to tasks webhook and update DB
func (s *Service) EmitTask(ctx context.Context, t *Task) error {
	err := s.emitTask(ctx, t)
	if updateErr := s.updateTaskAfterRun(ctx, t, err); updateErr != nil {
		// we don't return error here because this is not a retriable error, we don't want to emit the task twice
		log.Errorf(ctx, "failed up update task %d after emitting error: %s\n", t.ID, updateErr)
	}
	return err
}

func (s *Service) emitTask(ctx context.Context, t *Task) error {
	url := fmt.Sprintf("%s/%d", strings.TrimSuffix(t.WebhookURL, "/"), t.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest { // TODO check if we should care about the response
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

// updateTaskAfterRun update task and add another entry to its history with optional error, to keep track on each run
func (s *Service) updateTaskAfterRun(ctx context.Context, t *Task, runErr error) error {
	tx, err := s.dbClient.Tx(ctx)
	if err != nil {
		return err
	}
	updatedTask, err := tx.Task.UpdateOneID(t.ID).SetStatus(task.StatusDone).Save(ctx)
	if err != nil {
		return rollback(tx, err)
	}
	taskHistoryCreator := tx.TaskHistory.Create().SetTask(updatedTask)
	if runErr != nil {
		taskHistoryCreator.SetError(runErr.Error())
	}
	_, err = taskHistoryCreator.Save(ctx)
	if err != nil {
		return rollback(tx, err)
	}
	return tx.Commit()
}

func parseTask(t *ent.Task) *Task {
	return &Task{
		ID:         t.ID,
		WebhookURL: t.WebhookUrl,
		DueDate:    t.DueDate,
	}
}

// rollback rolls back a transaction and combine original error with rollback error if occurred
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		err = fmt.Errorf("%w: %v", err, rerr)
	}
	return err
}
