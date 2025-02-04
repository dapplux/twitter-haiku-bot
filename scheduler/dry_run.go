package scheduler

import (
	"context"
	"log"

	"github.com/dapplux/twitter-haiku-bot/services"
)

// DryRunScheduler sequentially executes all steps of your workflow.
// It fetches posts, processes haiku summaries, processes haiku texts,
// and posts haikus to the platform. Delays are inserted between steps
// for demonstration purposes.
func DryRunScheduler(ctx context.Context, haikuSvc *services.HaikuService, postSvc *services.PostService) {
	log.Println("Starting Dry Run Scheduler...")

	// // Step 1: Fetch and Save posts.
	// log.Println("Step 1: Fetching posts and saving them.")
	// if err := postSvc.FetchAndSave(ctx, 10); err != nil {
	// 	log.Printf("Error in FetchAndSave: %v", err)
	// } else {
	// 	log.Println("FetchAndSave executed successfully.")
	// }
	// // Optional delay between steps.
	// time.Sleep(2 * time.Second)

	// log.Println("Step 2: Creating haiku from unprocessed post.")
	// if err := haikuSvc.CreateHaikuFromUnprocessedPost(ctx); err != nil {
	// 	log.Printf("Error in CreateHaikuFromUnprocessedPost: %v", err)
	// } else {
	// 	log.Println("CreateHaikuFromUnprocessedPost executed successfully.")
	// }
	// time.Sleep(2 * time.Second)

	// // Step 3: Process Haiku Summary Generation.
	// log.Println("Step 3: Processing Haiku Summary.")
	// if err := haikuSvc.ProcessSummary(ctx); err != nil {
	// 	log.Printf("Error in ProcessSummary: %v", err)
	// } else {
	// 	log.Println("ProcessSummary executed successfully.")
	// }
	// time.Sleep(2 * time.Second)

	// // Step 4: Process Haiku Text Generation.
	// log.Println("Step 4: Processing Haiku Text Generation.")
	// if err := haikuSvc.ProcessHaikuText(ctx); err != nil {
	// 	log.Printf("Error in ProcessHaikuText: %v", err)
	// } else {
	// 	log.Println("ProcessHaikuText executed successfully.")
	// }
	// time.Sleep(2 * time.Second)

	// Step 5: Post Haiku to Platform.
	log.Println("Step 5: Posting Haiku to Platform.")
	if err := haikuSvc.PostHaiku(ctx); err != nil {
		log.Printf("Error in PostHaiku: %v", err)
	} else {
		log.Println("PostHaiku executed successfully.")
	}

	log.Println("Dry Run Scheduler completed.")
}
