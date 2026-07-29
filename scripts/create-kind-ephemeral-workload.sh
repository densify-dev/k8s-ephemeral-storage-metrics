#!/usr/bin/env bash

set -euo pipefail

cluster_name="${1:-ephemeral-storage}"

command -v kind >/dev/null || { echo "kind is required" >&2; exit 1; }
command -v kubectl >/dev/null || { echo "kubectl is required" >&2; exit 1; }

kind create cluster --name "$cluster_name" --config=- <<'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
  - role: worker
EOF

kubectl apply -f - <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ephemeral-storage-writer
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ephemeral-storage-writer
  template:
    metadata:
      labels:
        app: ephemeral-storage-writer
    spec:
      containers:
        - name: writer
          image: busybox:1.36
          command: ["sh", "-c"]
          args:
            - while true; do dd if=/dev/zero of=/data/$(date +%s) bs=1M count=10; sleep 60; done
          resources:
            requests:
              ephemeral-storage: 100Mi
            limits:
              ephemeral-storage: 500Mi
          volumeMounts:
            - name: ephemeral-data
              mountPath: /data
      volumes:
        - name: ephemeral-data
          emptyDir:
            sizeLimit: 500Mi
EOF

kubectl rollout status deployment/ephemeral-storage-writer --timeout=120s
kubectl get nodes
kubectl get pods -l app=ephemeral-storage-writer -o wide
