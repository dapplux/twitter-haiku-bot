package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dapplux/twitter-haiku-bot/config"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/ai"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/database/postgres"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/database/repositories"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/platforms"
	"github.com/dapplux/twitter-haiku-bot/scheduler"
	"github.com/dapplux/twitter-haiku-bot/services"
)

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	cfg, aErr := config.AutoLoad()
	if aErr != nil {
		log.Fatalln(aErr)

		return
	}
	db, err := postgres.New(cfg, "up", 0)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err.Error())
	}

	// Create repositories
	haikuRepo := repositories.NewHaikuRepository(db.DB)
	postRepo := repositories.NewPostRepository(db.DB)

	// Create a UnitOfWork (or transaction manager) from the base DB
	// (Assume you have an implementation that wraps db.DB for transactions.)
	txMgr := repositories.NewUnitOfWork(db.DB)

	// Initialize Platform Providers.
	// For example, TwitterPlatform and TelegramPlatform.
	// For this example, assume you have a TwitterPlatform and a TelegramPlatform.
	//twitterPlatform := platforms.NewTwitterMock()
	fmt.Println(cfg.Twitter.APIKey, cfg.Twitter.APISecret, cfg.Twitter.APIAccessToken, cfg.Twitter.APIAccessTokenSecret)
	twitterPlatform := platforms.NewTwitterProvider(cfg.Twitter.APIKey, cfg.Twitter.APISecret, cfg.Twitter.APIAccessToken, cfg.Twitter.APIAccessTokenSecret)

	// Initialize TextProcessor (for instance, a HuggingFace processor)
	textProcessor := ai.NewHuggingFaceProvider(cfg.HuggingFace.APIKey)

	// Initialize HaikuService
	haikuSvc := services.NewHaikuService(txMgr, haikuRepo, textProcessor, twitterPlatform)

	// Initialize PostService
	postSvc := services.NewPostService(postRepo, twitterPlatform)

	// Create and start the scheduler for all service functions.
	sched := scheduler.NewScheduler(haikuSvc, postSvc)
	sched.Start(rootCtx)

	// Optionally, run indefinitely.
	select {}
	// for {
	// 	scheduler.DryRunScheduler(rootCtx, haikuSvc, postSvc)
	// 	time.Sleep(1 * time.Second)
	// }
}
