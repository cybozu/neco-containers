package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricInert = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_lending_inert",
		Help: "1 when the kubelet CPU manager policy is not static and lending is disabled.",
	})
	metricLendableCPUs = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_lending_lendable_cpus",
		Help: "Number of CPUs currently lent to borrowers.",
	})
	metricBorrowers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_lending_borrower_containers",
		Help: "Number of borrower containers tracked on this node.",
	})
	metricUnconverged = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_lending_unconverged",
		Help: "1 when at least one borrower container failed to converge to the desired cpuset. Alert when this stays 1.",
	})
	metricReconciles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cpu_lending_reconciles_total",
		Help: "Total number of reconcile runs.",
	})
	metricReconcileErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cpu_lending_reconcile_errors_total",
		Help: "Total number of failed reconcile runs.",
	})
	metricUpdates = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cpu_lending_container_updates_total",
		Help: "Total number of successful borrower cpuset updates issued by reconcile.",
	})
	metricAdvertised = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_lending_advertised_milli_cpu",
		Help: "Lending capacity currently advertised as the node extended resource, in milli-CPUs (1000 per lent CPU).",
	})
)
