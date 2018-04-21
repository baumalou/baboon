SIZE=${SIZE:-16k} DIR=/testvol DATE=$(date +%F_%T) fio /app/fio/conf/rookeval-rand-io.fio
