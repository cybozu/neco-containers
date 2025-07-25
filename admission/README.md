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
the configuration of `ArgoCDApplicationValidator`.

If `VAPPLICATION_REPOSITORY_PERMISSIVE=true` envvar is set, this does not deny Applications but issues an warning.

ContourHTTPProxyMutator / ContourHTTPProxyValidator
---------------------------------------------------

### Ingress Class Name

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

### IP Allow Filter Policy

The virtual host route set for the HTTPProxy resource has a setting to allow matching requests.
The mutating webhook set this filtering policy by the annotation values with a predetermined value.

The validating webhook prevents updating `HTTPProxy` to change the annotation values. Also, this webhook allows updates to add annotations from an unannotated state.

See the [document](docs/configuration.md#httpproxymutator) for
the configuration of `HTTPProxyMutator`.

DeleteValidator
---------------

This is to protect important resources from accidental deletion by human errors.

Every resource passed to this validation webhook will be denied for DELETE
unless it has this special annotation `admission.cybozu.com/i-am-sure-to-delete: <name of the resource>`.

However, resources in namespaces that have `development: true` label can be deleted without the annotation.

DeploymentReplicaCountValidator / DeploymentReplicaCountScaleValidator
----------------------------------------------------------------------

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

PodMutator mutates Pod manifests to specify local ephemeral storage limit to 1GiB and request to 10MiB for each container.
The purpose of this mutator is to prevent Pods from overuse of local ephemeral storage.

If `VPOD_EPHEMERAL_STORAGE_PERMISSIVE=true` envvar is set, local ephemeral storage requests and limits specified in 
Pod manifests will not be overwritten.
If you want to use more ephemeral storage than the limit, you can use generic ephemeral volume instead of
local ephemeral storage.

PodCPURequestReducer
--------------------

PodCPURequestReducer mutates Pod manifests to reduce CPU requests by half for each container.
This is for deploying Pods in environments with a limited number of available CPUs, such as a test environment, without modifying the workload resources.
This mutator is enabled only when the `VPOD_CPU_REQUEST_REDUCE_ENABLE=true` envvar is set.

This mutator does not affect Pods created by DaemonSets or Pods with the `admission.cybozu.com/prevent-cpu-request-reduce: "true"` label.

PodValidator
------------

PodValidator validates Pod specifications as follows:

- Check that the container images have a valid prefix such as `ghcr.io/cybozu/`.
    - Valid prefixes are given through `--valid-image-prefix` command-line flags.
    - If `VPOD_IMAGE_PERMISSIVE=true` envvar is set, this does not deny Pods but issues an warning.

GrafanaDashboardValidator
-------------------------

GrafanaDashboardValidator validates [GrafanaDashboard](https://grafana-operator.github.io/grafana-operator/docs/api/#grafanadashboard).

This validating webhook ensures the GrafanaDashboard resource's `spec.plugins` is empty.

The purpose of this validator is to avoid installing any plugins to production Grafana by tenants.

Docker images
-------------

Docker images are available on [ghcr.io](https://github.com/cybozu/neco-containers/pkgs/container/neco-admission)
