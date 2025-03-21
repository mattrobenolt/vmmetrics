package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestCounterNew(t *testing.T) {
	NewSet().NewCounter("foo")
	NewSet().NewCounter("foo", "bar", "baz")

	// invalid label pairs
	assert.Panics(t, func() { NewSet().NewCounter("foo", "bar") })

	// duplicate
	set := NewSet()
	set.NewCounter("foo")
	assert.Panics(t, func() { set.NewCounter("foo") })
}

func TestCounterGetOrCreate(t *testing.T) {
	set := NewSet()
	set.GetOrCreateCounter("foo").Inc()
	set.GetOrCreateCounter("foo").Inc()
	assert.Equal(t, 2, set.GetOrCreateCounter("foo").Get())

	set.GetOrCreateCounter("foo", "a", "1").Inc()
	assert.Equal(t, 2, set.GetOrCreateCounter("foo").Get())
	assert.Equal(t, 1, set.GetOrCreateCounter("foo", "a", "1").Get())
}

func TestCounterVec(t *testing.T) {
	set := NewSet()
	c := set.NewCounterVec(CounterVecOpt{
		Family: "foo",
		Labels: []string{"a", "b"},
	})
	c.WithLabelValues("1", "2").Inc()
}

func TestCounterSerial(t *testing.T) {
	const name = "CounterSerial"
	set := NewSet()
	c := set.NewCounter(name)
	c.Inc()
	assert.Equal(t, c.Get(), 1)
	c.Dec()
	assert.Equal(t, c.Get(), 0)
	c.Set(123)
	assert.Equal(t, c.Get(), 123)
	c.Dec()
	assert.Equal(t, c.Get(), 122)
	c.Add(3)
	assert.Equal(t, c.Get(), 125)

	assertMarshal(t, set, []string{"CounterSerial 125"})
}

func TestCounterConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	c := NewSet().NewCounter("x")
	hammer(t, n, func() {
		nPrev := c.Get()
		for range inner {
			c.Inc()
			assert.Greater(t, c.Get(), nPrev)
		}
	})
	assert.Equal(t, c.Get(), n*inner)
}

func TestCounterGetOrCreateConcurrent(t *testing.T) {
	const n = 1000
	const inner = 10

	set := NewSet()
	fn := func() *Counter {
		return set.GetOrCreateCounter("x", "a", "1")
	}
	hammer(t, n, func() {
		nPrev := fn().Get()
		for range inner {
			fn().Inc()
			assert.Greater(t, fn().Get(), nPrev)
		}
	})
	assert.Equal(t, fn().Get(), n*inner)
}
