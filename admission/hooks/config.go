package hooks

// Config is a config for neco-admission
type Config struct {
	HTTPProxyMutatorConfig           HTTPProxyMutatorConfig           `json:"HTTPProxyMutatorConfig"`
	NamespaceDeletionValidatorConfig NamespaceDeletionValidatorConfig `json:"NamespaceDeletionValidator"`
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
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}
