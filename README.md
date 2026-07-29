# K8s Ephemeral Storage Metrics

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/densify-dev/k8s-ephemeral-storage-metrics/actions/workflows/test.yaml/badge.svg)](https://github.com/densify-dev/k8s-ephemeral-storage-metrics/actions/workflows/test.yaml)

Prometheus exporter for Kubernetes pod, container, node, and volume ephemeral storage metrics.

Project addresses missing ephemeral-storage monitoring described in
[kubernetes/kubernetes#69507](https://github.com/kubernetes/kubernetes/issues/69507).
CSI-backed ephemeral storage, including
[generic ephemeral volumes](https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes),
is not monitored.

## Metrics

Exporter provides:

- Node available, capacity, and usage percentage
- Pod byte and inode usage
- Container rootfs/log byte, percentage, and inode usage
- Container `emptyDir` volume usage and limit percentage
- Container ephemeral-storage limit percentage

Every metric includes `node_name`. Pod and container metrics also include
`pod_name` and `pod_namespace`; container metrics include `container`.
Volume metrics add `volume_name` and `mount_path`.

Upstream-compatible metric names and labels remain unchanged. Scrape-driven
cleanup removes pod series after `SCRAPE_MISS_TOLERANCE` successful scrapes
without that pod; default is `2`.

## Helm install

Helm chart lives in
[`densify-dev/helm-charts`](https://github.com/densify-dev/helm-charts/tree/master/charts/k8s-ephemeral-storage-metrics).
Chart values and release notes there are source of truth.

```bash
helm repo add kubex https://densify-dev.github.io/helm-charts
helm repo update
helm upgrade --install my-deployment kubex/k8s-ephemeral-storage-metrics
```

## Development

Requires Go version declared in [`go.mod`](go.mod).

```bash
make fmt
make vet
make test-unit
```

Cluster-based Helm and e2e work belongs with external chart repository.

## Automated image release

Pushing SemVer tag (`X.Y.Z` or `vX.Y.Z`) runs
[`.github/workflows/release-tag.yml`](.github/workflows/release-tag.yml) and
publishes Docker Hub images. Stable tags also update `latest`; prerelease tags
do not.

Required repository secrets:

- `DOCKER_USERNAME`
- `DOCKER_PASSWORD`

Optional `DOCKER_IMAGE_REPO` repository variable overrides default
`docker.io/${DOCKER_USERNAME}/k8s-ephemeral-storage-metrics` target.
See [`RELEASING.md`](RELEASING.md).

## License

[MIT](LICENSE)
