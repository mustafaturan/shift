// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package counter

import (
	"sync"
	"time"
)

type bucket map[string]uint32

// TimeBucketCounter is a capped bucket counter with a feature of auto drops of
// the stale buckets on given duration
type TimeBucketCounter struct {
	mutex sync.RWMutex

	stats   bucket
	buckets []bucket

	duration time.Duration
	timer    *time.Timer
}

// NewTimeBucketCounter inits and returns stats with given options
func NewTimeBucketCounter(capacity int, duration time.Duration) (*TimeBucketCounter, error) {
	if capacity < 1 {
		return nil, &InvalidOptionError{
			Name: "time bucket counter capacity",
			Type: "positive integer",
		}
	}

	if duration < time.Second {
		return nil, &InvalidOptionError{
			Name: "time bucket counter duration",
			Type: "positive duration(greater than or equal to a second)",
		}
	}

	counter := &TimeBucketCounter{
		buckets:  make([]bucket, capacity),
		duration: duration,
	}
	defer counter.Reset()

	return counter, nil
}

// Reset resets the stats and buckets
func (c *TimeBucketCounter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset attributes
	c.resetStats()
	c.resetBuckets()
	c.resetTimer()
}

// Increment increments the given metric by 1
func (c *TimeBucketCounter) Increment(metric string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stats[metric]++
	c.buckets[len(c.buckets)-1][metric]++
}

// Stats returns the metric values for given metrics
func (c *TimeBucketCounter) Stats(metrics ...string) map[string]uint32 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := make(map[string]uint32)
	for _, metric := range metrics {
		stats[metric] = c.stats[metric]
	}
	return stats
}

func (c *TimeBucketCounter) drop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Assign a new task to drop in the future
	c.timer = time.AfterFunc(c.duration, c.drop)

	// Drop the metrics for the fist bucket
	for metric := range c.stats {
		c.stats[metric] -= c.buckets[0][metric]
	}

	// Drop the stale bucket(shift left until the last bucket)
	for i := 0; i < len(c.buckets)-1; i++ {
		c.buckets[i] = c.buckets[i+1]
	}

	// Reset the last bucket
	c.buckets[len(c.buckets)-1] = make(bucket)
}

func (c *TimeBucketCounter) resetStats() {
	c.stats = make(bucket)
}

func (c *TimeBucketCounter) resetBuckets() {
	for i := 0; i < len(c.buckets); i++ {
		c.buckets[i] = make(bucket)
	}
}

func (c *TimeBucketCounter) resetTimer() {
	if c.timer != nil {
		c.timer.Stop()
	}

	// Assign a new task to drop the stale bucket in the future
	c.timer = time.AfterFunc(c.duration, c.drop)
}
