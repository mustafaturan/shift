package counter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeBucketCounter(t *testing.T) {
	t.Run("with invalid capacity", func(t *testing.T) {
		c, err := NewTimeBucketCounter(0, time.Second)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, c)
	})

	t.Run("with invalid duration", func(t *testing.T) {
		c, err := NewTimeBucketCounter(1, time.Millisecond)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, c)
	})

	t.Run("with valid options", func(t *testing.T) {
		capacity, duration := 3, 5*time.Second
		c, err := NewTimeBucketCounter(capacity, duration)

		assert.NoError(t, err)
		assert.NotNil(t, c.timer)
		assert.Equal(t, capacity, len(c.buckets))
		assert.Equal(t, duration, c.duration)
	})
}

func TestIncrement(t *testing.T) {
	metric, anotherMetric := "test", "another"
	capacity, duration := 2, time.Second

	t.Run("increments on stats and buckets", func(t *testing.T) {
		count := uint32(1)
		c, _ := NewTimeBucketCounter(capacity, duration)
		for i := 0; i < int(count); i++ {
			c.Increment(metric)
		}

		metrics := c.Stats(metric, anotherMetric)
		assert.Equal(t, count, metrics[metric])
		assert.NotEqual(t, count, metrics[anotherMetric])
	})

	t.Run("decrements with scheduled auto drop", func(t *testing.T) {
		count := uint32(2)
		c, _ := NewTimeBucketCounter(capacity, duration)
		for i := 0; i < int(count); i++ {
			c.Increment(metric)
		}

		// Verify the current metrics before dropping the existing bucket
		metrics := c.Stats(metric)
		assert.Equal(t, count, metrics[metric])

		// The incremented metrics will be dropped at (capacity+1)*duration later
		time.Sleep(time.Duration(capacity+1) * duration)

		metrics = c.Stats(metric)
		assert.Equal(t, uint32(0), metrics[metric])
	})
}

func TestStats(t *testing.T) {
	metric1, metric2 := "test_1", "test_2"

	capacity, duration := 3, time.Second
	c, _ := NewTimeBucketCounter(capacity, duration)
	c.Increment(metric1)
	c.Increment(metric1)
	c.Increment(metric2)
	metrics := c.Stats(metric1, metric2)

	assert.Equal(t, uint32(2), metrics[metric1])
	assert.Equal(t, uint32(1), metrics[metric2])
}

func TestReset(t *testing.T) {
	count, metric := uint32(0), "test"

	capacity, duration := 3, time.Second
	c, _ := NewTimeBucketCounter(capacity, duration)
	c.Increment(metric)

	timer := c.timer
	c.Reset()

	assert.Equal(t, count, c.stats[metric])
	assert.Equal(t, count, c.buckets[2][metric])
	assert.NotEqual(t, timer, c.timer)
}
