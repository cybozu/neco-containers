module github.com/cybozu/neco-containers/machines-endpoints

go 1.24.4

// cf. https://github.com/darccio/mergo/blob/master/README.md#100
replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.7

// due to repository move
replace github.com/armon/go-metrics v0.5.3 => github.com/hashicorp/go-metrics v0.5.3

require (
	github.com/cybozu-go/log v1.7.0
	github.com/cybozu-go/sabakan/v3 v3.1.5
	github.com/google/go-cmp v0.7.0
	github.com/hashicorp/serf v0.10.2
	github.com/spf13/pflag v1.0.6
	k8s.io/api v0.32.6
	k8s.io/apimachinery v0.32.6
	k8s.io/client-go v0.32.6
)

require (
	github.com/armon/go-metrics v0.5.3 // indirect
	github.com/cybozu-go/netutil v1.4.9 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.12.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/memberlist v0.5.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/miekg/dns v1.1.63 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/gomega v1.36.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/vishvananda/netlink v1.3.0 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/oauth2 v0.26.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241212222426-2c72e554b1e7 // indirect
	k8s.io/utils v0.0.0-20241210054802-24370beab758 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.5.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
