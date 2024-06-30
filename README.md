## How to run db for this project
`docker run -it --name some-postgres -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=user -e POSTGRES_DB=db -p 5432:5432 -e PGDATA=/var/lib/postgres/whatever postgres`

## Run redis
`docker run -d -it --name some-redis -e REDIS_PASSWORD=pass -e REDIS_PORT=6379 -v /data/redis:/data -p 6379:6379 redis`

//`docker run --name some-redis -p 6379:6379 redis` 