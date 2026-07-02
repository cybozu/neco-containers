package hooks

// Config is a config for neco-admission
type Config struct {
	ArgoCDApplicationValidatorConfig ArgoCDApplicationValidatorConfig `json:"ArgoCDApplicationValidator"`
	HTTPProxyMutatorConfig           HTTPProxyMutatorConfig           `json:"HTTPProxyMutatorConfig"`
	NamespaceDeletionValidatorConfig NamespaceDeletionValidatorConfig `json:"NamespaceDeletionValidator"`
}

// ArgoCDApplicationValidatorConfig is a config for application validator
type ArgoCDApplicationValidatorConfig struct {
	Rules []ArgoCDApplicationRule `json:"rules"`
}

// ArgoCDApplicationRule is a rule for applications
type ArgoCDApplicationRule struct {
	Repository       string   `json:"repository"`
	RepositoryPrefix string   `json:"repositoryPrefix"`
	Projects         []string `json:"projects"`
}

type HTTPProxyMutatorConfig struct {
	Policies []HTTPProxyPolicy `json:"policies"`
}

type HTTPProxyPolicy struct {
	Name          string                    `json:"name"`
	IpAllowPolicy []HTTPProxyIPFilterPolicy `json:"ipAllowPolicy"`
}

type HTTPProxyIPFilterPolicy struct {
	Source string `json:"source"`
	Cidr   string `json:"cidr"`
}

type NamespaceDeletionValidatorConfig struct {
	ProtectedResources []NamespaceDeletionProtectedResource `json:"protectedResources"`
}

type NamespaceDeletionProtectedResource struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
