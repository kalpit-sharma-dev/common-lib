package ratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	// SlidingWindowAlgorithm string key for the sliding window algorithm
	SlidingWindowAlgorithm = "slidingWindow"

	configStorageKey = "ratelimit_config"

	intervalMin             int64 = 30           // seconds
	intervalMax             int64 = 60 * 60 * 24 // one day
	limitMin                int64 = 1
	inMemoryCacheTTLMin     int64 = 5  // seconds
	inMemoryCacheTTLDefault int64 = 60 // seconds
)

var (
	inMemoryConfig cachedConfig

	errNilInMemoryConfig = errors.New("in-memory config is not initialized")
)

type (
	// Config contains parameters for the limiter
	Config struct {
		// Enabled - limiter is enabled or not
		Enabled bool `json:"enabled"`

		// IntervalInSec - time period of one bucket in seconds (related to all groups)
		IntervalInSec int64 `json:"intervalInSec"`

		// InMemoryCacheTTL - in seconds (in-memory cache to store the configuration)
		InMemoryCacheTTL int64 `json:"inMemoryCacheTTL"`

		// Algorithm - type of algorithm used to calculate the number of requests per time interval
		Algorithm string `json:"algorithm"`

		// Groups - limits configuration for target groups
		Groups Groups `json:"groups"`
	}

	// Groups contains group params by key
	Groups map[string]*GroupConfig

	// GroupConfig structure for holding group params
	GroupConfig struct {
		// Overrides - contains limit for the unique endpoint key in the current group
		Overrides map[string]int64 `json:"overrides"`

		// Limit - contains limit for the current group
		Limit int64 `json:"limit"`
	}

	countParams struct {
		Now      int64
		Group    string
		Key      string
		Interval int64
		Limit    int64
	}

	countGetter func(storage Storage, params countParams) (int64, error)

	cachedConfig struct {
		config *Config
		mu     sync.RWMutex
	}
)

func (cfg *Config) limit(group, key string) int64 {
	g, ok := cfg.Groups[group]
	if !ok {
		return 0
	}
	limit, ok := g.Overrides[key]
	if ok {
		return limit
	}
	return g.Limit
}

func (cfg *Config) validate() error {
	if cfg.IntervalInSec < intervalMin || cfg.IntervalInSec > intervalMax {
		return fmt.Errorf("not valid interval: '%d'. The value should be between %d and %d",
			cfg.IntervalInSec, intervalMin, intervalMax)
	}

	for group, c := range cfg.Groups {
		if err := validateLimit(c.Limit, group, ""); err != nil {
			return err
		}
		for key, limit := range c.Overrides {
			if err := validateLimit(limit, group, key); err != nil {
				return err
			}
		}
	}

	if _, err := resolveCountGetter(cfg.Algorithm); err != nil {
		return err
	}

	return nil
}

func validateLimit(limit int64, group, key string) error {
	if limit < limitMin {
		s := fmt.Sprintf("limit cannot be less that %d for group: %s", limitMin, group)
		if key != "" {
			s = fmt.Sprintf("%s, key: %s", s, key)
		}
		return errors.New(s)
	}
	return nil
}

func resolveCountGetter(algorithm string) (countGetter, error) {
	switch algorithm {
	case SlidingWindowAlgorithm:
		return slidingWindow, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm %s", algorithm)
	}
}

func config(storage Storage) (*Config, error) {
	cfg, err := configFromMemory()
	if err != nil {
		cfg, err = configFromStorage(storage)
		if err != nil {
			return nil, err
		}
		updateConfigInMemory(cfg)
	}
	return cfg, nil
}

func configFromStorage(storage Storage) (*Config, error) {
	cfg := new(Config)
	res, err := storage.Get(configStorageKey)
	if err != nil {
		return nil, err
	}

	var b []byte
	switch v := res.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil, fmt.Errorf("invalid data type, value: %#v", v)
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func configFromMemory() (*Config, error) {
	inMemoryConfig.mu.RLock()
	defer inMemoryConfig.mu.RUnlock()

	if inMemoryConfig.config == nil {
		return nil, errNilInMemoryConfig
	}
	return inMemoryConfig.config, nil
}

func updateConfig(storage Storage, cfg *Config) error {
	populateDefaultConfigValue(cfg)
	err := cfg.validate()
	if err != nil {
		return err
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = storage.Set(configStorageKey, b)
	if err != nil {
		return err
	}

	updateConfigInMemory(cfg)

	return nil
}

func updateConfigInMemory(cfg *Config) {
	inMemoryConfig.mu.Lock()
	inMemoryConfig.config = cfg
	inMemoryConfig.mu.Unlock()
}

func populateDefaultConfigValue(cfg *Config) {
	if cfg.InMemoryCacheTTL < inMemoryCacheTTLMin {
		cfg.InMemoryCacheTTL = inMemoryCacheTTLDefault
	}
}

func processInMemoryConfigValidity(ctx context.Context, storage Storage, interval int64) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			cfg, err := configFromStorage(storage)
			if err != nil {
				logError("", "InMemoryConfigValidity", err)
				continue
			}
			if interval != cfg.InMemoryCacheTTL {
				interval = cfg.InMemoryCacheTTL
				ticker.Stop()
				ticker = time.NewTicker(time.Second * time.Duration(interval))
			}
			updateConfigInMemory(cfg)
		}
	}
}
