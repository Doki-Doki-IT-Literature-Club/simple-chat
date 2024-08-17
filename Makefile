build: 
	docker build -t simple-chat:latest .

run:
	docker-compose up

lint:
	go fmt ./...

logs:
	kubectl logs -f deployment/simple-chat

redeploy: build
	kubectl delete deploy simple-chat --grace-period=0
	kubectl apply -f ./deploy/deployment.yml
