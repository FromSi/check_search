.PHONY: install
install:
	- docker container rm postgres_check_search redisearch_check_search meilisearch_check_search
	- docker volume rm pgdata_volume_check_search redisdata_volume_check_search meilidata_volume_check_search
	docker compose down --rmi all
	docker compose up --build --force-recreate

.PHONY: run
run:
	docker compose up postgres redisearch meilisearch

.PHONY: stop
stop:
	docker compose stop

.PHONY: redisearch
redisearch:
	docker compose exec redisearch sh -c "redis-cli"
