package redis

import (
	"fmt"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/utils"

	"github.com/go-redis/redis"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/circuit"
)

const (
	// ErrInvalidConfigurationError : Error Invalid Configuration Error
	ErrInvalidConfigurationError = "ErrInvalidConfigurationError"
	// defaultCommandName : Default command name for redis circuit breaker
	defaultCommandName = "RedisCommand"
)

// ClientImpl : Redis client implementation
type clientImpl struct {
	config *Config
	client *redis.Client
}

// Z represents sorted set member. :Redis version
type Z struct {
	Score  float64
	Member interface{}
}

// GetService is a function to return service instance
func GetService(transactionID string, config *Config) Client {
	if transactionID == "" {
		transactionID = utils.GetTransactionID()
	}
	if config.CommandName == "" {
		config.CommandName = fmt.Sprintf("%s_%s", defaultCommandName, transactionID)
	}
	circuit.Register(transactionID, config.CommandName, &config.CircuitBreaker, nil)
	return &clientImpl{config: config}
}

func (c *clientImpl) Init() error {
	if c.client == nil {
		redisClient, err := c.generateRedisClient()
		if err != nil {
			return err
		}
		c.client = redisClient
	}
	return nil
}

func (c *clientImpl) generateRedisClient() (*redis.Client, error) {

	if c.config == nil {
		return nil, fmt.Errorf(ErrInvalidConfigurationError)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:               c.config.ServerAddress[0],
		Password:           c.config.Password,
		DB:                 c.config.DB,
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

	return redisClient, nil
}

func (c *clientImpl) Close() error {
	if c.client != nil {
		return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
			return c.client.Close()
		}, nil)
	}
	return nil
}

func (c *clientImpl) Set(key string, value interface{}) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.client.Set(key, value, -1).Err()
	}, nil)
}

func (c *clientImpl) Get(key string) (interface{}, error) {
	var (
		result interface{}
		err    error
	)

	breakerErr := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.client.Get(key).Result()
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
func (c *clientImpl) MGet(keys ...string) ([]interface{}, error) {
	var results []interface{}

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		results, execErr = c.client.MGet(keys...).Result()
		return execErr
	}, nil)

	return results, err
}

func (c *clientImpl) Delete(key ...string) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.client.Del(key...).Err()
	}, nil)
}

func (c *clientImpl) Expire(key string, duration time.Duration) (bool, error) {
	var expire bool

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		expire, execErr = c.client.Expire(key, duration).Result()
		return execErr
	}, nil)

	return expire, err
}

func (c *clientImpl) Incr(key string) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.client.Incr(key).Result()
		return execErr
	}, nil)

	return value, err
}

// IncrBy - returns the value from Redis that is increased by input value
func (c *clientImpl) IncrBy(key string, count int64) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.client.IncrBy(key, count).Result()
		return execErr
	}, nil)

	return value, err
}

// Decr - returns the decremented value from Redis
func (c *clientImpl) Decr(key string) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.client.Decr(key).Result()
		return execErr
	}, nil)

	return value, err
}

// DecrBy - returns the value from Redis that is decreased by input value
func (c *clientImpl) DecrBy(key string, count int64) (int64, error) {
	var value int64

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		value, execErr = c.client.DecrBy(key, count).Result()
		return execErr
	}, nil)

	return value, err
}

// Keys: receive all keys according to specified pattern
func (c *clientImpl) Keys(pattern string) ([]string, error) {
	var results []string

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var err error
		results, err = c.client.Keys(pattern).Result()
		return err
	}, nil)

	return results, err
}

func (c *clientImpl) SetWithExpiry(key string, value interface{}, duration time.Duration) error {
	return circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		return c.client.Set(key, value, duration).Err()
	}, nil)
}

func (c *clientImpl) Scan(cursor uint64, match string, count int64) (keys []string, outCursor uint64, err error) {
	err = circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		keys, outCursor, execErr = c.client.Scan(cursor, match, count).Result()
		return execErr
	}, nil)

	return
}

func (c *clientImpl) SubscribeChannel(pattern string) (<-chan *redis.Message, error) {
	pubSub := c.client.Subscribe(pattern)
	return pubSub.Channel(), nil
}

func (c *clientImpl) CreatePipeline() Pipeliner {
	return &pipe{
		pipeliner: c.client.Pipeline(),
	}
}

// ClosePipeliner : Close Pipeliner
func (p *pipe) ClosePipeliner() error {
	if p.pipeliner != nil {
		return p.pipeliner.Close()
	}
	return nil
}

// PSet : Pipeliner Set
func (p *pipe) PSet(key string, value interface{}) error {
	return p.pipeliner.Set(key, value, -1).Err()
}

// PSetWithExpiry : Pipeliner Set With Expiry
func (p *pipe) PSetWithExpiry(key string, value interface{}, duration time.Duration) error {
	return p.pipeliner.Set(key, value, duration).Err()
}

// PGet : Pipeliner Get
func (p *pipe) PGet(key string) error {
	return p.pipeliner.Get(key).Err()
}

// PSAdd : Pipeliner Set Add Member
func (p *pipe) PSAdd(key string, member ...interface{}) *redis.IntCmd {
	return p.pipeliner.SAdd(key, member...)
}

// PSRem : Pipeliner Set Remove Member
func (p *pipe) PSRem(key string, member ...interface{}) *redis.IntCmd {
	return p.pipeliner.SRem(key, member...)
}

// Incr : Pipeliner Incr value by key
func (p *pipe) Incr(key string) *redis.IntCmd {
	return p.pipeliner.Incr(key)
}

// Expire : Pipeliner Expire value by key and duration
func (p *pipe) Expire(key string, duration time.Duration) *redis.BoolCmd {
	return p.pipeliner.Expire(key, duration)
}

// Exec : Pipeliner Exec
func (p *pipe) Exec() ([]CmdOut, error) {
	out, err := p.pipeliner.Exec()
	outArray := []CmdOut{}
	if len(out) > 0 {
		for _, outData := range out {
			outArray = append(outArray, CmdOut{
				Name: outData.Name(),
				Args: outData.Args(),
				Err:  outData.Err(),
			})
		}
	}
	return outArray, err
}

// SAdd: Add member/members to a set
func (c *clientImpl) SAdd(key string, members ...interface{}) (int64, error) {
	var sAddResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sAddResult, execErr = c.client.SAdd(key, members...).Result()
		return execErr
	}, nil)
	return sAddResult, err
}

// SMembers: Get List of members from a set
func (c *clientImpl) SMembers(key string) ([]string, error) {
	var SMembersResult []string
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		SMembersResult, execErr = c.client.SMembers(key).Result()
		return execErr
	}, nil)
	return SMembersResult, err
}

// SRem: Remove member/members from a set
func (c *clientImpl) SRem(key string, members ...interface{}) (int64, error) {
	var sRemResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sRemResult, execErr = c.client.SRem(key, members...).Result()
		return execErr
	}, nil)
	return sRemResult, err
}

// SIsMember: Check if member is present in set or not
func (c *clientImpl) SIsMember(key string, member interface{}) (bool, error) {
	var sIsMemberResult bool
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sIsMemberResult, execErr = c.client.SIsMember(key, member).Result()
		return execErr
	}, nil)
	return sIsMemberResult, err
}

// SUnionStore: Perform union on sets with specified keys and store it in destination
func (c *clientImpl) SUnionStore(destination string, keys ...string) (int64, error) {
	var sUnionStoreResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		sUnionStoreResult, execErr = c.client.SUnionStore(destination, keys...).Result()
		return execErr
	}, nil)
	return sUnionStoreResult, err
}

// ZAdd: Add member to a sorted set, or update its score if it already exists
func (c *clientImpl) ZAdd(key string, members ...Z) (int64, error) {
	var zAddResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		z := make([]redis.Z, len(members))
		for i := 0; i < len(members); i++ {
			z[i].Score = members[i].Score
			z[i].Member = members[i].Member
		}
		zAddResult, execErr = c.client.ZAdd(key, z...).Result()
		return execErr
	}, nil)
	return zAddResult, err
}

// ZRange:  Return a range of members in a sorted set, by index( Start: starting index, STOP : ending index)
func (c *clientImpl) ZRange(key string, start, stop int64) ([]string, error) {
	var zRangeResult []string
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		zRangeResult, execErr = c.client.ZRange(key, start, stop).Result()
		return execErr
	}, nil)
	return zRangeResult, err
}

// ZRem: Remove one or more members from a sorted set
func (c *clientImpl) ZRem(key string, member interface{}) (int64, error) {
	var zRemResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		zRemResult, execErr = c.client.ZRem(key, member).Result()
		return execErr
	}, nil)
	return zRemResult, err
}

// Exists: check existance of key in set
func (c *clientImpl) Exists(key string) (int64, error) {
	var existResult int64
	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		existResult, execErr = c.client.Exists(key).Result()
		return execErr
	}, nil)
	return existResult, err
}

// Ping: check Availabilty of redis
func (c *clientImpl) Ping() (string, error) {
	var result string

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.client.Ping().Result()
		return execErr
	}, nil)

	return result, err
}

// TTL: Returns the remaining time to live of a key that has a timeout
func (c *clientImpl) TTL(key string) (time.Duration, error) {
	var result time.Duration

	err := circuit.Do(c.config.CommandName, c.config.CircuitBreaker.Enabled, func() error {
		var execErr error
		result, execErr = c.client.TTL(key).Result()
		return execErr
	}, nil)

	return result, err
}
