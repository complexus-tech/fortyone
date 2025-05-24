otel:
	docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

stop-otel:
	docker stop jaeger

dev:
	~/go/bin/air    

worker:
	go run cmd/worker/main.go

develop:
	go run cmd/api/main.go 

tidy:
	go mod tidy
