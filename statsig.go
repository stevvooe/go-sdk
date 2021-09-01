// Package statsig implements feature gating and a/b testing
package statsig

import (
	"fmt"
	"sync"
)

const DefaultEndpoint = "https://api.statsig.com/v1"

var (
	instance *Client
	once     sync.Once
)

// Initializes the global Statsig instance with the given sdkKey
func Initialize(sdkKey string) {
	var o *Options
	InitializeWithOptions(sdkKey, o.defaults())

}

// Initializes the global Statsig instance with the given sdkKey and options
func InitializeWithOptions(sdkKey string, options *Options) {
	once.Do(func() {
		instance = NewClientWithOptions(sdkKey, options)
	})
}

// Checks the value of a Feature Gate for the given user
func CheckGate(user User, gate string) bool {
	if instance == nil {
		panic(fmt.Errorf("must Initialize() statsig before calling CheckGate"))
	}
	return instance.CheckGate(user, gate)
}

// Gets the DynamicConfig value for the given user
func GetConfig(user User, config string) DynamicConfig {
	if instance == nil {
		panic(fmt.Errorf("must Initialize() statsig before calling GetConfig"))
	}
	return instance.GetConfig(user, config)
}

// Gets the DynamicConfig value of an Experiment for the given user
func GetExperiment(user User, experiment string) DynamicConfig {
	if instance == nil {
		panic(fmt.Errorf("must Initialize() statsig before calling GetExperiment"))
	}
	return instance.GetExperiment(user, experiment)
}

// Logs an event to the Statsig console
func LogEvent(event Event) {
	if instance == nil {
		panic(fmt.Errorf("must Initialize() statsig before calling LogEvent"))
	}
	instance.LogEvent(event)
}

// Cleans up Statsig, persisting any Event Logs and cleanup processes
// Using any method is undefined after Shutdown() has been called
func Shutdown() {
	if instance == nil {
		return
	}
	instance.Shutdown()
}
