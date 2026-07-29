# Many Pods test chart

Test fixture used by scrape-driven eviction e2e coverage. Deploys configurable
BusyBox replicas; default is five.

```bash
helm upgrade --install many-pods ./tests/charts/many-pods \
  --namespace many-pods \
  --create-namespace
```

```bash
helm uninstall many-pods --namespace many-pods
```
