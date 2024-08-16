build_bin:
	GOOS=linux CGO_ENABLED=0 go build -o simple-chat .

build: build_bin 
	docker build -t simple-chat:latest .

lint:
	go fmt ./...

logs:
	kubectl logs -f deployment/simple-chat

redeploy: build
	kubectl delete deploy simple-chat --grace-period=0
	kubectl apply -f ./deploy/deployment.yml
