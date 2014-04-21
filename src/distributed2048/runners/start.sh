#!/bin/bash

if [ -z $1 ]; then
    echo "Usage: ./start.sh [numGameServers]"
    exit 1
fi

# Commands
CENTRAL_SERVER="go run cservrunner.go"
GAME_SERVER="go run grunner.go"
CLIENT_SERVER="go run crunner.go"
CENTRAL_PORT=15340
NUM_GAME_SERVERS=$1

echo "SCRIPT STARTING CENTRAL SERVER ON PORT ${CENTRAL_PORT}"
${CENTRAL_SERVER} -port=${CENTRAL_PORT} -gameservers=${NUM_GAME_SERVERS} &
${CLIENT_SERVER} &
sleep 2

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    # Pick random ports between [10000, 20000).
    GAME_SERVER_PORT=$(((RANDOM % 10000) + 10000))
    echo "SCRIPT STARTING GAME SERVER ON PORT ${GAME_SERVER_PORT}"
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=localhost:${CENTRAL_PORT} &
    GAME_SERVER_PID[$i]=$!
    sleep 1
done

echo "SCRIPT STARTED ALL CRAP"

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    wait ${GAME_SERVER_PID[$i]}
done

exit 0