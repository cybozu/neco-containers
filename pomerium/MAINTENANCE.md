# pomerium

1. Check the release page and upgrade guide.
   - <https://github.com/pomerium/pomerium/releases>
   - <https://www.pomerium.com/docs/core/upgrading>
2. Check the diff of the Dockerfile.

   ```bash
   cd /path/to/pomerium
   git switch --detach v${NewVersion}
   git diff v${CurrentVersion} Dockerfile
   ```

3. Update `Dockerfile`.
   - Pomerium version
   - Golang version
   - Node.js version
4. Update `TAG` file.
