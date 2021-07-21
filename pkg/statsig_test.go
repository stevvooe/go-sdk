package statsig

import (
	"fmt"
	"os"
	"statsig/pkg/types"
	"testing"
)

type data struct {
	Entries []entry `json:"data"`
}

type entry struct {
	User    types.StatsigUser              `json:"user"`
	Gates   map[string]bool                `json:"feature_gates"`
	Configs map[string]types.DynamicConfig `json:"dynamic_configs"`
}

var secret string
var testAPIs = []string{
	"https://api.statsig.com/v1",
	"https://us-west-2.api.statsig.com/v1",
	"https://us-east-2.api.statsig.com/v1",
	"https://ap-south-1.api.statsig.com/v1",
	"https://latest.api.statsig.com/v1",
}

func TestMain(m *testing.M) {
	secret = "secret-9IWfdzNwExEYHEW4YfOQcFZ4xreZyFkbOXHaNbPsMwW"
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	for _, api := range testAPIs {
		fmt.Println("Testing for " + api)
		test_helper(api, t)
	}
}

func test_helper(apiOverride string, t *testing.T) {
	fmt.Println("Testing for " + apiOverride)
	c := NewWithOptions(secret, &types.StatsigOptions{API: apiOverride})
	var d data
	err := c.net.PostRequest("/rulesets_e2e_test", nil, &d)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, entry := range d.Entries {
		u := entry.User
		for gate, value := range entry.Gates {
			sdkV := c.CheckGate(u, gate)
			if sdkV != value {
				t.Errorf("%s failed for user %s: expected %t, got %t", gate, u, value, sdkV)
			}
		}
	}
}
