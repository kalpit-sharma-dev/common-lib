package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"
)

// ClientImpl : Redis client implementation
type clusterClientImpl struct {
	config        *Config
	clusterClient *redis.ClusterClient
}

// GetClusterClientService is a function to return service instance
func GetClusterClientService(transactionID string, config *Config) Client {
	if transactionID == "" {
		transactionID = utils.GetTransactionID()
	}
	if config.CommandName == "" {
		config.CommandName = fmt.Sprintf("%s_%s", defaultCommandName, transactionID)
	}
	circuit.Register(transactionID, config.CommandName, &config.CircuitBreaker, nil)
	return &clusterClientImpl{config: config}
}

func (c *clusterClientImpl) Init() error {
	if c.clusterClient == nil {
		redisClient, err := c.generateRedisClusterClient()
		if err != nil {
			return err
		}
		c.clusterClient = redisClient
	}
	return nil
}

func (c *clusterClientImpl) generateRedisClusterClient() (*redis.ClusterClient, error) {

	if c.config == nil {
		return nil, fmt.Errorf(ErrInvalidConfigurationError)
	}

	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              c.config.ServerAddress,
		Password:           c.config.Password,
		MaxRedirects:       c.config.MaxRedirects,
		ReadOnly:           c.config.ReadOnly,
		RouteByLatency:     c.config.RouteByLatency,
		RouteRandomly:      c.config.RouteRandomly,
		PoolSize:           c.config.PoolSize,
		MinIdleConns:       c.config.MinIdleConns,
		MaxConnAge:         c.config.MaxConnAge,
		PoolTimeout:        c.config.PoolTimeout,
		IdleTimeout:        c.config.IdleTimeout,
		IdleCheckFrequency: c.config.IdleCheckFrequency,
		MaxRetries:         c.config.MaxRetries,
		MinRetryBackoff:    (c.config.MinRetryBackoffInMillisecond * time.Millisecond),
		MaxRetryBackoff:    (c.config.MaxRetryBackoffInMillisecond * time.Millisecond),
		DialTimeout:        (c.config.DialTimeoutInMillisecond * time.Millisecond),
		ReadTimeout:        (c.config.ReadTimeoutInMillisecond * time.Millisecond),
		WriteTimeout:       (c.config.WriteTimeoutInMillisecond * time.Millisecond),
	})

	return clusterClient, nil
}

func (c *clusterClientImpl) Close() error {
	if c.clusterClient != nil {
		return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
			return c.clusterClient.Close()
		}, nil)
	}
	return nil
}

// SAdd: Add member to a set
func (c *clusterClientImpl) SAdd(key string, members ...interface{}) (int64, error) {
	var sAddResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sAddResult, execErr = c.clusterClient.SAdd(key, members...).Result()
		return execErr
	}, nil)
	return sAddResult, err
}

// SMembers: Get List of members from a set
func (c *clusterClientImpl) SMembers(key string) ([]string, error) {
	var SMembersResult []string
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		SMembersResult, execErr = c.clusterClient.SMembers(key).Result()
		return execErr
	}, nil)
	return SMembersResult, err
}

// SRem: Remove member/members from a set
func (c *clusterClientImpl) SRem(key string, members ...interface{}) (int64, error) {
	var sRemResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sRemResult, execErr = c.clusterClient.SRem(key, members...).Result()
		return execErr
	}, nil)
	return sRemResult, err
}

// SIsMember: Check if member is present in set or not
func (c *clusterClientImpl) SIsMember(key string, member interface{}) (bool, error) {
	var sIsMemberResult bool
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sIsMemberResult, execErr = c.clusterClient.SIsMember(key, member).Result()
		return execErr
	}, nil)
	return sIsMemberResult, err
}

// SUnionStore: Perform union on sets with specified keys and store it in destination
func (c *clusterClientImpl) SUnionStore(destination string, keys ...string) (int64, error) {
	var sUnionStoreResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sUnionStoreResult, execErr = c.clusterClient.SUnionStore(destination, keys...).Result()
		return execErr
	}, nil)
	return sUnionStoreResult, err
}

// ZAdd: Add member to a sorted set, or update its score if it already exists
func (c *clusterClientImpl) ZAdd(key string, members ...Z) (int64, error) {
	var zAddResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		z := make([]redis.Z, len(members))
		for i := 0; i < len(members); i++ {
			z[i].Score = members[i].Score
			z[i].Member = members[i].Member
		}
		zAddResult, execErr = c.clusterClient.ZAdd(key, z...).Result()
		return execErr
	}, nil)
	return zAddResult, err
}

// ZRange:  Return a range of members in a sorted set, by index( Start: starting index, STOP : ending index)
func (c *clusterClientImpl) ZRange(key string, start, stop int64) ([]string, error) {
	var zRangeResult []string
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		zRangeResult, execErr = c.clusterClient.ZRange(key, start, stop).Result()
		return execErr
	}, nil)
	return zRangeResult, err
}

// ZRem: Remove one or more members from a sorted set
func (c *clusterClientImpl) ZRem(key string, member interface{}) (int64, error) {
	var zRemResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		zRemResult, execErr = c.clusterClient.ZRem(key, member).Result()
		return execErr
	}, nil)
	return zRemResult, err
}

// Exists: check existance of key in set
func (c *clusterClientImpl) Exists(key string) (int64, error) {
	var existResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		existResult, execErr = c.clusterClient.Exists(key).Result()
		return execErr
	}, nil)
	return existResult, err
}

func (c *clusterClientImpl) Set(key string, value interface{}) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.clusterClient.Set(key, value, -1).Err()
	}, nil)
}

func (c *clusterClientImpl) Get(key string) (interface{}, error) {
	var (
		result interface{}
		err    error
	)

	breakerErr := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.clusterClient.Get(key).Result()
		if execErr == redis.Nil {
			err = redis.Nil
			execErr = nil
		}
		return execErr
	}, nil)

	if breakerErr != nil {
		return result, breakerErr
	}

	return result, err
}

// MGet: Returns the values of all specified keys
func (c *clusterClientImpl) MGet(keys ...string) ([]interface{}, error) {
	var results []interface{}

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		results, execErr = c.clusterClient.MGet(keys...).Result()
		return execErr
	}, nil)

	return results, err
}

func (c *clusterClientImpl) Delete(key ...string) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.clusterClient.Del(key...).Err()
	}, nil)
}

func (c *clusterClientImpl) Expire(key string, duration time.Duration) (bool, error) {
	var expire bool

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		expire, execErr = c.clusterClient.Expire(key, duration).Result()
		return execErr
	}, nil)

	return expire, err
}

func (c *clusterClientImpl) Incr(key string) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.clusterClient.Incr(key).Result()
		return execErr
	}, nil)

	return value, err
}

// IncrBy - returns the value from Redis that is increased by input value
func (c *clusterClientImpl) IncrBy(key string, count int64) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.clusterClient.IncrBy(key, count).Result()
		return execErr
	}, nil)

	return value, err
}

func (c *clusterClientImpl) Decr(key string) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.clusterClient.Decr(key).Result()
		return execErr
	}, nil)

	return value, err
}

// DecrBy - returns the value from Redis that is decreased by input value
func (c *clusterClientImpl) DecrBy(key string, count int64) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.clusterClient.DecrBy(key, count).Result()
		return execErr
	}, nil)

	return value, err
}

// Keys: receive all keys according to specified pattern
func (c *clusterClientImpl) Keys(pattern string) ([]string, error) {
	var results []string

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		results, execErr = c.clusterClient.Keys(pattern).Result()
		return execErr
	}, nil)

	return results, err
}

func (c *clusterClientImpl) SetWithExpiry(key string, value interface{}, duration time.Duration) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.clusterClient.Set(key, value, duration).Err()
	}, nil)
}

func (c *clusterClientImpl) Scan(cursor uint64, match string, count int64) (keys []string, outCursor uint64, err error) {
	err = circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		keys, outCursor, execErr = c.clusterClient.Scan(cursor, match, count).Result()
		return execErr
	}, nil)

	return
}

func (c *clusterClientImpl) SubscribeChannel(pattern string) (<-chan *redis.Message, error) {
	pubSub := c.clusterClient.Subscribe(pattern)
	return pubSub.Channel(), nil
}

func (c *clusterClientImpl) CreatePipeline() Pipeliner {
	return &pipe{
		pipeliner: c.clusterClient.Pipeline(),
	}
}

func (c *clusterClientImpl) Ping() (string, error) {
	var result string

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.clusterClient.Ping().Result()
		return execErr
	}, nil)

	return result, err
}

// TTL: Returns the remaining time to live of a key that has a timeout
func (c *clusterClientImpl) TTL(key string) (time.Duration, error) {
	var result time.Duration

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.clusterClient.TTL(key).Result()
		return execErr
	}, nil)

	return result, err
}
