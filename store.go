package statsig

import (
	"encoding/json"
	"strconv"
	"time"
)

type configSpec struct {
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Salt         string          `json:"salt"`
	Enabled      bool            `json:"enabled"`
	Rules        []configRule    `json:"rules"`
	DefaultValue json.RawMessage `json:"defaultValue"`
}

type configRule struct {
	Name           string            `json:"name"`
	ID             string            `json:"id"`
	Salt           string            `json:"salt"`
	PassPercentage float64           `json:"passPercentage"`
	Conditions     []configCondition `json:"conditions"`
	ReturnValue    json.RawMessage   `json:"returnValue"`
}

type configCondition struct {
	Type             string                 `json:"type"`
	Operator         string                 `json:"operator"`
	Field            string                 `json:"field"`
	TargetValue      interface{}            `json:"targetValue"`
	AdditionalValues map[string]interface{} `json:"additionalValues"`
}

type downloadConfigSpecResponse struct {
	HasUpdates     bool         `json:"has_updates"`
	Time           int64        `json:"time"`
	FeatureGates   []configSpec `json:"feature_gates"`
	DynamicConfigs []configSpec `json:"dynamic_configs"`
}

type downloadConfigsInput struct {
	SinceTime       string   `json:"sinceTime"`
	StatsigMetadata metadata `json:"statsigMetadata"`
}

type store struct {
	FeatureGates   map[string]configSpec
	DynamicConfigs map[string]configSpec
	lastSyncTime   int64
	transport      *transport
	ticker         *time.Ticker
}

func newStore(transport *transport) *store {
	store := &store{
		FeatureGates:   make(map[string]configSpec),
		DynamicConfigs: make(map[string]configSpec),
		transport:      transport,
		ticker:         time.NewTicker(10 * time.Second),
	}

	specs := store.fetchConfigSpecs()
	store.update(specs)
	go store.pollForChanges()
	return store
}

func (s *store) StopPolling() {
	s.ticker.Stop()
}

func (s *store) update(specs downloadConfigSpecResponse) {
	if specs.HasUpdates {
		newGates := make(map[string]configSpec)
		for _, gate := range specs.FeatureGates {
			newGates[gate.Name] = gate
		}

		newConfigs := make(map[string]configSpec)
		for _, config := range specs.DynamicConfigs {
			newConfigs[config.Name] = config
		}

		s.FeatureGates = newGates
		s.DynamicConfigs = newConfigs
	}
}

func (s *store) fetchConfigSpecs() downloadConfigSpecResponse {
	input := &downloadConfigsInput{
		SinceTime:       strconv.FormatInt(s.lastSyncTime, 10),
		StatsigMetadata: s.transport.metadata,
	}
	var specs downloadConfigSpecResponse
	s.transport.postRequest("/download_config_specs", input, &specs)
	s.lastSyncTime = specs.Time
	return specs
}

func (s *store) pollForChanges() {
	for range s.ticker.C {
		specs := s.fetchConfigSpecs()
		s.update(specs)
	}
}
