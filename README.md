# rookchaos

operator created to perform stresstests and resilience tests on rook cluster

## create rook-tool build container
```
docker login -u gitlab-runner -p ${NEXUSPASS} docker.workshop21.ch
docker build -f Dockerfile.build -t rook-build-container .
docker tag $(docker images -q rook-build-container) docker.workshop21.ch/boilerplate/build/rook-go:latest
docker push docker.workshop21.ch/boilerplate/build/rook-go:latest
```



## storage class

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
   name: rook-test
provisioner: rook.io/block
parameters:
  pool: block-repl-low
```