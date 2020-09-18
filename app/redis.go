package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
	"time"
)

const (
	// 全局计数器
	URLIDKey = "next.url.id"
	// 映射短地址到url
	ShortLinkKey = "shortlink:%s:url"
	// 映射url的hash值到短地址
	URLHashKey = "urlhash:%s:shortlink"
	// 映射短地址到url的详细信息
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

// 创建短地址
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	// 将url转换成sha1 hash
	h := toSHA1(url)
	// 检查是否在缓存中
	d, err := r.Cli.Get(fmt.Sprintf(URLHashKey, h)).Result()
	if err == redis.Nil {
		// 不存在，什么都不做
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// 过期了，什么都不做 ???
		} else {
			// 在缓存中直接返回
			return d, nil
		}
	}

	// 自增全局自增器
	err = r.Cli.Incr(URLIDKey).Err()
	if err != nil {
		return "", nil
	}

	// 把全局key编码为base62
	id, err := r.Cli.Get(URLIDKey).Int64()
	if err != nil {
		return "", nil
	}

	// 短地址
	eid := base62.EncodeInt64(id)

	// 映射短地址到url
	err = r.Cli.Set(fmt.Sprintf(ShortLinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", nil
	}

	// 映射hash值到短地址
	err = r.Cli.Set(fmt.Sprintf(URLHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", nil
	}

	// 映射短地址到url的详细信息
	detail, err := json.Marshal(
		&URLDetail{
			URL:                 url,
			CreatedAt:           time.Now().String(),
			ExpirationInMinutes: time.Duration(exp),
		})
	if err != nil {
		return "", err
	}
	err = r.Cli.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", nil
	}
	return eid, nil
}

// 获取短地址信息
func (r *RedisCli) ShortLinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(fmt.Sprintf(ShortlinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("UnKnown short URL")}
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}

// 短地址转换为长地址
func (r *RedisCli) UnShorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortLinkKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("UnKnown short URL")}
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

// 定义短地址的详细信息
type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, pwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{c}
}

func toSHA1(url string) string {
	return ""
}
