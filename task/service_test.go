package task

import (
	"context"
	"github.com/Av1shay/timers-scheduler-demo/ent"
	"github.com/Av1shay/timers-scheduler-demo/ent/task"
	"github.com/Av1shay/timers-scheduler-demo/log"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

type mockQueue struct {
	publishedTasks []*Task
}

func (q *mockQueue) Publish(_ context.Context, task *Task) error {
	q.publishedTasks = append(q.publishedTasks, task)
	return nil
}

func TestService_ProcessTasks(t *testing.T) {
	ctx := context.Background()

	dbClient, err := ent.Open("mysql", "user:password@tcp(localhost:3320)/task_scheduler?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer dbClient.Close()

	defer clearDb(ctx, dbClient)

	err = dbClient.Schema.Create(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dueDate := time.Now().UTC().Truncate(time.Second)
	taskNames := []string{"task1", "task2", "task3"}
	tasks := make([]*ent.Task, 3)
	taskIds := make([]int, 3)
	for i, name := range taskNames {
		ta, err := dbClient.Task.Create().SetWebhookUrl("https://" + name + ".com").SetDueDate(dueDate).Save(ctx)
		if err != nil {
			t.Fatal(err)
		}
		tasks[i] = ta
		taskIds[i] = ta.ID
	}

	q := &mockQueue{publishedTasks: make([]*Task, 0, 3)}
	service := NewService(dbClient, q, nil)

	// process tasks
	_ = service.processTasks(ctx, tasks)

	// check the queue
	if len(q.publishedTasks) != 3 {
		t.Fatalf("expected to have 3 published tasks, got %d", len(q.publishedTasks))
	}
	for _, publishedTask := range q.publishedTasks {
		var dbTask *ent.Task
		for _, dt := range tasks {
			if dt.ID == publishedTask.ID {
				dbTask = dt
				break
			}
		}
		if dbTask == nil {
			t.Errorf("task not %d not found in DB", publishedTask.ID)
			continue
		}
		if publishedTask.ID != dbTask.ID || publishedTask.WebhookURL != dbTask.WebhookUrl {
			t.Errorf("expected the two objects to have the same ID and WebhookURL: %v, %v", publishedTask, dbTask)
		}
	}

	// check that tasks status changed
	updatedTasks, err := dbClient.Task.Query().Where(task.IDIn(taskIds...)).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, ta := range updatedTasks {
		if ta.Status != task.StatusRunning {
			t.Errorf("expected task %d to have status running, got %s", ta.ID, ta.Status)
		}
	}
}

func TestService_ProcessOldTasks(t *testing.T) {
	ctx := context.Background()

	dbClient, err := ent.Open("mysql", "user:password@tcp(localhost:3320)/task_scheduler?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer dbClient.Close()

	defer clearDb(ctx, dbClient)

	err = dbClient.Schema.Create(ctx)
	if err != nil {
		t.Fatal(err)
	}

	q := &mockQueue{publishedTasks: make([]*Task, 0, 2)}
	service := NewService(dbClient, q, nil)

	// create some old tasks
	task1, err := dbClient.Task.Create().SetWebhookUrl("https://old-task1.com").SetDueDate(time.Now().Add(-10 * time.Minute)).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	task2, err := dbClient.Task.Create().SetWebhookUrl("https://old-task2.com").SetDueDate(time.Now().Add(-5 * time.Second)).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbClient.Task.Create().SetWebhookUrl("https://new-task3.com").SetDueDate(time.Now().Add(10 * time.Second)).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = service.ProcessOldTasks(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// check tasks status and that they've been added to queue
	if len(q.publishedTasks) != 2 {
		t.Fatalf("expected to have 2 published tasks, got %d", len(q.publishedTasks))
	}

	task1InDB, err := dbClient.Task.Get(ctx, task1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if task1InDB.Status != task.StatusRunning {
		t.Errorf("expected task %d to have status running, got %s", task1InDB.ID, task1InDB.Status)
	}
	task2InDB, err := dbClient.Task.Get(ctx, task2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if task2InDB.Status != task.StatusRunning {
		t.Errorf("expected task %d to have status running, got %s", task2InDB.ID, task2InDB.Status)

	}
}

func clearDb(ctx context.Context, dbClient *ent.Client) {
	if _, err := dbClient.TaskHistory.Delete().Exec(ctx); err != nil {
		log.Error(ctx, "failed to delete TaskHistory data")
	}
	if _, err := dbClient.Task.Delete().Exec(ctx); err != nil {
		log.Error(ctx, "failed to delete Task data")
	}
}
