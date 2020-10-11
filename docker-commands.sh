# create network + subnet
docker network create -d bridge --subnet=20.20.0.0/16 chuku

# db
docker run --rm --net chuku --ip 20.20.1.1 --name chukudb -d -v $(pwd)/testdata:/testdata -v $(pwd)/testdata/postgres:/var/lib/postgresql/data -e POSTGRES_PASSWORD=password -e POSTGRES_DB=chukudb -d -p 5432:5432 postgres:12.2-alpine

# redis
docker run --rm --net chuku --ip 20.20.1.2 --name chuku-redis -d -p 6379:6379 redis

# app docker image build
docker image build -t chuku:latest .

# app run
docker run --rm --net chuku --ip 20.20.1.2 --name chuku-api -p 8000:8000 chuku:latest
