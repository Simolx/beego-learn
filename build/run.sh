#!/bin/bash

docker run -it --rm --name zkserver --hostname KafkaService -p 9090:9090 zkserver
