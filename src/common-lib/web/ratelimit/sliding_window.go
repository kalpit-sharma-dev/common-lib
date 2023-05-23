package ratelimit

import "math"

func slidingWindow(storage Storage, p countParams) (int64, error) {
	ts := windowTimestamp(p.Now, p.Interval)
	key := storageKey(p.Group, p.Key, ts)
	current, err := count(storage, key)
	if err != nil {
		return 0, err
	}
	if current > p.Limit {
		return current, nil
	}
	prev, err := windowElapsedCount(storage, ts, p)
	if err != nil {
		return 0, err
	}
	windowCount := current + prev
	if windowCount > p.Limit {
		return windowCount, nil
	}
	_, err = increment(storage, key, p.Interval)
	if err != nil {
		return 0, err
	}
	return windowCount + 1, nil
}

func windowElapsedCount(storage Storage, windowTS int64, p countParams) (int64, error) {
	key := storageKey(p.Group, p.Key, windowTimestamp(p.Now-p.Interval, p.Interval))
	c, err := count(storage, key)
	if err != nil {
		return 0, err
	}
	return int64(math.Ceil(float64(c) * windowElapsedPercent(p.Now, windowTS, p.Interval))), nil
}

func windowElapsedPercent(now, windowTS, interval int64) float64 {
	return 1 - (float64(now-windowTS) / float64(interval))
}
