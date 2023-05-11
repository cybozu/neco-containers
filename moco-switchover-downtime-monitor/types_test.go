package main

import "testing"

func TestPods(t *testing.T) {
	pod := Pod{
		Status: PodStatus{
			ContainerStatuses: []ContainerStatus{
				{
					Name:         "foo",
					RestartCount: 5,
				},
				{
					Name:         "bar",
					RestartCount: 10,
				},
			},
		},
	}

	if pod.RestartCount("foo") != 5 {
		t.Error("RestartCount of foo is not 5")
	}
	if pod.RestartCount("bar") != 10 {
		t.Error("RestartCount of bar is not 10")
	}
	if pod.RestartCount("baz") != 0 {
		t.Error("RestartCount of baz is not 0")
	}
}

func TestMySQLClusters(t *testing.T) {
	mysqlCluster := MySQLCluster{
		Status: MySQLClusterStatus{
			Conditions: []MySQLClusterCondition{
				{
					Type:   "Aaa",
					Status: "Alpha",
				},
				{
					Type:   "Healthy",
					Status: "True",
				},
				{
					Type:   "Available",
					Status: "False",
				},
				{
					Type:   "Zzz",
					Status: "Zulu",
				},
			},
		},
	}

	if !mysqlCluster.Healthy() {
		t.Error("Healthy() is not true")
	}
	if mysqlCluster.Available() {
		t.Error("Available() is not false")
	}

	mysqlCluster.Status.Conditions[1].Status = "False"
	mysqlCluster.Status.Conditions[2].Status = "True"

	if mysqlCluster.Healthy() {
		t.Error("Healthy() is not false")
	}
	if !mysqlCluster.Available() {
		t.Error("Available() is not true")
	}

	mysqlCluster.Status.Conditions[1].Status = "Foo"
	mysqlCluster.Status.Conditions[2].Status = "Bar"

	if mysqlCluster.Healthy() {
		t.Error("Healthy() is not false")
	}
	if mysqlCluster.Available() {
		t.Error("Available() is not false")
	}

	mysqlCluster.Status.Conditions = mysqlCluster.Status.Conditions[:1]

	if mysqlCluster.Healthy() {
		t.Error("Healthy() is not false")
	}
	if mysqlCluster.Available() {
		t.Error("Available() is not false")
	}
}
