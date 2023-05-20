package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Av1shay/timers-scheduler-demo/ent"
	task2 "github.com/Av1shay/timers-scheduler-demo/ent/task"
	"github.com/Av1shay/timers-scheduler-demo/task"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	dbClient *ent.Client
	ts       *httptest.Server
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	dbClient, err = ent.Open("mysql", "user:password@tcp(localhost:3320)/task_scheduler?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer dbClient.Close()
	err = dbClient.Schema.Create(ctx)
	if err != nil {
		log.Fatal(err)
	}
	taskService := task.NewService(dbClient, nil, nil)
	srv := New(taskService)

	router := mux.NewRouter()
	srv.MountHandlers(router)

	ts = httptest.NewServer(router)
	defer ts.Close()

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestServer_NewTimer(t *testing.T) {
	ctx := context.Background()
	taskURL := "https://walla.com"
	dummyRequest := SetTimerReq{
		Hours:   5,
		Minutes: 0,
		Seconds: 2,
		URL:     taskURL,
	}
	b, _ := json.Marshal(dummyRequest)
	res, err := http.Post(ts.URL+"/timers", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status code %d", res.StatusCode)
	}
	resp, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var respData SetTimerResp
	if err := json.Unmarshal(resp, &respData); err != nil {
		t.Fatal(err)
	}

	// check the task against the db
	taskEnt, err := dbClient.Task.Get(ctx, respData.ID)
	if err != nil {
		t.Fatal(err)
	}
	if taskEnt.WebhookUrl != taskURL {
		t.Errorf("want task URL to be %s, got %s", taskURL, taskEnt.WebhookUrl)
	}
	if taskEnt.Status != task2.StatusPending {
		t.Errorf("want task status to be %s, got %s", task2.StatusPending, taskEnt.Status)
	}

	// check validation
	dummyRequest.URL = "not-a-valid-url"
	b, _ = json.Marshal(dummyRequest)
	res, err = http.Post(ts.URL+"/timers", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 400 {
		t.Errorf("expxected status code 400, got %d", res.StatusCode)
	}

	dummyRequest.URL = "http://walla.com"
	dummyRequest.Seconds = -5
	b, _ = json.Marshal(dummyRequest)
	res, err = http.Post(ts.URL+"/timers", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 400 {
		t.Errorf("expxected status code 400, got %d", res.StatusCode)
	}
}

func TestServer_GetTimer(t *testing.T) {
	ctx := context.Background()

	n := time.Now()
	taskInFuture, err := dbClient.Task.Create().SetDueDate(n.Add(30 * time.Second)).SetWebhookUrl("https://example.com").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(fmt.Sprintf("%s/timers/%d", ts.URL, taskInFuture.ID))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status code %d", res.StatusCode)
	}
	resp, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var respData GetTimerResp
	if err := json.Unmarshal(resp, &respData); err != nil {
		t.Fatal(err)
	}
	if respData.TimeLeft < 28 || respData.TimeLeft > 32 {
		t.Errorf("expxected timeLeft to be in range [28,32], got %d", respData.TimeLeft)
	}

	// check task in past
	taskInPast, err := dbClient.Task.Create().SetDueDate(n.Add(-30 * time.Second)).SetWebhookUrl("https://example.com").Save(ctx)
	if err != nil {
		t.Fatal(err)
	}

	resX, err := http.Get(fmt.Sprintf("%s/timers/%d", ts.URL, taskInPast.ID))
	if err != nil {
		t.Fatal(err)
	}
	if resX.StatusCode != 200 {
		t.Fatalf("status code %d", resX.StatusCode)
	}
	respX, err := io.ReadAll(resX.Body)
	resX.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(respX, &respData); err != nil {
		t.Fatal(err)
	}
	if respData.TimeLeft != 0 {
		t.Errorf("expxected timeLeft to be 0, got %d", respData.TimeLeft)
	}
}
