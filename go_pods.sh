kubectl delete pod nr1
kubectl delete pod nr2
kubectl delete pod nr3
kubectl delete service conexion
echo "--------- Esperar un poco para dar tiempo que terminen Pods previos"
sleep 1
kubectl create -f pods_go.yaml
