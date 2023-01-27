#!/bin/bash

curl -X GET 'http://localhost:9090/server/health'; echo
curl -X POST 'http://localhost:9090/server/start'; echo
