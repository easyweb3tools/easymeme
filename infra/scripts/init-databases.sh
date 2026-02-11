#!/bin/sh
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE base_db;
    CREATE DATABASE easymeme_db;
EOSQL
