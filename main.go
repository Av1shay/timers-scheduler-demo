package main

import (
	"context"
	"encoding/json"
	"github.com/Av1shay/timers-scheduler-demo/ent"
	"github.com/Av1shay/timers-scheduler-demo/logx"
	"github.com/Av1shay/timers-scheduler-demo/rabbitmq_queue"
	"github.com/Av1shay/timers-scheduler-demo/server"
	"github.com/Av1shay/timers-scheduler-demo/task"
	"github.com/go-co-op/gocron"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultPort      = "8081"
	defaultQueueName = "tasks_queue"

	// defaults for demo purposes only
	defaultMysqlConn    = "user:password@tcp(localhost:3320)/task_scheduler?parseTime=true"
	defaultRabbitMqConn = "amqp://guest:guest@localhost:5672/"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(".env"); err != nil {
		log.Println("failed to load .env file, using default configuration. Error:", err)
	}

	port, found := os.LookupEnv("PORT")
	if !found {
		port = defaultPort
	}
	queueName, found := os.LookupEnv("QUEUE_NAME")
	if !found {
		queueName = defaultQueueName
	}

	mysqlDbAddr, found := os.LookupEnv("MYSQL_CONNECTION")
	if !found {
		mysqlDbAddr = defaultMysqlConn
	}

	rabbitMqAddr, found := os.LookupEnv("RABBITMQ_CONNECTION")
	if !found {
		rabbitMqAddr = defaultRabbitMqConn
	}

	dbClient, err := ent.Open("mysql", mysqlDbAddr)
	must(err, "failed opening connection to mysql")
	defer dbClient.Close()

	err = dbClient.Schema.Create(ctx)
	must(err, "failed creating schema resources")

	rabbitMqConn, err := amqp.Dial(rabbitMqAddr)
	must(err, "failed to connect to RabbitMQ")
	defer rabbitMqConn.Close()

	rabbitMqCh, err := rabbitMqConn.Channel()
	must(err, "failed to open a RabbitMQ channel")
	defer rabbitMqCh.Close()

	queue, err := rabbitmq_queue.New(rabbitMqCh, queueName)
	must(err, "init rabbitMQ client")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	taskService := task.NewService(dbClient, queue, httpClient)

	srv := server.New(taskService)
	must(err, "init server")

	router := mux.NewRouter()
	srv.MountHandlers(router)

	// if we have old tasks that didn't run, try to run them now
	if err := taskService.ProcessOldTasks(ctx); err != nil {
		log.Println("failed to process old tasks", err)
	}

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Seconds().Do(func() {
		ctx := logx.ContextWithTraceID(context.Background())
		if err := taskService.ProcessCurrentTasks(ctx); err != nil {
			logx.Error(ctx, err)
		}
	})
	s.StartAsync()

	err = startConsumeMessages(queue, taskService)
	must(err, "failed to listen for queue")

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to listen server: %v\n", err)
		}
	}()
	log.Println("server is listening on port", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("shutdown server...")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server exited")
}

// startConsumeMessages start consume messages from rabbitMQ, emit every message to taskService.
// 	This listener can sit in different place or service, so this is not part of the Queue interface
// 	TODO configure to queue to have retries
func startConsumeMessages(queue *rabbitmq_queue.Client, taskService *task.Service) error {
	return queue.Consume(func(d *amqp.Delivery) {
		defer d.Ack(false) // ack the message as soon as this function exist normally

		ctx := logx.ContextWithTraceID(context.Background())
		var t task.Task
		if err := json.Unmarshal(d.Body, &t); err != nil {
			logx.Error(ctx, "failed to parse task from queue:", err)
			return
		}
		logx.Info(ctx, "emitting task", t)
		if err := taskService.EmitTask(ctx, &t); err != nil {
			logx.Error(ctx, "failed to emit task:", err)
		}
	})
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
