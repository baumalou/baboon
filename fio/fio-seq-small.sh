SIZE=${1:-4k} DIR=/testvol DATE=$(date +%F_%T) fio /app/fio/conf/rookeval-seq-io.fio
