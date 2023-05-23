//go:generate mockgen -package mock -destination=mock/mocks.go . Limiter,Storage

package ratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	limiterMW middleware

	errNilResponder = errors.New("responder is not initialized")
	errNilStorage   = errors.New("storage is not initialized")
	errNilLimiter   = errors.New("limiter is not initialized")
)

type (
	// Limiter is an interface to track rate limits
	Limiter interface {
		// CheckRateLimit checks the number of requests per time interval
		CheckRateLimit(next http.HandlerFunc, group, key string) http.HandlerFunc
	}

	// Storage is an interface for store usage
	Storage interface {
		Get(key string) (interface{}, error)
		Set(key string, value interface{}) error
		Incr(key string) (int64, error)
		Expire(key string, duration time.Duration) (bool, error)
	}

	// Responder is an interface to return http response back to the caller
	Responder interface {
		// RespondRateLimit handle the status code and the error received from the limiter
		RespondRateLimit(resp Response, r *http.Request, w http.ResponseWriter)
	}

	// Response is a container that stores information about error and http status code which should be returned to the caller
	Response struct {
		Error      error
		StatusCode int
	}

	middleware interface {
		Limiter
		setConfig(w http.ResponseWriter, r *http.Request)
		getConfig(w http.ResponseWriter, r *http.Request)
		setEnabled(w http.ResponseWriter, r *http.Request)
	}

	middlewareImpl struct {
		storage Storage
		rp      Responder
		count   countGetter
	}
)

// Init initializes the Limiter instance or returns an error when something goes wrong
func Init(ctx context.Context, responder Responder, storage Storage, config Config) error {
	cfg := &config
	l, err := newLimiter(responder, storage, cfg)
	if err != nil {
		return err
	}
	limiterMW = l
	go processInMemoryConfigValidity(ctx, storage, cfg.InMemoryCacheTTL)
	return nil
}

// Middleware returns the initialized Limiter instance
func Middleware() Limiter {
	return limiterMW
}

func newLimiter(responder Responder, storage Storage, cfg *Config) (middleware, error) {
	if responder == nil {
		return nil, errNilResponder
	}
	if storage == nil {
		return nil, errNilStorage
	}
	l := &middlewareImpl{
		storage: storage,
		rp:      responder,
	}
	if err := l.initConfig(cfg); err != nil {
		return nil, err
	}
	return l, nil
}

// CheckRateLimit checks the number of requests per time interval
func (l *middlewareImpl) CheckRateLimit(next http.HandlerFunc, group, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config(l.storage)
		if err != nil {
			l.rp.RespondRateLimit(Response{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}, r, w)
			return
		}

		_, ok := cfg.Groups[group]
		if !cfg.Enabled || !ok {
			next(w, r)
			return
		}

		limit := cfg.limit(group, key)
		current, err := l.count(
			l.storage,
			countParams{
				Now:      time.Now().Unix(),
				Group:    group,
				Key:      key,
				Interval: cfg.IntervalInSec,
				Limit:    limit,
			},
		)

		if err != nil {
			l.rp.RespondRateLimit(Response{
				Error:      err,
				StatusCode: http.StatusInternalServerError,
			}, r, w)
			return
		}

		if current > limit {
			l.rp.RespondRateLimit(Response{
				Error:      errors.New("allowed amount of API calls exceeded, please try later"),
				StatusCode: http.StatusTooManyRequests,
			}, r, w)
			return
		}

		limitsHeader(w, limit, limit-current)
		next(w, r)
	}
}

func (l *middlewareImpl) setConfig(w http.ResponseWriter, r *http.Request) {
	txID := transactionID(r)
	cfg := new(Config)
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		sendBadRequest(w, err, txID)
		return
	}
	if err := l.initConfig(cfg); err != nil {
		sendBadRequest(w, err, txID)
		return
	}
	sendOK(w, cfg, txID)
}

func (l *middlewareImpl) getConfig(w http.ResponseWriter, r *http.Request) {
	txID := transactionID(r)
	cfg, err := config(l.storage)
	if err != nil {
		sendInternalServerError(w, err, txID)
		return
	}
	sendOK(w, cfg, txID)
}

func (l *middlewareImpl) setEnabled(w http.ResponseWriter, r *http.Request) {
	txID := transactionID(r)
	var status enabledStatusView
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		sendBadRequest(w, err, txID)
		return
	}
	cfg, err := config(l.storage)
	if err != nil {
		sendInternalServerError(w, fmt.Errorf("setEnabled: get config error %s", err), txID)
		return
	}
	cfg.Enabled = status.Enabled
	if err = updateConfig(l.storage, cfg); err != nil {
		sendInternalServerError(w, fmt.Errorf("setEnabled: update config error %s", err), txID)
		return
	}
	sendOK(w, status, txID)
}

func (l *middlewareImpl) initConfig(cfg *Config) error {
	f, err := resolveCountGetter(cfg.Algorithm)
	if err != nil {
		return err
	}
	if err = updateConfig(l.storage, cfg); err != nil {
		return err
	}
	l.count = f
	return nil
}
