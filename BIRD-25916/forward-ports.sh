cat forward-ports.sh 
#!/bin/bash

while :
do
ps aux | awk '$0 ~ /port-forward/ { print $2}' | xargs kill -9
kubectl  port-forward svc/consul-consul-dns 8500:8500 &
kubectl  port-forward svc/consul-consul-dns 8502:8502 &
kubectl  port-forward svc/consul-consul-dns 8301:8301 &
sleep 300
kill %1 %2 %3 %4
sleep 1
done
