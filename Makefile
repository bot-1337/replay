.DEFAULT_GOAL := help

help: # automatically documents the makefile, by outputing everything behind a ##
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.init: # check that dependencies are installed
	@./scripts/check_docker.sh
	@./scripts/check_docker_compose.sh

clean: .init ## ♻️  Cleanup dev environment
	# docker resources
	-docker stop $(shell docker ps -a -q)
	docker system prune -a -f
	# data assets
	rm -rf /tmp/ehub_data

dev: .init ## 🛠  Setup dev environment
	# possible future TODO: a docker image with this data already loaded in
	mkdir -p /tmp/ehub_data
	aws s3 cp s3://net.energyhub.assets/public/dev-exercises/audit-data.tar.gz /tmp/ehub_data/audit-data.tar.gz
	tar --extract --file /tmp/ehub_data/audit-data.tar.gz -C /tmp/ehub_data
	rm /tmp/ehub_data/audit-data.tar.gz

test: .init ## 📝 Run tests
	docker-compose run project go test -v ./...

build: .init ## 🛠 Build the CLI locally
	docker-compose run project env GOOS=darwin GOARCH=amd64 go build -o replay main.go
