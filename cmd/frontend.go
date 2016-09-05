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
	"fmt"
	"net/http"

	humanize "github.com/dustin/go-humanize"
	"github.com/mfojtik/gitshift/pkg/client"
	"github.com/spf13/cobra"
)

// frontendCmd represents the frontend command
var frontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", handler)
		http.ListenAndServe("0.0.0.0:8080", nil)
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	pulls := client.GetAllPulls()
	fmt.Fprintf(w, `
	<html>
	<head>
	<meta http-equiv="refresh" content="30">
	<!-- Latest compiled and minified CSS -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
	<!-- Optional theme -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
	<!-- Latest compiled and minified JavaScript -->
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
  </head>
	<body>
	<table class="table table-hover table-striped">
	<tr>
		<th>Pull Number</th>
		<th>Pull Title</th>
		<th>Test Status</th>
		<th>Merge Status</th>
		<th>Last Update</th>
		<th>Github User</th>
	</tr>
	`)
	for _, p := range pulls {
		mergeStatus := ""
		jenkinsCss := ""
		testStatus := ""
		if p.IsFailure() {
			jenkinsCss = ` class="danger"`
		}
		if p.Approved {
			mergeStatus = "<code>LGTM</code>"
		}
		if len(p.MergeStatus) > 0 {
			if p.IsMergeFailure() {
				mergeStatus = fmt.Sprintf("<a href=%q><code>MERGE FAILED</code></a>", p.MergeURL)
				jenkinsCss = ` class="danger"`
			} else if p.IsMerged() {
				mergeStatus = fmt.Sprintf("MERGED")
				jenkinsCss = ` class="info"`
			} else {
				mergeStatus = fmt.Sprintf("<a href=%q><code>MERGE %s #%d</code></a>", p.MergeURL, p.MergeStatus, p.Position)
				jenkinsCss = ` class="active"`
			}
		}
		if len(p.JenkinsTestStatus) > 0 {
			testStatus = "<a href=" + p.JenkinsTestURL + "><code>" + p.JenkinsTestStatus + "</code></a>"
			if p.IsSuccess() {
				jenkinsCss = ` class="success"`
			}
		}
		milestone := ""
		if len(p.Milestone) > 0 {
			milestone = `<span class="badge">` + p.Milestone + `</span> `
		}
		fmt.Fprintf(w, `<tr%s>
		<td><a href="https://github.com/openshift/origin/pull/%d">#%d</a></td>
		<td>%s%s</td>
		<td>%s</td>
		<td><b>%s</b></td>
		<td title="last updated %s"><small>%s</small></td>
		<td>by %s</td>
		</tr>`,
			jenkinsCss,
			p.Number, p.Number,
			milestone,
			p.Title,
			testStatus,
			mergeStatus,
			p.UpdatedAt,
			humanize.Time(p.CreatedAt),
			p.Author,
		)
	}
	fmt.Fprintf(w, `
	</table>
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
	</body>
	</html>
	`)
}

func init() {
	RootCmd.AddCommand(frontendCmd)
}
