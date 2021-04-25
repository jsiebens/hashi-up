#! /bin/bash

apt update
apt install unzip -y
docker run -d -p 5432:5432 -e POSTGRES_USER=boundary -e POSTGRES_PASSWORD=boundary123 -e POSTGRES_DB=boundary postgres:12