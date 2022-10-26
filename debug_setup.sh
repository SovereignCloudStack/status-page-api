#!/bin/bash

set -ex

podman network exists statuspage || podman network create statuspage

# Start up Postgres compatible database
podman container exists statuspage_database && podman container start statuspage_database || \
  podman run --name statuspage_database --network statuspage -p 5432:5432 -e POSTGRES_PASSWORD=debug -d postgres
