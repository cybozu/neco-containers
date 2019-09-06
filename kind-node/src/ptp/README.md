ptp plugin customized for Calico
================================

This is the customized [ptp plugin](https://github.com/containernetworking/plugins/tree/master/plugins/main/ptp) for Calico.
(ptp plugin is licensed under the Apache License, Version 2.0)

This plugin is modified to attach host veth name according to the Calico's naming rule. That is:
```go
func generateHostVethName(prefix, namespace, podname string) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%s.%s", namespace, podname)))
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(h.Sum(nil))[:11])
}
```
