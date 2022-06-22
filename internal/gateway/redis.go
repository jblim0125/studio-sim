package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jblim0125/cache-sim/models"
	"github.com/sirupsen/logrus"
	"strconv"
)

// RedisGateway redis struct
type RedisGateway struct {
	mainClient *redis.Client
	subClient  *redis.Client
	log        *logrus.Logger
}

var redisGWInstance *RedisGateway

// New create instance
func (RedisGateway) New(log *logrus.Logger, redisConf models.RedisConfig) (*RedisGateway, error) {
	redisGateway := &RedisGateway{
		log: log,
	}
	redisAddr := fmt.Sprintf("%s:%d", redisConf.IP, redisConf.Port)
	mainNamespace, _ := strconv.Atoi(redisConf.Database)
	redisGateway.mainClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: redisConf.ID,
		Password: redisConf.PW,
		DB:       mainNamespace,
	})
	// connection check
	isConnected := RedisGateway{}.PingPong(redisGateway.mainClient)
	if !isConnected {
		redisGateway.log.Error("failed to check connection to redis")
		return nil, fmt.Errorf("failed to check connection to redis")
	}
	subNamespace, _ := strconv.Atoi(redisConf.SubDatabase)
	redisGateway.subClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: redisConf.ID,
		Password: redisConf.PW,
		DB:       subNamespace,
	})

	redisGWInstance = redisGateway
	redisGateway.log.Errorf("[ Redis Gateway ] Start ........................................................... [ OK ]")
	return redisGateway, nil
}

// GetInstance return redis gateway instance
func (RedisGateway) GetInstance() *RedisGateway {
	return redisGWInstance
}

// PingPong check connection
func (RedisGateway) PingPong(client *redis.Client) bool {
	result, err := client.Ping(context.Background()).Result()
	if err != nil || result != "PONG" {
		return false
	}
	return true
}

// Destroy redis gateway destory
func (redisGateway *RedisGateway) Destroy() error {
	var err error
	err = redisGateway.mainClient.Close()
	if err != nil {
		redisGateway.log.Error(err)
	}
	err = redisGateway.subClient.Close()
	if err != nil {
		redisGateway.log.Error(err)
	}
	return nil
}

// Get get cache data
func (redisGateway *RedisGateway) Get(key string) ([]byte, error) {
	result, err := redisGateway.mainClient.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetSubData return cache sub data
func (redisGateway *RedisGateway) GetSubData(key string) (*models.CacheSubData, error) {
	result, err := redisGateway.mainClient.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}
	data := &models.CacheSubData{}
	err = json.Unmarshal(result, data)
	if err != nil {
		return nil, err
	}
	data.DSL = key
	return data, nil
}
