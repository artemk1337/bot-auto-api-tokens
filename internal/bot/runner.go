package bot

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/artemk1337/bot-auto-api-tokens/internal/telegram"
)

type Poller interface {
	GetUpdates(ctx context.Context, offset int, timeout time.Duration) ([]telegram.Update, error)
}

type Runner struct {
	poller      Poller
	service     *Service
	pollTimeout time.Duration
	jobs        chan telegram.Message
}

func NewRunner(poller Poller, service *Service, pollTimeout time.Duration, queueSize int) Runner {
	if queueSize < 1 {
		queueSize = 1
	}
	return Runner{
		poller:      poller,
		service:     service,
		pollTimeout: pollTimeout,
		jobs:        make(chan telegram.Message, queueSize),
	}
}

func (r Runner) Run(ctx context.Context) error {
	errs := make(chan error, 1)
	go r.work(ctx)
	go r.poll(ctx, errs)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errs:
		return err
	}
}

func (r Runner) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-r.jobs:
			if err := r.service.HandleMessage(ctx, msg); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("handle message: %v", err)
			}
		}
	}
}

func (r Runner) poll(ctx context.Context, errs chan<- error) {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := r.poller.GetUpdates(ctx, offset, r.pollTimeout)
		if err != nil {
			log.Printf("get updates: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			if update.Message.Text == "" {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case r.jobs <- update.Message:
			default:
				errs <- errors.New("message queue is full")
				return
			}
		}
	}
}
