Actions runner container
==============

Build Docker container image for [runner][], which enables us to run github actions workflow on our staging environment .

This Dockerfile is referencing to that from [actions-runner-controller](https://github.com/summerwind/actions-runner-controller/blob/584590e97c5c6d03e734fdf8b31b0f46227b4721/runner/Dockerfile)

`entrypoint_wrapper.sh` controls Pods' lifecycle to enable us to extend the Pods that failed CI.
1. Send slack notification to the Slack Agent
2. Check `/home/runner/delete-immediately` file and run the following procedure
   1. If the file exists, label its Runner resource with a deletion time. The deletion time is immediate. Then compare this label with the current time and delete the Runner if the time has passed.
   2. If the file not exist, remove the RunnerReplicaSet management label from its Runner resource by `kubectl label` command. After that, add a label to the Runner source to indicate the deletion time. The deletion time is about the current time + 30 minutes. Compare this label with the current time, and if it has passed, delete the Runner.

[Runner]: https://github.com/actions/runner

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/actions-runner)
