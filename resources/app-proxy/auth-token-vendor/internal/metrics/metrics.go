package metrics

import (
	"github.com/go-chi/telemetry"
)

type AppMetrics struct {
	*telemetry.Scope
}

func (m *AppMetrics) RecordAppHit(label string) {
	m.RecordHit(label, nil)
}

func (m *AppMetrics) RecordAppGauge(label string, value float64) {
	m.RecordGauge(label, nil, value)
}

func NewAppMetrics(scope string) (am *AppMetrics) {
	am = &AppMetrics{telemetry.NewScope(scope)}
	return am
}
