kubectl delete deploy/de-1
kubectl delete deploy/de-2
kubectl delete deploy/de-3
kubectl delete service deploy-1
kubectl delete service deploy-2
kubectl delete service deploy-3
kubectl create -f deploy_go.yaml
