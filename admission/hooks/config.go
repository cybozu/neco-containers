package hooks

// Config is a config for neco-admission
type Config struct {
	ArgoCDApplicationValidatorConfig ArgoCDApplicationValidatorConfig `json:"ArgoCDApplicationValidator"`
}

// ArgoCDApplicationValidatorConfig is a config for application validator
type ArgoCDApplicationValidatorConfig struct {
	Rules []ArgoCDApplicationRule `json:"rules"`
}

// ArgoCDApplicationRule is a rule for applications
type ArgoCDApplicationRule struct {
	Repository string   `json:"repository"`
	Projects   []string `json:"projects"`
}
