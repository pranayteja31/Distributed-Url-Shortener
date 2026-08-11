package analytics

import (
	"context"
	"log"
	"pranayteja31/Urlshortener/internal/models"
	"pranayteja31/Urlshortener/internal/repository"
)

//struct
type Worker struct {
	clickRepo *repository.ClickRepository
	event chan models.URLClick
}
//constructor
func NewWorker(clickRepo *repository.ClickRepository,bufferSize int) *Worker{
	return &Worker{
		clickRepo: clickRepo,
		event: make(chan models.URLClick,bufferSize),
	}
}
//methods
func (w *Worker)Record(click models.URLClick) {
	select{
	case w.event <- click :
	default:
		log.Println("Analytics queue if full: Click event dropped")
	}
}

func (w *Worker)Start(ctx context.Context){
	for {
		select{
		case click := <-w.event :
			if err:= w.clickRepo.Create(&click); err !=nil {
				log.Printf("Failed to store click event %v",err)
			}
		case <-ctx.Done():
			return
		}
	}
}