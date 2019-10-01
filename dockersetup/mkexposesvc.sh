kubectl expose deployment todolists-deployment --type=LoadBalancer --name=todolist-service --port=80 --target-port=8008
kubectl describe service/todolist-service

