# rookchaos

operator created to perform stresstests and resilience tests on rook cluster

## create rook-tool build container
```
docker login -u gitlab-runner -p ${NEXUSPASS} docker.workshop21.ch
docker build -f Dockerfile.build -t rook-build-container .
docker tag $(docker images -q rook-build-container) docker.workshop21.ch/boilerplate/build/rook-go:latest
docker push docker.workshop21.ch/boilerplate/build/rook-go:latest
```

rook zerschoss crd k8s kei crd manipulatione me zuelah. dex het kei manipulatione an crd mer zuugelassen.
crd von rook gelöschd

neuer cluster deployed mit neuem operator
nur device sdb auf 4-6 mit blustore --> migrationspfad selector auf disk
neu repl2 auf pool
kei fragmentation
von hand eingetragen / regex 

aktuell master version von rook 0.7.0.XX


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