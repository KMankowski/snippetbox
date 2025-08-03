#!/bin/bash
source .env
if [[ -z $mysql_password ]]; then
	exit 1
fi

dsn="web:${mysql_password}@/snippetbox?parseTime=true"

go run ./cmd/web -dsn=${dsn}
