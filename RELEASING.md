## Image release

### 1. Pick release tag

Use `X.Y.Z` or `vX.Y.Z` for stable releases. Use SemVer prerelease suffixes,
such as `vX.Y.Z-rc.1`, for release candidates. Build metadata (`+...`) is not
supported because Docker tags cannot contain `+`.

```bash
export T=v1.21.3
```

### 2. Create and push tag

Pushing tag triggers `.github/workflows/release-tag.yml`.

```bash
git tag "$T"
git push origin "$T"
```

### 3. Verify release

Confirm GitHub Actions run succeeds and image reaches Docker Hub. Stable tags
also update `latest`; prerelease tags do not.

## Registry configuration

- `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets are required.
- Optional `DOCKER_IMAGE_REPO` repository variable overrides default
  `docker.io/${DOCKER_USERNAME}/k8s-ephemeral-storage-metrics` target.

Helm chart releases happen separately in
[`densify-dev/helm-charts`](https://github.com/densify-dev/helm-charts/tree/master/charts/k8s-ephemeral-storage-metrics).
