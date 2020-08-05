#!/usr/bin/env bash

# TODO: If container already exist just run `docker start miniflux_db`
docker run --name miniflux_db -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=miniflux2 postgres
