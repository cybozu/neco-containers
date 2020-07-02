neco-admission
==============

`neco-admission` is a custom [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) for Neco.

It has the following webhooks / controllers.

DeleteValidator
---------------

This is to protect important resources from accidental deletion by human errors.

Every resource passed to this validation webhook will be denied for DELETE
unless it has this special annotation `i-am-sure-to-delete: <name of the resource>`.

ArgoCDApplicationValidator
--------------------------

ArgoCD's [Application resource](https://github.com/argoproj/argo-cd/blob/master/manifests/crds/application-crd.yaml)
can specify an [AppProject resource](https://github.com/argoproj/argo-cd/blob/master/manifests/crds/appproject-crd.yaml)
by a `spec.project` property.
The ability of the Application is restricted by the AppProject it belongs to.

Each application team should specify appropriate `spec.project`s for
their applications.
To enforce this, `ArgoCDApplicationValidator` validates Application resources.

See the [document](docs/configuration.md#argocdapplicationvalidator) for
the configuration of `ArgoCDApplicationValitor`.

CalicoNetworkPolicyValidator
----------------------------

Calico's [NetworkPolicy resource](https://docs.projectcalico.org/v3.10/reference/resources/networkpolicy) can have
[`order`](https://docs.projectcalico.org/v3.10/reference/resources/networkpolicy#spec) field.  Calico applies the
policy with the lowest value first.

To give priority to the system policies defined by Neco team while allowing
users to define NetworkPolicies within their namespaces, policies with
lower order value than threshold (`<= 1000` by default) are prohibited.

This webhook called `CalicoNetworkPolicyValidator` is a validating admission
webhook that denies such Calico NetworkPolicies.  The threshold value can be
changed per namespace by adding `admission.cybozu.com/min-policy-order`
annotation to the namespace.

NetworkPolicies w/o order field are permitted because they are applied last.

ContourHTTPProxyMutator / ContourHTTPProxyValidator
---------------------------------------------------

Contour's [HTTPProxy resource](https://projectcontour.io/docs/master/httpproxy/) and
[IngressRoute resource](https://projectcontour.io/docs/master/ingressroute/) can specify
the Ingress class that should interpret and serve the Ingress.
The [annotations](https://projectcontour.io/docs/master/annotations/)
`kubernetes.io/ingress.class` and `projectcontour.io/ingress.class` are used
for this specification.

Though the Contour documentation says that all Ingress controllers serve
the Ingress if the annotations are not set, this default behavior is dangerous.
It may cause unexpected disclosure of services which are intended only for
limited network.

The mutating webhook enforces the default annotation of `kubernetes.io/ingress.class: <configured value>` for `HTTPProxy` to prevent such accidents.
The default value can be configured with the `--httpproxy-default-class` option for `neco-admission`.

The validating webhook prevents creating `HTTPProxy` without the annotations, and prevents updating `HTTPProxy` to change the annotation values.

`neco-admission` does not watch `IngressRoute` because it is deprecated.

PodMutator
----------

PodMutator mutates Pod manifests to mount writable emptyDir to `/tmp` for each containers.
The purpose of this mutator is to prevent Pods from unexpected death by writing to read-only filesystem. 

However, Pods that already have another volumes mounted under `/tmp/*` are excluded from the mutating target.

GrafanaDashboardValidator
-------------------------

GrafanaDashboardValidator validates [GrafanaDashboard](https://github.com/integr8ly/grafana-operator/blob/v3.2.0/documentation/dashboards.md).

This validating webhook ensures the GrafanaDashboard resource's `spec.plugins` is empty.

The purpose of this validator is to avoid installing any plugins to production Grafana by tenants.

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/neco-admission)
