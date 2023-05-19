package rabbitmq_queue

import (
	"context"
	"encoding/json"
	"github.com/Av1shay/timers-scheduler-demo/task"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	ch    *amqp.Channel
	queue *amqp.Queue
}

func New(ch *amqp.Channel, queueName string) (*Client, error) {
	queue, err := ch.QueueDeclare(
		queueName,
		true, // durable - the queue won't lose items if the process crashes
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Client{ch, &queue}, nil
}

func (c *Client) Consume(cb func(d *amqp.Delivery)) error {
	msgs, err := c.ch.Consume(
		c.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			cb(&d)
		}
	}()

	return nil
}

func (c *Client) Publish(ctx context.Context, task *task.Task) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return c.ch.PublishWithContext(ctx,
		"",
		c.queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         b,
		})
}
