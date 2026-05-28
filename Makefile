.PHONY: build dashboard dev test

build: dashboard
	go build -o bin/eventhook ./cmd/eventhook

dashboard:
	cd dashboard && npm run build
	rm -rf assets/dashboard
	mkdir -p assets/dashboard
	cp -r dashboard/dist/* assets/dashboard/

dev:
	docker-compose up

test:
	go test ./...
