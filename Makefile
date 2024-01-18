cluster:
	@kind create cluster

docker-img:
	@cd node-app; docker build -t node-app:v1 .
	@cd go-app; docker build -t go-app:v1 .


docker-img-to-cluster:
	@kind load docker-image node-app:v1
	@kind load docker-image go-app:v1

set-pods:
	@cd node-app/k8s ; kubectl apply -f deployment.yaml ; kubectl apply -f service.yaml
	@cd go-app/k8s ; kubectl apply -f configs.yaml; kubectl apply -f deployment.yaml ; kubectl apply -f service.yaml

port-forward:
	@kubectl port-forward svc/go-app 8081:8081 &
	@kubectl port-forward svc/node-app 8080:8080 &

unset-pods:
	@cd node-app/k8s ; kubectl delete -f deployment.yaml ; kubectl delete -f service.yaml
	@cd go-app/k8s ; kubectl delete -f configs.yaml; kubectl delete -f deployment.yaml ; kubectl delete -f service.yaml

drop-docker-img:
	@docker rmi -f node-app:v1
	@docker rmi -f go-app:v1