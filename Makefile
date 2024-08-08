start-env:
	docker compose --profile dev up -d --wait

stop-env:
	docker compose --profile dev down -v

shell:
	docker compose exec sqidsencoder_env bash
