# bird

![Regular Update](../regular_update.svg)

1. Check the latest version on the [download page](https://bird.nic.cz/download/) of the official website.
2. Check the `https://gitlab.nic.cz/labs/bird/-/blob/vX.Y.Z/NEWS` in the official GitLab repository at the specified version tag for release notes.
3. Check the release announcement from the [news page](https://bird.nic.cz/news/) on the official website.
4. Update base image in `Dockerfile`.
5. Update `BIRD_VERSION` and `BIRD_HASH` variables in `Dockerfile`.
6. Update image tag in `README.md`.
7. Update `BRANCH` and `TAG` files.
