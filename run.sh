#!/usr/bin/env bash

PORT=4001 DATABASE_URL='user=postgres password=postgres dbname=miniflux2 sslmode=disable host=localhost' BASE_URL='http://localhost:4001/' ./miniflux.app $@
