cd /app/log
for f in $(ls | grep _bw);  do mv $f ${f/.log*/_bw.log}; done
for f in $(ls | grep _lat);  do mv $f ${f/.log*/_lat.log}; done
for f in $(ls | grep _slat);  do mv $f ${f/.log*/_slat.log}; done
for f in $(ls | grep _clat);  do mv $f ${f/.log*/_clat.log}; done
for f in $(ls | grep _iops);  do mv $f ${f/.log*/_iops.log}; done
fio_generate_plots "$(date +%F_%T)"
mv *.svg /app/pictures
fio2gnuplot -p "*results_bw*" -g
mv *.png /app/pictures
rm /app/log/*