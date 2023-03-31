TOOLS_DIR := $(CURDIR)/../bin
HELM := $(TOOLS_DIR)/helm
KIND := $(TOOLS_DIR)/kind
KUBECTL := $(TOOLS_DIR)/kubectl
YQ := $(TOOLS_DIR)/yq

.PHONY: setup
setup:
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) repo add jetstack https://charts.jetstack.io
	$(HELM) repo update bitnami
	$(HELM) repo update jetstack

.PHONY: start
start:
	mkdir -p tmp
	cd ..; docker build . --no-cache -t necosen:dev

	$(KIND) create cluster --name necosen
	$(KIND) load docker-image necosen:dev --name necosen
	$(HELM) install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true --set startupapicheck.enabled=false --wait
	$(HELM) install contour bitnami/contour --namespace contour --create-namespace --wait \
		--set envoy.service.type=ClusterIP \
		--set envoy.service.externalTrafficPolicy=""
	$(KUBECTL) apply -f issuer.yaml
	$(KUBECTL) apply -f auth.yaml
	$(KUBECTL) apply -f auth-free.yaml
	$(KUBECTL) apply -f service.yaml
	$(KUBECTL) apply -f client.yaml

	$(KUBECTL) wait --for=condition=Available=true deploy/client
	AUTHORIZED_IP=$$($(KUBECTL) get pod -l app.kubernetes.io/name=client -o yaml | $(YQ) '.items[0].status.podIP'); \
	$(YQ) -n ".sourceIP.allowedCIDRs=[\"$${AUTHORIZED_IP}/32\"]" > tmp/config.yaml
	$(KUBECTL) create cm -n contour necosen-config --from-file=tmp/config.yaml

	$(KUBECTL) wait -A --all deployments --for=condition=Available=true --timeout=300s

.PHONY: test
test:
	go test -v -race . -ginkgo.v -ginkgo.fail-fast

.PHONY: stop
stop:
	$(KIND) delete cluster --name necosen
