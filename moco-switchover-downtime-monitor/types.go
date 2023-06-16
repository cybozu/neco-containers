package main

// ObjectMeta is a subset of the k8s ObjectMeta definition.
type ObjectMeta struct {
	CreationTimestamp string `json:"creationTimestamp"`
}

// ContainerStatus is a subset of the k8s ContainerStatus definition.
type ContainerStatus struct {
	Name         string `json:"name"`
	RestartCount int    `json:"restartCount"`
}

// PodStatus is a subset of the k8s PodStatus definition.
type PodStatus struct {
	ContainerStatuses []ContainerStatus `json:"containerStatuses"`
}

// Pod is a subset of the k8s Pod definition.
type Pod struct {
	ObjectMeta `json:"metadata"`
	Status     PodStatus `json:"status"`
}

func (p *Pod) RestartCount(name string) int {
	for _, stat := range p.Status.ContainerStatuses {
		if stat.Name == name {
			return stat.RestartCount
		}
	}
	return 0
}

// MySQLClusterSpec is a subset of the moco MySQLClusterSpec definition.
type MySQLClusterSpec struct {
	Replicas int `json:"replicas"`
}

// MySQLClusterCondition is a subset of the moco MySQLClusterCondition definition.
type MySQLClusterCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// MySQLClusterStatus is a subset of the moco MySQLClusterStatus definition.
type MySQLClusterStatus struct {
	Conditions          []MySQLClusterCondition `json:"conditions"`
	CurrentPrimaryIndex int                     `json:"currentPrimaryIndex"`
}

// `MySQLCluster` is a subset of the moco MySQLCluster definition.
type MySQLCluster struct {
	Spec   MySQLClusterSpec   `json:"spec"`
	Status MySQLClusterStatus `json:"status"`
}

func (c *MySQLCluster) BoolCondition(name string, defaultValue bool) bool {
	for _, cond := range c.Status.Conditions {
		if cond.Type != name {
			continue
		}
		return cond.Status == "True"
	}
	return defaultValue
}

func (c *MySQLCluster) Healthy() bool {
	return c.BoolCondition("Healthy", false)
}

func (c *MySQLCluster) Available() bool {
	return c.BoolCondition("Available", false)
}
