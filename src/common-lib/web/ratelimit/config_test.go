package ratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	mockGroup = "internal"
	mockKey   = "123"
)

func TestConfig_limit(t *testing.T) {
	cfg := mockConfig()
	tt := []struct {
		name     string
		group    string
		key      string
		expected int64
	}{
		{
			name:     "Internal group",
			group:    mockGroup,
			key:      "",
			expected: 10,
		},
		{
			name:     "Internal group: Overrides",
			group:    mockGroup,
			key:      mockKey,
			expected: 15,
		},
		{
			name:     "Group doesn't exist",
			group:    "some_group",
			key:      "",
			expected: 0,
		},
	}

	for _, tc := range tt {
		group := tc.group
		key := tc.key
		expected := tc.expected
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, expected, cfg.limit(group, key))
		})
	}
}

func TestConfig_validate(t *testing.T) {
	invalidInterval := mockConfig()
	invalidInterval.IntervalInSec = intervalMin - 1

	invalidGroupLimit := mockConfig()
	invalidGroupLimit.Groups[mockGroup].Limit = 0

	invalidOverridesLimit := mockConfig()
	invalidOverridesLimit.Groups[mockGroup].Overrides[mockKey] = 0

	invalidAlgorithm := mockConfig()
	invalidAlgorithm.Algorithm = "invalidAlgorithm"

	tt := []struct {
		name   string
		config Config
		expErr bool
	}{
		{
			name:   "Valid",
			config: mockConfig(),
			expErr: false,
		},
		{
			name:   "Invalid: interval",
			config: invalidInterval,
			expErr: true,
		},
		{
			name:   "Invalid: group limit",
			config: invalidGroupLimit,
			expErr: true,
		},
		{
			name:   "Invalid: overrides limit",
			config: invalidOverridesLimit,
			expErr: true,
		},
		{
			name:   "Invalid: algorithm",
			config: invalidAlgorithm,
			expErr: true,
		},
	}

	for _, tc := range tt {
		cfg := tc.config
		expErr := tc.expErr
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, expErr, cfg.validate() != nil)
		})
	}
}

func Test_resolveCountGetter(t *testing.T) {
	tt := []struct {
		name      string
		algorithm string
		expErr    bool
	}{
		{
			name:      SlidingWindowAlgorithm,
			algorithm: SlidingWindowAlgorithm,
			expErr:    false,
		},
		{
			name:      "Algorithm doesn't exist",
			algorithm: "unknownAlgorithm",
			expErr:    true,
		},
	}

	for _, tc := range tt {
		algorithm := tc.algorithm
		expErr := tc.expErr
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolveCountGetter(algorithm)
			require.Equal(t, expErr, err != nil)
		})
	}
}

func Test_populateDefaultConfigValue(t *testing.T) {
	cfg := mockConfig()
	cfg.InMemoryCacheTTL = inMemoryCacheTTLMin - 1

	populateDefaultConfigValue(&cfg)

	require.Equal(t, inMemoryCacheTTLDefault, cfg.InMemoryCacheTTL)
}

func Test_inMemoryConfig(t *testing.T) {
	defer mockInMemoryConfig(nil)()
	require.Nil(t, inMemoryConfig.config)

	cfg := mockConfig()
	updateConfigInMemory(&cfg)
	got, err := configFromMemory()

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, cfg, *got)
}

func Test_configFromStorage(t *testing.T) {
	storage, finish := mockStorage(t)
	defer finish()
	cfg := mockConfig()
	cfgByte, _ := json.Marshal(cfg)
	cfgStr := string(cfgByte)
	cfgInvalidType := make([]string, 0)

	tt := []struct {
		name        string
		returnValue interface{}
		returnErr   error
		expectedCfg interface{}
		expectedErr bool
	}{
		{
			name:        "Success: string",
			returnValue: cfgStr,
			returnErr:   nil,
			expectedCfg: cfg,
			expectedErr: false,
		},
		{
			name:        "Success: []byte",
			returnValue: cfgByte,
			returnErr:   nil,
			expectedCfg: cfg,
			expectedErr: false,
		},
		{
			name:        "Failed: unsupported type",
			returnValue: cfgInvalidType,
			returnErr:   nil,
			expectedCfg: nil,
			expectedErr: true,
		},
		{
			name:        "Failed: error from storage",
			returnValue: nil,
			returnErr:   errors.New("storage error"),
			expectedCfg: nil,
			expectedErr: true,
		},
	}

	for _, tc := range tt {
		expCfg := tc.expectedCfg
		expErr := tc.expectedErr
		retV := tc.returnValue
		retErr := tc.returnErr
		t.Run(tc.name, func(t *testing.T) {
			storage.EXPECT().Get(configStorageKey).Return(retV, retErr)
			got, err := configFromStorage(storage)
			require.Equal(t, expErr, err != nil)
			if expCfg == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, expCfg, *got)
		})
	}
}

func Test_updateConfig(t *testing.T) {
	storage, finish := mockStorage(t)
	defer finish()
	cfg := mockConfig()
	cfgByte, _ := json.Marshal(cfg)

	tt := []struct {
		name string
		err  error
	}{
		{
			name: "Success",
			err:  nil,
		},
		{
			name: "Failed",
			err:  errors.New("storage error"),
		},
	}

	for _, tc := range tt {
		expErr := tc.err
		t.Run(tc.name, func(t *testing.T) {
			storage.EXPECT().Set(configStorageKey, cfgByte).Return(expErr)
			err := updateConfig(storage, &cfg)
			require.Equal(t, expErr != nil, err != nil)
		})
	}
}

func Test_config(t *testing.T) {
	storage, finish := mockStorage(t)
	defer finish()
	cfg := mockConfig()
	cfgByte, _ := json.Marshal(cfg)

	t.Run("From memory", func(t *testing.T) {
		defer mockInMemoryConfig(&cfg)()
		got, err := config(storage)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, cfg, *got)
	})
	t.Run("From storage", func(t *testing.T) {
		defer mockInMemoryConfig(nil)()
		storage.EXPECT().Get(configStorageKey).Return(cfgByte, nil)
		got, err := config(storage)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, cfg, *got)
	})
}

func Test_processInMemoryConfigValidity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := mockConfig()
	cfg.Groups[mockGroup].Overrides["testKey1"] = limitMin
	cfgByte, _ := json.Marshal(cfg)
	storage, finish := mockStorage(t)
	storage.EXPECT().Get(configStorageKey).Return(cfgByte, nil)
	defer finish()
	defer mockInMemoryConfig(nil)()

	go processInMemoryConfigValidity(ctx, storage, 1)

	require.Nil(t, inMemoryConfig.config)
	time.Sleep(time.Second * 2)
	require.NotNil(t, inMemoryConfig.config)
	require.Equal(t, cfg, *inMemoryConfig.config)
}

func mockConfig() Config {
	return Config{
		Enabled:          true,
		Algorithm:        SlidingWindowAlgorithm,
		IntervalInSec:    intervalMin,
		InMemoryCacheTTL: inMemoryCacheTTLMin,
		Groups: map[string]*GroupConfig{
			mockGroup: {
				Overrides: map[string]int64{mockKey: 15},
				Limit:     10,
			},
		},
	}
}

func mockInMemoryConfig(cfg *Config) func() {
	orig := inMemoryConfig.config
	inMemoryConfig.config = cfg
	return func() {
		inMemoryConfig.config = orig
	}
}
