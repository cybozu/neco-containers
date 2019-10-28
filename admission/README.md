[![Docker Repository on Quay](https://quay.io/repository/cybozu/neco-admission/status "Docker Repository on Quay")](https://quay.io/repository/cybozu/neco-admission)

neco-admission
==============

`neco-admission` is a custom [admission webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) for Neco.

It has the following webhooks / controllers.

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
