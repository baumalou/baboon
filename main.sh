#!/bin/bash
export ENV=TEST
export TestMode=true
mkdir -p /go/src/git.workshop21.ch/${CI_PROJECT_PATH} 
cp -r ./* /go/src/git.workshop21.ch/${CI_PROJECT_PATH} 
git config --global credential.helper store
cp /.git-credentials ~/.git-credentials

# add custom deps first
IFS=',' read -r -a depArray <<< "$GODEP"
for dep in "${depArray[@]}"
do
   project_url="$(echo ${dep} | cut -d'@' -f1)"
   branchcommit="$(echo ${dep} | cut -d'@' -f2)"
   branch="$(echo ${branchcommit} | cut -d':' -f1)"
   commit="$(echo ${branchcommit} | cut -d':' -f2)"
   git clone --branch ${branch} ${project_url} /go/src/${project_url/*:\/\//}
   
   if [[ ${branch} != ${commit} ]]; then cd /go/src/${project_url/*:\/\//}; git reset --hard ${commit}; fi
done

cd /go/src/git.workshop21.ch/${CI_PROJECT_PATH}
go-wrapper download -t ./...


go build -o `pwd`/main

mkdir /${CI_PROJECT_DIR}/target
cp `pwd`/main /${CI_PROJECT_DIR}/target/main

# to remove - old version
cp `pwd`/main /${CI_PROJECT_DIR}/main

