package statsig

// Wrappers are used to initialize clients from other languages. These should be
// used by Go clients.

func WrapperSDK(sdkKey string, options *Options, sdkName string, sdkVersion string) {
	once.Do(func() {
		instance = newClientWithOptionsAndMetadata(sdkKey, options, sdkName, sdkVersion)
	})
}

func WrapperSDKInstance(sdkKey string, options *Options, sdkName string, sdkVersion string) *Client {
	return newClientWithOptionsAndMetadata(sdkKey, options, sdkName, sdkVersion)
}
