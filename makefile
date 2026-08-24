.PHONY: build run down clean
build: 
	@echo "Building short_url docker images.."
	docker build -t short_url:v1.0 .

run: build
	@echo "Starting container service"
	@docker compose up -d
	
	
down: 
	@echo "Stoping container service"
	@docker compose down 

#单独的清理数据目标，明确区分
clean: down
	@echo "Deleting all volumes (data will be lost)"
	@docker compose down -v

