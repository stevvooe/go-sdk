package statsig

import (
	"fmt"
	"strings"
)

// An instance of a StatsigClient for interfacing with Statsig Feature Gates, Dynamic Configs, Experiments, and Event Logging
type Client struct {
	sdkKey    string
	evaluator *evaluator
	logger    *logger
	transport *transport
	options   *Options
}

// Initializes a Statsig Client with the given sdkKey
func NewClient(sdkKey string) *Client {
	return NewClientWithOptions(sdkKey, &Options{API: DefaultEndpoint})
}

// Advanced options for configuring the Statsig SDK
type Options struct {
	API         string      `json:"api"`
	Environment Environment `json:"environment"`
}

func (o *Options) defaults() *Options { // allows call on a nil value.
	return &Options{API: DefaultEndpoint}
}

// Initializes a Statsig Client with the given sdkKey and options
func NewClientWithOptions(sdkKey string, options *Options) *Client {
	return WrapperSDKInstance(sdkKey, options, "", "")
}

func newClientWithOptionsAndMetadata(sdkKey string, options *Options, sdkName, sdkVersion string) *Client {
	if len(options.API) == 0 {
		options.API = "https://api.statsig.com/v1"
	}
	transport := newTransport(sdkKey, options.API, sdkName, sdkVersion)
	logger := newLogger(transport)
	evaluator := newEvaluator(transport)
	if !strings.HasPrefix(sdkKey, "secret") {
		panic("Must provide a valid SDK key.")
	}
	return &Client{
		sdkKey:    sdkKey,
		evaluator: evaluator,
		logger:    logger,
		transport: transport,
		options:   options,
	}
}

// Checks the value of a Feature Gate for the given user
func (c *Client) CheckGate(user User, gate string) bool {
	if user.UserID == "" {
		fmt.Println("A non-empty StatsigUser.UserID is required. See https://docs.statsig.com/messages/serverRequiredUserID")
		return false
	}
	user = normalizeUser(user, *c.options)
	res := c.evaluator.CheckGate(user, gate)
	if res.FetchFromServer {
		serverRes := fetchGate(user, gate, c.transport)
		res = &evalResult{Pass: serverRes.Value, Id: serverRes.RuleID}
	}
	c.logger.LogGateExposure(user, gate, res.Pass, res.Id)
	return res.Pass
}

// Gets the DynamicConfig value for the given user
func (c *Client) GetConfig(user User, config string) DynamicConfig {
	if user.UserID == "" {
		fmt.Println("A non-empty StatsigUser.UserID is required. See https://docs.statsig.com/messages/serverRequiredUserID")
		return *NewConfig(config, nil, "")
	}
	user = normalizeUser(user, *c.options)
	res := c.evaluator.GetConfig(user, config)
	if res.FetchFromServer {
		serverRes := fetchConfig(user, config, c.transport)
		res = &evalResult{
			ConfigValue: *NewConfig(config, serverRes.Value, serverRes.RuleID),
			Id:          serverRes.RuleID}
	}
	c.logger.LogConfigExposure(user, config, res.Id)
	return res.ConfigValue
}

// Gets the DynamicConfig value of an Experiment for the given user
func (c *Client) GetExperiment(user User, experiment string) DynamicConfig {
	if user.UserID == "" {
		fmt.Println("A non-empty StatsigUser.UserID is required. See https://docs.statsig.com/messages/serverRequiredUserID")
		return *NewConfig(experiment, nil, "")
	}
	return c.GetConfig(user, experiment)
}

// Logs an event to Statsig for analysis in the Statsig Console
func (c *Client) LogEvent(event Event) {
	event.User = normalizeUser(event.User, *c.options)
	if event.EventName == "" {
		return
	}
	c.logger.Log(event)
}

// Cleans up Statsig, persisting any Event Logs and cleanup processes
// Using any method is undefined after Shutdown() has been called
func (c *Client) Shutdown() {
	c.logger.Flush(true)
	c.evaluator.Stop()
}

type gateResponse struct {
	Name   string `json:"name"`
	Value  bool   `json:"value"`
	RuleID string `json:"rule_id"`
}

type configResponse struct {
	Name   string                 `json:"name"`
	Value  map[string]interface{} `json:"value"`
	RuleID string                 `json:"rule_id"`
}

type checkGateInput struct {
	GateName        string   `json:"gateName"`
	User            User     `json:"user"`
	StatsigMetadata metadata `json:"statsigMetadata"`
}

type getConfigInput struct {
	ConfigName      string   `json:"configName"`
	User            User     `json:"user"`
	StatsigMetadata metadata `json:"statsigMetadata"`
}

func fetchGate(user User, gateName string, t *transport) gateResponse {
	input := &checkGateInput{
		GateName:        gateName,
		User:            user,
		StatsigMetadata: t.metadata,
	}
	var res gateResponse
	err := t.postRequest("/check_gate", input, &res)
	if err != nil {
		return gateResponse{
			Name:   gateName,
			Value:  false,
			RuleID: "",
		}
	}
	return res
}

func fetchConfig(user User, configName string, t *transport) configResponse {
	input := &getConfigInput{
		ConfigName:      configName,
		User:            user,
		StatsigMetadata: t.metadata,
	}
	var res configResponse
	err := t.postRequest("/get_config", input, &res)
	if err != nil {
		return configResponse{
			Name:   configName,
			RuleID: "",
		}
	}
	return res
}

func normalizeUser(user User, options Options) User {
	var env map[string]string
	if len(options.Environment.Params) > 0 {
		env = options.Environment.Params
	} else {
		env = make(map[string]string)
	}

	if options.Environment.Tier != "" {
		env["tier"] = options.Environment.Tier
	}
	for k, v := range user.StatsigEnvironment {
		env[k] = v
	}
	user.StatsigEnvironment = env
	return user
}
