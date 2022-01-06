kubectl delete pod raft1
kubectl delete pod raft2
kubectl delete pod raft3
kubectl delete service conexion
echo "--------- Esperar un poco para dar tiempo que terminen Pods previos"
sleep 1
kubectl create -f pods_go.yaml
