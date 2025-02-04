package scheduler

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/robfig/cron/v3"

	"github.com/dapplux/twitter-haiku-bot/services"
)

type Scheduler struct {
	cron          *cron.Cron
	haikuService  *services.HaikuService
	postService   *services.PostService
	platformIndex uint64 // for round-robin if needed
	// Optionally, you can maintain counters for monthly usage.
	monthlyFetchCount int64
	monthlyFetchLimit int64 // e.g., 100
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(haikuSvc *services.HaikuService, postSvc *services.PostService) *Scheduler {
	return &Scheduler{
		cron:              cron.New(cron.WithSeconds()),
		haikuService:      haikuSvc,
		postService:       postSvc,
		monthlyFetchLimit: 100,
	}
}

// Start configures and starts all scheduled jobs.
func (s *Scheduler) Start(ctx context.Context) {
	// Schedule fetching posts every Tuesday evening at 18:00.
	// Cron format: "MINUTE HOUR DOM MONTH DOW"
	// This expression means: At minute 0, hour 18, every day-of-month, every month, on Tuesday.
	_, err := s.cron.AddFunc("0 0 18 * * 2", func() {
		if atomic.LoadInt64(&s.monthlyFetchCount) >= s.monthlyFetchLimit {
			log.Println("Monthly fetch limit reached, skipping fetch cycle.")
			return
		}

		log.Println("Running PostService.FetchAndSave")
		// Fetch posts (limit could be 10 posts per fetch)
		if err := s.postService.FetchAndSave(ctx, 10); err != nil {
			log.Printf("Error in FetchAndSave: %v", err)
			return
		}

		atomic.AddInt64(&s.monthlyFetchCount, 1)
	})
	if err != nil {
		log.Printf("Failed to schedule FetchAndSave: %v", err)
	}

	// Schedule sCreateHaikuFromUnprocessedPost job.
	_, err = s.cron.AddFunc("@every 1m", func() {
		log.Println("Running HaikuService.CreateHaikuFromUnprocessedPost")
		if err := s.haikuService.CreateHaikuFromUnprocessedPost(ctx); err != nil {
			log.Printf("Error in CreateHaikuFromUnprocessedPost: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule CreateHaikuFromUnprocessedPost: %v", err)
	}

	// Schedule summary processing job.
	// Runs every 2 hours (adjust if necessary).
	_, err = s.cron.AddFunc("@every 2m", func() {
		log.Println("Running HaikuService.ProcessSummary")
		if err := s.haikuService.ProcessSummary(ctx); err != nil {
			log.Printf("Error in ProcessSummary: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule ProcessSummary: %v", err)
	}

	// Schedule haiku text generation job.
	// Runs every 2 hours, offset by 30 minutes from the summary processing job.
	_, err = s.cron.AddFunc("@every 2m", func() {
		log.Println("Running HaikuService.ProcessHaikuText")
		if err := s.haikuService.ProcessHaikuText(ctx); err != nil {
			log.Printf("Error in ProcessHaikuText: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule ProcessHaikuText: %v", err)
	}

	// Schedule posting haikus.
	// The cron expression "0 0 */3 * * *" means: at second 0, minute 0, every 3rd hour of every day.
	_, err = s.cron.AddFunc("0 0 */3 * * *", func() {
		log.Println("Running HaikuService.PostHaiku")
		if err := s.haikuService.PostHaiku(ctx); err != nil {
			log.Printf("Error in PostHaiku: %v", err)
		}
	})
	if err != nil {
		log.Printf("Failed to schedule PostHaiku: %v", err)
	}

	s.cron.Start()
	log.Println("Scheduler started")
}

// Stop stops the cron scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}
