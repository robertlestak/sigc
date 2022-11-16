#!/bin/bash
set -e

BUILDER=b-`uuidgen`

cleanup() {
  docker buildx rm $BUILDER
}

trap cleanup EXIT

docker buildx create --use --name $BUILDER
docker buildx inspect --bootstrap

IMAGE=registry.kypr.sh/sigc

IMAGE_INTERNAL=${IMAGE/registry.kypr.sh/registry.kypr.svc.cluster.local}

TAG=$GIT_COMMIT

bash /bin/build.sh . \
    -f devops/docker/Dockerfile \
    --platform linux/amd64 \
    --output type=image,name=$IMAGE_INTERNAL:$TAG,push=true,registry.insecure=true

sed -e "s,$IMAGE:.*,$IMAGE:$TAG,g" \
  devops/k8s/*.yaml | kubectl apply -f -
