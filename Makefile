build: 
	docker build -t simple-chat:latest .

up:
	mkdir -p postgres_data && docker compose up
down:
	docker compose down
lint:
	cd ./back/messaging/ && go fmt ./...

logs:
	kubectl logs -f deployment/simple-chat

redeploy: build
	kubectl delete deploy simple-chat --grace-period=0
	kubectl apply -f ./deploy/deployment.yml
