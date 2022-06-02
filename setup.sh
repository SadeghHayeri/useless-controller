#!/bin/bash
CLUSTER=kind

kind create cluster --name $CLUSTER
k ctx $CLUSTER

IMAGES=(
  "quay.io/jetstack/cert-manager-cainjector:v1.8.0"
  "quay.io/jetstack/cert-manager-controller:v1.8.0"
  "quay.io/jetstack/cert-manager-webhook:v1.8.0"
)

for image in "${IMAGES[@]}"; do
  echo "Pulling image: $image"
#  docker pull $image
  kind load docker-image $image --name $CLUSTER
done

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
