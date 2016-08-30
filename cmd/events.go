// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/mfojtik/gitshift/pkg/client"
	"github.com/mfojtik/gitshift/pkg/processor"
	"github.com/spf13/cobra"
)

var streamEventsCmd = &cobra.Command{
	Use:   "fetch-events",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_, github, _, err := client.GetAll()
		if err != nil {
			log.Printf("ERROR: %v", err)
			os.Exit(1)
		}
		for {
			for _, e := range github.Events() {
				payload, err := json.Marshal(e)
				if err != nil {
					log.Printf("ERROR: unable to encode %+v", e)
					continue
				}
				if job, err := client.AddJob("process-event", payload, time.Duration(60*time.Second), time.Duration(120*time.Second)); err != nil {
					log.Printf("ERROR: unable to create job: %v", err)
					continue
				} else {
					log.Printf("added %q job %q for %q event", "process-event", job.UUID, *e.Type)
				}
			}
			time.Sleep(30 * time.Second)
		}
	},
}

var processEventsCmd = &cobra.Command{
	Use:   "process-events",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Processing Github events ...")
		if err := processor.ProcessEvents(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(streamEventsCmd)
	RootCmd.AddCommand(processEventsCmd)

}
