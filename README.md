![Go](https://github.com/EgorBedov/TechnoPark-RDBMS/workflows/Go/badge.svg)
[![Build Status](https://travis-ci.org/EgorBedov/TechnoPark-RDBMS.svg?branch=master)](https://travis-ci.org/EgorBedov/TechnoPark-RDBMS)

# TechnoPark-RDBMS

This is simple API server for requests from [tech-db-forum](https://github.com/bozaro/tech-db-forum)

## Purposes
 - Get deep SQL knowledge
 - Practice Golang
 - Get experience working with PostgreSQL
 - Get experience serving high-loaded application

## Why it's cool?
 - 1500 RPS
 - Clean architecture
 - REST API
 - Docker
 
## Docker installation

```
docker build -t e.bedov https://github.com/EgorBedov/TechnoPark-RDBMS.git
docker run -p 5000:5000 --name e.bedov -t e.bedov
```

## Local installation

Clone and run application (provided you have PostgreSQL up and running)
```
git clone https://github.com/EgorBedov/TechnoPark-RDBMS
createdb docker
psql docker -f $PATH_TO/scripts/init.sql
cd TechnoPark-RDBMS
go run cmd/server/main.go
```
> You can check whether app is running or not by visiting http://localhost:5000/api/service/status

Get task application
```
go get -u -v github.com/bozaro/tech-db-forum
go build github.com/bozaro/tech-db-forum
```

Run functionality test
```
./tech-db-forum func -u http://localhost:5000/api
```
 
Run perfomance test
```
./tech-db-forum fill --timeout=900
./tech-db-forum perf --duration=600 --step=60
```
> Note that this could fail with timeout during filling and any error during perf test

## Additionally
 - All the info about testing application you can find [here](https://github.com/bozaro/tech-db-forum)
 - You can enable/disable logger by commenting [this line](https://github.com/EgorBedov/TechnoPark-RDBMS/blob/b8ec3e615029e07ea0e08ef8786f2003d7749494/internal/app/server/settings.go#L50)
