#!/bin/bash


if [ $# -lt 3 ]
then
  echo 'must provide the database password, qrz user and qrz password'
  exit 0
fi

rm master
go build  -o master cmd/web/*.go
./master -sqlpw $1 -qrzuser $2 -qrzpw $3 -lines $4
