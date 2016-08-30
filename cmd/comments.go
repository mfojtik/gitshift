package cmd

import (
	"log"
	"os"
	"time"

	"github.com/mfojtik/gitshift/pkg/client"
	"github.com/mfojtik/gitshift/pkg/processor"
	"github.com/spf13/cobra"
)

var streamCommentsCmd = &cobra.Command{
	Use:   "fetch-comments",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Streaming Github comments ...")
		for {
			for _, pull := range client.GetAllPulls() {
				if pull.IsMerged() {
					continue
				}
				for _, comment := range pull.CommentsToPayload() {
					job, err := client.AddJob("process-comment", comment, time.Duration(60*time.Second), time.Duration(120*time.Second))
					if err != nil {
						log.Printf("ERROR: failed to add process-comment job: %v", err)
						continue
					}
					log.Printf("added %q job %q for comment %q", "process-comment", job.UUID, string(comment))
				}
			}
			log.Printf("sleeping 2 minutes before next fetch ...")
			time.Sleep(2 * time.Minute)
		}
	},
}

var processCommentsCmd = &cobra.Command{
	Use:   "process-comments",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Processing Github comments...")
		if err := processor.ProcessComments(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(streamCommentsCmd)
	RootCmd.AddCommand(processCommentsCmd)
}
