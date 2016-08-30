package api

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	redis "gopkg.in/redis.v4"
)

func EnvToConfig() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) < 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result
}

func StorePullRequest(pr *PullRequest) error {
	config := EnvToConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config["REDIS_SERVICE_HOST"] + ":" + config["REDIS_SERVICE_PORT"],
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()
	now := time.Now()
	pr.UpdatedAt = &now
	return client.Set(fmt.Sprintf("%d", pr.Number), pr.ToJSON(), 0).Err()
}

func GetAllPulls() []*PullRequest {
	config := EnvToConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config["REDIS_SERVICE_HOST"] + ":" + config["REDIS_SERVICE_PORT"],
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()
	result := []*PullRequest{}
	keys, _ := client.Keys("*").Result()
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, k := range keys {
		rawPR, _ := client.Get(k).Result()
		pr := &PullRequest{}
		pr.FromJSON(rawPR)
		result = append(result, pr)
	}
	return result
}

func GetPull(number int) *PullRequest {
	config := EnvToConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config["REDIS_SERVICE_HOST"] + ":" + config["REDIS_SERVICE_PORT"],
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()
	data, err := client.Get(fmt.Sprintf("%d", number)).Result()
	if err != nil {
		log.Printf("ERROR: Unable to get pull request %d: %v", number, err)
		return nil
	}
	result := &PullRequest{}
	return result.FromJSON(data)
}
