package metrics

import "hash/maphash"

// CounterOpt are the options for creating a Counter.
type CounterOpt struct {
	Family Ident
	Tags   []Tag
}

// NewCounter registers and returns new Counter with the given name in the s.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewCounter("family", "label1", "value1", "label2", "value2")
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewCounter(family string, tags ...string) *Counter {
	return s.NewCounterOpt(CounterOpt{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
	})
}

// NewCounterOpt registers and returns new Counter with the opts in the s.
//
// The returned Counter is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewCounterOpt(opt CounterOpt) *Counter {
	c := &Counter{}
	s.mustRegisterMetric(c, opt.Family, opt.Tags)
	return c
}

// GetOrCreateCounter returns registered Counter in s with the given name
// and tags creates new Counter if s doesn't contain Counter with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateCounter("family", "label1", "value1", "label2", "value2")
//
// The returned Counter is safe to use from concurrent goroutines.
//
// Prefer [NewCounter] or [NewCounterOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateCounter(family string, tags ...string) *Counter {
	hash := getHashStrings(family, tags)

	s.mu.Lock()
	nm := s.metrics[hash]
	s.mu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Counter{}, hash, family, tags)
	}
	return nm.metric.(*Counter)
}

type CounterVecOpt struct {
	Family string
	Labels []string
}

type CounterVec struct {
	s           *Set
	family      Ident
	partialTags []Tag
	partialHash *maphash.Hash
}

func (c *CounterVec) WithLabelValues(values ...string) *Counter {
	hash := hashFinish(c.partialHash, values)

	c.s.mu.Lock()
	nm := c.s.metrics[hash]
	c.s.mu.Unlock()

	if nm == nil {
		nm = c.s.getOrRegisterMetricFromVec(
			&Counter{}, hash, c.family, c.partialTags, values,
		)
	}
	return nm.metric.(*Counter)
}

func (s *Set) NewCounterVec(opt CounterVecOpt) *CounterVec {
	family := MustIdent(opt.Family)

	// copy labels into partial tags. partial tags
	// have a validated label, but no value.
	partialTags := make([]Tag, len(opt.Labels))
	for i, label := range opt.Labels {
		partialTags[i].label = MustIdent(label)
	}

	return &CounterVec{
		s:           s,
		family:      family,
		partialTags: partialTags,
		partialHash: hashStart(family.String(), opt.Labels),
	}
}
