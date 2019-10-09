package libol

import (
	"github.com/go-redis/redis"
)

//
// set := client.Set("key", "value", 0)
// set.Err()
// set.Val()
// get := client.Get(key)
// get.Err() # redis.Nil //not existed.
// get.Val()
// hset := client.HSet("hash", "key", "hello")
// hset.Err()
// hset.Val()
//

type RedisCli struct {
	addr     string `json:"address"`
	password string `json:"password"`
	db       int    `json:"database"`

	Client *redis.Client
}

func NewRedisCli(addr string, password string, db int) (r *RedisCli) {
	r = &RedisCli{
		addr:     addr,
		password: password,
		db:       db,
	}

	return
}

func (r *RedisCli) Open() error {
	if r.Client != nil {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	r.Client = client

	return nil
}

func (r *RedisCli) Close() error {
	return nil
}

func (r *RedisCli) HMSet(key string, value map[string]interface{}) error {
	if err := r.Open(); err != nil {
		return err
	}

	if _, err := r.Client.HMSet(key, value).Result(); err != nil {
		return err
	}
	return nil
}

func (r *RedisCli) HMDel(key string, field string) error {
	if err := r.Open(); err != nil {
		return err
	}

	if field == "" {
		if _, err := r.Client.Del(key).Result(); err != nil {
			return err
		}
	} else {
		if _, err := r.Client.HDel(key, field).Result(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisCli) HGet(key string, field string) interface{} {
	if err := r.Open(); err != nil {
		return err
	}

	hGet := r.Client.HGet(key, field)
	if hGet.Err() == nil || hGet.Err() == redis.Nil {
		return nil
	}

	return hGet.Val()
}
