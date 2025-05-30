package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"github.com/yplog/ticktockbox/internal/database"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"expired_messages",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		"expired_messages_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Warning: Failed to create default queue: %v", err)
	} else {
		err = ch.QueueBind(
			queue.Name,
			"",
			"expired_messages",
			false,
			nil,
		)
		if err != nil {
			log.Printf("Warning: Failed to bind default queue: %v", err)
		} else {
			log.Printf("Default queue 'expired_messages_queue' created and bound to exchange")
		}
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) PublishExpiredMessages(messages []database.Message) error {
	for _, msg := range messages {
		body, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			continue
		}

		err = r.channel.Publish(
			"expired_messages",
			"",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		)
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		}
	}
	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
