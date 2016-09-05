package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/mfojtik/gitshift/pkg/client"
	"github.com/spf13/cobra"
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		for {
			log.Printf("Pruning old pulls ...")
			toDelete := []string{}
			for _, pull := range client.GetAllPulls() {
				if pull.UpdatedAt == nil {
					continue
				}
				expiresAt := pull.UpdatedAt.Add(48 * time.Hour)
				if expiresAt.Before(time.Now()) {
					continue
				}
				toDelete = append(toDelete, fmt.Sprintf("%d", pull.Number))
			}
			if len(toDelete) > 0 {
				log.Printf("Found %d PR's to prune: %#v", len(toDelete), toDelete)
				if err := client.DeletePulls(toDelete...); err != nil {
					log.Printf("Unable to prune: %v", err)
				}
			}
			time.Sleep(1 * time.Minute)
		}
	},
}

func init() {
	RootCmd.AddCommand(pruneCmd)
}
