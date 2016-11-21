package cache

import (
	"time"

	. "github.com/coral/log"

	"github.com/garyburd/redigo/redis"
)

// CachePool 类型， 是一个redis-cache容器
type CachePool struct {
	Pool map[string]*_Redis
}

//_Redis 类型，用于 内部封装redis连接池
type _Redis struct {
	conn *redis.Pool
}

// Cache 全局变量，允许用户在任何方法中调用并操作cache
var Cache *CachePool

// InitCache 方法，初始化全局变量，cache在使用之前必须初始化，且只能初始化一次
func InitCache() *CachePool {
	Cache = &CachePool{}
	Cache.Pool = make(map[string]*_Redis)
	return Cache
}

// AddRedis 方法，添加一个redis实例
func (cache *CachePool) AddRedis(
	name, server, auth string,
	maxActive, maxIdle int) {
	redis := &_Redis{}
	redis.conn = newPool(server, auth, maxActive, maxIdle)
	cache.Pool[name] = redis
}

func newPool(server, password string, maxActive, maxIdle int) *redis.Pool {
	return &redis.Pool{
		MaxActive:   maxActive,
		MaxIdle:     maxIdle,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				Error("can not connect to redis ", server, err.Error())
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					Error("redis auth faild ", server, err.Error())
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				Error("can not ping redis", server, err.Error())
			}
			return err
		},
	}
}

func (redis *_Redis) do(cmd string,
	args ...interface{}) (reply interface{}, err error) {

	conn := redis.conn.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}

// Get 方法
func (cache *CachePool) Get(name, key string) interface{} {
	val, err := redis.Bytes(cache.Pool[name].do("GET", key))
	if err != nil {
		Error("redis get error", name, key, err.Error())
		return nil
	}
	return string(val)
}

// Set 方法
func (cache *CachePool) Set(name, key string, val interface{}) bool {
	val, err := cache.Pool[name].do("SET", key, val)
	if err != nil {
		Error("redis set error", name, key, err.Error())
		return false
	}
	return val == "OK"
}

// Expire 方法
func (cache *CachePool) Expire(name, key string, expire int) bool {
	val, err := cache.Pool[name].do("EXPIRE", key, expire)
	if err != nil {
		Error("redis set error", name, key, err.Error())
		return false
	}
	return val == "OK"
}
