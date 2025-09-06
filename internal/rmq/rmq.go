package rmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func NewPublisher(url, queue string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, ch: ch, q: q}, nil
}

func (p *Publisher) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

func (p *Publisher) PublishJSON(ctx context.Context, v any, key string) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx,
		"", p.q.Name, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			MessageId:    key,
			Body:         body,
		})
}
