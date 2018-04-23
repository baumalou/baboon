export SIZE=${1:-32k} && export DIR=/testvol && export DATE=$(date +%F_%T) && fio /app/fio/conf/rookeval-rand-io.fio
