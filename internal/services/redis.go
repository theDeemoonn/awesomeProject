package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

var ctx = context.Background()

type RedisService struct {
	Client *redis.Client
}

func NewRedisService() *RedisService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // адрес Redis сервера
		Password: "",           // пароль (если есть)
		DB:       0,            // используемая база данных
	})

	return &RedisService{
		Client: rdb,
	}
}

func (r *RedisService) CacheEntity(entityID string, entity interface{}) error {
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	err = r.Client.Set(ctx, entityID, entityJSON, 1*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisService) GetCachedEntity(entityID string, entity interface{}) error {
	entityJSON, err := r.Client.Get(ctx, entityID).Result()
	if err == redis.Nil {
		return err
	} else if err != nil {
		return err
	}

	if entityJSON == "" {
		return errors.New("no data in cache for entity: " + entityID)
	}

	err = json.Unmarshal([]byte(entityJSON), &entity)
	if err != nil {
		return err
	}

	return nil
}
