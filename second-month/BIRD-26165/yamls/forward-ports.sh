cat forward-ports.sh 
#!/bin/bash

while :
do
ps aux | awk '$0 ~ /port-forward/ { print $2}' | xargs kill -9
kubectl  port-forward svc/magento 80:80 -n magento-app &
sleep 300
kill %1 %2 %3 %4
sleep 1
done
