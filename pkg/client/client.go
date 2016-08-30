package client

import (
	"fmt"

	"golang.org/x/oauth2"

	gh "github.com/google/go-github/github"
	workq "github.com/iamduo/go-workq"
	"github.com/mfojtik/gitshift/pkg/api"
	redis "gopkg.in/redis.v4"
)

// GetAll gets all required clients for external services
func GetAll() (*workq.Client, *Github, *redis.Client, error) {
	var err error
	config := api.EnvToConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config["REDIS_SERVICE_HOST"] + ":" + config["REDIS_SERVICE_PORT"],
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("redis: %v", err)
	}
	workqClient, err := workq.Connect(config["WORKQ_SERVER_SERVICE_HOST"] + ":" + config["WORKQ_SERVER_SERVICE_PORT"])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("workq: %v", err)
	}
	githubClient := &Github{gh.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config["GITHUB_API_KEY"]})))}
	return workqClient, githubClient, redisClient, nil
}
