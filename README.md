# rookchaos

operator created to perform stresstests and resilience tests on rook cluster

## create rook-tool build container
```
docker login -u gitlab-runner -p ${NEXUSPASS} docker.workshop21.ch
docker build -f Dockerfile -t rook-build-container .
docker tag $(docker images -q rook-build-container) docker.workshop21.ch/boilerplate/build/rook-go:latest
docker push docker.workshop21.ch/boilerplate/build/rook-go:latest
```