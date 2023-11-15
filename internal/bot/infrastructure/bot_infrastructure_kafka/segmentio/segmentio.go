package bot_infastructure_kafka_segmentio

import (
	"connectly-interview/internal/bot/infrastructure/bot_infrastructure_kafka"
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

const (
	DefaultWriteTimeDuration = time.Second * 10
)

type SegmentioDialer struct {
	ctx           context.Context
	addr          string
	writeDeadline time.Duration
}

type Option func(dialer *SegmentioDialer)

func WithWriteDeadline(duration time.Duration) Option {
	return func(dialer *SegmentioDialer) {
		dialer.writeDeadline = duration
	}
}

func New(ctx context.Context, addr string, opts ...Option) bot_infrastructure_kafka.Kafka {
	d := &SegmentioDialer{
		ctx:           ctx,
		addr:          addr,
		writeDeadline: time.Second * 10,
	}

	for _, o := range opts {
		o(d)
	}

	return d
}

func (s *SegmentioDialer) Send(topic bot_infrastructure_kafka.Topic, msg []byte) error {
	var err error

	partition := 0

	conn, err := kafka.DialLeader(s.ctx, "tcp", s.addr, topic.String(), partition)
	if err != nil {
		return fmt.Errorf("could not dial kafka: %w", err)
	}

	err = conn.SetWriteDeadline(time.Now().Add(s.writeDeadline))
	if err != nil {
		return fmt.Errorf("could not set write deadline: %w", err)
	}

	_, err = conn.WriteMessages(kafka.Message{Value: msg})
	if err != nil {
		return fmt.Errorf("could not write message %q: %s", msg, err)
	}

	if err = conn.Close(); err != nil {
		return fmt.Errorf("could not close kafka dial connection")
	}

	return nil
}
