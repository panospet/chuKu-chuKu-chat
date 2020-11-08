# create network + subnet
docker network create -d bridge --subnet=20.20.0.0/16 chuku

# db
docker run --rm --net chuku --ip 20.20.1.1 --name chukudb -d -v $(pwd)/testdata:/testdata -v $(pwd)/testdata/postgres:/var/lib/postgresql/data -e POSTGRES_PASSWORD=password -e POSTGRES_DB=chukudb -d -p 5432:5432 postgres:12.2-alpine

# migrate
make migrate

# redis
docker run --rm --net chuku --ip 20.20.1.2 --name chuku-redis -d -p 6379:6379 redis

# app docker image build
docker image build -t chuku:latest .

# app run
docker run --rm --net chuku --ip 20.20.1.3 --name chuku-api -p 9009:8000 -d chuku:latest

# frontend build
docker build -t takxous:dev .

# frontend app
docker run -it --rm --name chuku-frontend -v $(pwd):/app -v /app/node_modules -p 3001:3000 -e CHOKIDAR_USEPOLLING=true -d takxous:dev