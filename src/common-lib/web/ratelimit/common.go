package ratelimit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const expireMultiplier int64 = 2

func windowTimestamp(now, interval int64) int64 {
	return (now / interval) * interval
}

func storageKey(group, key string, timestamp int64) string {
	return fmt.Sprintf("%s:%s:%d", group, key, timestamp)
}

func increment(storage Storage, key string, interval int64) (int64, error) {
	c, err := storage.Incr(key)
	if err != nil {
		return 0, err
	}
	if err = touch(storage, key, interval); err != nil {
		return 0, err
	}
	return c, nil
}

func touch(storage Storage, key string, interval int64) error {
	res, err := storage.Expire(key, time.Second*time.Duration(interval*expireMultiplier))
	if err != nil {
		return err
	}
	if !res {
		return fmt.Errorf("failed to set TTL for key %s", key)
	}
	return nil
}

func count(storage Storage, key string) (int64, error) {
	v, err := storage.Get(key)
	if err != nil && !isNotFoundError(err) {
		return 0, err
	}
	c, err := toInt64(v)
	if err != nil {
		return 0, err
	}
	return c, nil
}

func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		if v == "" {
			return 0, nil
		}
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type to convert, value: %#v", v)
	}
}

func isNotFoundError(err error) bool {
	str := strings.ToLower(err.Error())
	if strings.Contains(str, "redis: nil") {
		return true
	}
	return strings.Contains(str, "not found")
}

func logError(transactionID, errorCode string, err error) {
	logger.Get().Error(transactionID, errorCode, "%v", err)
}
