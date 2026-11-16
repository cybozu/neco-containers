# pomerium-ingress-controller

Update this together with [pomerium](../pomerium/MAINTENANCE.md) and use the same version.

1. Check the release page.
   - <https://github.com/pomerium/ingress-controller/releases>
2. Check the diff of the Dockerfile and the Makefile.

   ```bash
   cd /path/to/ingress-controller
   git switch --detach v${NewVersion}
   git diff v${CurrentVersion} Dockerfile Makefile .tool-versions
   ```

3. Update `Dockerfile`.
   - `INGRESS_CONTROLLER_VERSION`, `INGRESS_CONTROLLER_CHECKSUM` and `INGRESS_CONTROLLER_COMMIT`
   - Golang version
   - Node.js version
4. Update `TAG` file.
