package metrics

// CounterOpt are the options for creating a Counter.
type GauageOpt struct {
	Family Ident
	Tags   []Tag
	// Func is an optional callback for making observations.
	Func func() float64
}

// NewGauge registers and returns gauge with the given name in s, which calls fn
// to obtain gauge value.
//
// family must be a Prometheus compatible identifier format.
//
// fn is an optional callback for making observations.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	NewGauge("family", observeFn, "label1", "value1", "label2", "value2")
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// This will panic if values are invalid or already registered.
func (s *Set) NewGauge(family string, fn func() float64, tags ...string) *Gauge {
	return s.NewGaugeOpt(GauageOpt{
		Family: MustIdent(family),
		Tags:   MustTags(tags...),
		Func:   fn,
	})
}

// NewGaugeOpt registers and returns new Gauge with the opts in the s.
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// This will panic if already registered.
func (s *Set) NewGaugeOpt(opt GauageOpt) *Gauge {
	g := &Gauge{fn: opt.Func}
	s.mustRegisterMetric(g, opt.Family, opt.Tags)
	return g
}

// GetOrCreateGauge returns registered gauge with the given name in s
// or creates new gauge if s doesn't contain gauge with the given name.
//
// family must be a Prometheus compatible identifier format.
//
// Optional tags must be specified in [label, value] pairs, for instance,
//
//	GetOrCreateGauge("family", "label1", "value1", "label2", "value2")
//
// The returned Gauge is safe to use from concurrent goroutines.
//
// Prefer [NewGauge] or [NewGaugeOpt] when performance is critical.
//
// This will panic if values are invalid.
func (s *Set) GetOrCreateGauge(family string, tags ...string) *Gauge {
	hash := getHashStrings(family, tags)

	s.mu.Lock()
	nm := s.metrics[hash]
	s.mu.Unlock()

	if nm == nil {
		nm = s.getOrAddMetricFromStrings(&Gauge{}, hash, family, tags)
	}
	return nm.metric.(*Gauge)
}
