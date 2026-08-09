package cache

import (
	"context"
	"pranayteja31/Urlshortener/internal/metrics"
	"time"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Cache{
	return &RedisCache{
		client: client,
	}
}
//get method
func (r *RedisCache)Get(ctx context.Context,key string)([]byte,bool,error){
	redisStart := time.Now()
	data,err := r.client.Get(ctx,key).Bytes()
	metrics.ObserveRedisLatency(redisStart)

	if err == redis.Nil {
		return nil,false,nil
	}
	if err != nil {
		return nil,false,err
	}
	return data,true,nil
}
//set method
func (r *RedisCache)Set(ctx context.Context,key string, value []byte,ttl time.Duration) error{
	return r.client.Set(ctx,key,value,ttl).Err()
}
//delete method 
func(r *RedisCache)Delete(ctx context.Context,key string)error{
	return r.client.Del(ctx,key).Err()
}