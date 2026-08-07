package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) Cache{
	return &RedisCache{
		client: client,
	}
}
//get method
func (r *RedisCache)Get(ctx context.Context,key string)([]byte,error){
	return r.client.Get(ctx,key).Bytes()
}
//set method
func (r *RedisCache)Set(ctx context.Context,key string, value []byte,ttl time.Duration) error{
	return r.client.Set(ctx,key,value,ttl).Err()
}
//delete method 
func(r *RedisCache)Delete(ctx context.Context,key string)error{
	return r.client.Del(ctx,key).Err()
}