package cache

import (
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"

	log "github.com/sirupsen/logrus"
)

var (
	Client     *redis.Client
	KeysPrefix string = "keys:"
)

func Init() error {
	l := log.WithFields(log.Fields{
		"package": "cache",
	})
	l.Debug("Initializing redis client")
	Client = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password:    os.Getenv("REDIS_PASS"), // no password set
		DB:          0,                       // use default DB
		DialTimeout: 30 * time.Second,
		ReadTimeout: 30 * time.Second,
	})
	cmd := Client.Ping()
	if cmd.Err() != nil {
		l.Error("Failed to connect to redis")
		return cmd.Err()
	}
	l.Debug("Connected to redis")
	return nil
}
