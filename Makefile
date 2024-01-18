cluster:
	@kind create cluster

run-node-app:
	@cd node-app; npm run dev

docker-img:
	@cd node-app; docker build -t node-app:v1 .

docker-img-to-cluster:
	@kind load docker-image node-app:v1

set-pods:
	@cd node-app/k8s ; kubectl apply -f deployment.yaml ; kubectl apply -f service.yaml

