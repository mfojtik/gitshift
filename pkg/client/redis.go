package client

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/mfojtik/gitshift/pkg/api"
)

func StorePullRequest(pr *api.PullRequest) error {
	_, _, redis, err := GetAll()
	if err != nil {
		return err
	}
	defer redis.Close()
	now := time.Now()
	pr.UpdatedAt = &now
	log.Printf("SET %d=%q", pr.Number, pr)
	return redis.Set(fmt.Sprintf("%d", pr.Number), pr.ToJSON(), 0).Err()
}

func GetAllPulls() []*api.PullRequest {
	result := []*api.PullRequest{}
	_, _, redis, err := GetAll()
	if err != nil {
		log.Printf("ERROR: unable to connect to redis: %v", err)
		return result
	}
	defer redis.Close()
	keys, _ := redis.Keys("*").Result()
	intKeys := []int{}
	for _, k := range keys {
		i, _ := strconv.ParseInt(k, 10, 64)
		intKeys = append(intKeys, int(i))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(intKeys)))
	for _, k := range intKeys {
		rawPR, _ := redis.Get(fmt.Sprintf("%d", k)).Result()
		pr := &api.PullRequest{}
		pr.FromJSON(rawPR)
		result = append(result, pr)
	}
	return result
}

func GetPull(number int) *api.PullRequest {
	_, _, redis, err := GetAll()
	if err != nil {
		log.Printf("ERROR: unable to connect to redis: %v", err)
		return nil
	}
	defer redis.Close()
	data, err := redis.Get(fmt.Sprintf("%d", number)).Result()
	if err != nil {
		log.Printf("ERROR: Unable to get pull request %d: %v", number, err)
		return nil
	}
	result := &api.PullRequest{}
	return result.FromJSON(data)
}
