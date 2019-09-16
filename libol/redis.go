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
    addr string `json:"address"` 
    password string  `json:"password"`
    db int `json:"database"`
    
    Client *redis.Client 
}

func NewRedisCli(addr string, password string, db int) (this *RedisCli) {
    this = &RedisCli {
        addr: addr,
        password: password,
        db: db,
    }

    return 
}

func (this *RedisCli) Open() error {
    if this.Client != nil {
        return nil
    }

    client := redis.NewClient(&redis.Options {
        Addr:     "localhost:6379",
        Password: "", 
        DB:       0,
    })

    _, err := client.Ping().Result()
    if err != nil {
        return err
    } 

    this.Client = client
    
    return nil
}

func (this *RedisCli) Close() error {
    return nil
}
