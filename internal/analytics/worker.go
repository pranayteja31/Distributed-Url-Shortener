package analytics

import (
	"context"
	"log"
	"sync"

	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
)

type Worker struct {
	clickRepo *repository.ClickRepository
	events    chan models.URLClick

	wg sync.WaitGroup
}

func NewWorker(
	clickRepo *repository.ClickRepository,
	bufferSize int,
) *Worker {
	return &Worker{
		clickRepo: clickRepo,
		events:    make(chan models.URLClick, bufferSize),
	}
}

// Record adds an analytics event to the worker queue.
func (w *Worker) Record(click models.URLClick) {
	select {
	case w.events <- click:
	default:
		// Never block the redirect request if the queue is full.
		log.Println("analytics queue full: click event dropped")
	}
}

// Start starts the analytics worker.
func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	for {
		select {
		case click := <-w.events:
			if err := w.clickRepo.Create(&click); err != nil {
				log.Printf("failed to store click event: %v", err)
			}

		case <-ctx.Done():
			// Stop accepting new work and finish queued events.
			w.drain()

			return
		}
	}
}

// drain processes events that are already present in the queue.
func (w *Worker) drain() {
	for {
		select {
		case click := <-w.events:
			if err := w.clickRepo.Create(&click); err != nil {
				log.Printf("failed to store click event during shutdown: %v", err)
			}

		default:
			return
		}
	}
}

// Wait waits until the worker has completely stopped.
func (w *Worker) Wait() {
	w.wg.Wait()
}