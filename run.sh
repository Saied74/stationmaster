#!/bin/bash


if [ $# -lt 1 ]
then
  echo 'must provide the database password'
  exit 0
fi

rm master
go build -o master cmd/web/*.go
./master -pw $1
