package store

var (
	fromAPIs map[string]string
	toAPI    string
)

func SetConfigFroms(froms map[string]string) {
	fromAPIs = froms
}

func SetConfigToAPI(token string) {
	toAPI = token
}

func GetConfigFromAPITokens() map[string]string {
	return fromAPIs
}

func GetConfigFromAPI(workspaceName string) (token string) {
	return fromAPIs[workspaceName]
}

func GetConfigToAPI() string {
	return toAPI
}
