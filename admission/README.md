neco-admission
==============

`neco-admission` is a custom [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) for Neco.

It has the following webhooks / controllers.

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

If `VAPPLICATION_REPOSITORY_PERMISSIVE=true` envvar is set, this does not deny Applications but issues an warning.

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

Contour's [HTTPProxy resource](https://projectcontour.io/docs/main/config/fundamentals/) can specify
[the Ingress class](https://projectcontour.io/docs/main/config/ingress/) that should interpret and serve the Ingress.
The [annotations](https://projectcontour.io/docs/main/config/annotations/)
`kubernetes.io/ingress.class` and `projectcontour.io/ingress.class` in addition of `.spec.ingressClassName` field are used
for this specification.

Though the Contour documentation says that all Ingress controllers serve
the Ingress if neither the annotations nor the field are not set, this default behavior is dangerous.
It may cause unexpected disclosure of services which are intended only for
limited network.

The mutating webhook enforces the default ingress class of `.spec.ingressClassName=<configured value>` for `HTTPProxy` to prevent such accidents.
The default value can be configured with the `--httpproxy-default-class` option for `neco-admission`.

The validating webhook prevents creating `HTTPProxy` without the annotations nor the field, and prevents updating `HTTPProxy` to change the annotation values.

DeleteValidator
---------------

This is to protect important resources from accidental deletion by human errors.

Every resource passed to this validation webhook will be denied for DELETE
unless it has this special annotation `admission.cybozu.com/i-am-sure-to-delete: <name of the resource>`.

However, resources in namespaces that have `development: true` label can be deleted without the annotation.

DeploymentReplicaCountValidator
-------------------------------

This validator enforces that the number of replicas for a Deployment is 0 if the Deployment has a specific annotation.
This may be useful to prevent the number of replicas from being rewritten by ArgoCD or operators when the number of replicas is intentionally set to 0.

To enable this validator, annotate a Deployment with `admission.cybozu.com/force-replica-count: "0"`.

PreventDeleteValidator
----------------------

Unlike DeleteValidator, this prevents resources from accidental deletion only
if the resource is annotated with `admission.cybozu.com/prevent: delete`.

However, topolvm-controller can remove PersistentVolumeClaims, even if they are annotated with the above.

PodMutator
----------

PodMutator mutates Pod manifests to specify local ephemeral storage limit to 1GiB and request to 200MiB for each container.
The purpose of this mutator is to prevent Pods from overuse of local ephemeral storage.

If you want to use more ephemeral storage than the limit, you can use generic ephemeral volume instead of
local ephemeral storage.

PodValidator
------------

PodValidator validates Pod specifications as follows:

- Check that the container images have a valid prefix such as `quay.io/cybozu/`.
    - Valid prefixes are given through `--valid-image-prefix` command-line flags.
    - If `VPOD_IMAGE_PERMISSIVE=true` envvar is set, this does not deny Pods but issues an warning.

GrafanaDashboardValidator
-------------------------

GrafanaDashboardValidator validates [GrafanaDashboard](https://github.com/grafana-operator/grafana-operator/blob/v3.2.0/documentation/dashboards.md).

This validating webhook ensures the GrafanaDashboard resource's `spec.plugins` is empty.

The purpose of this validator is to avoid installing any plugins to production Grafana by tenants.

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/neco-admission)
