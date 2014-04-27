#!/bin/bash

#if [ -z $1 ]; then
#    echo "Usage: ./simpletest.sh"
#    exit 1
#fi

# Commands
CENTRAL_SERVER="go run ../../runners/cservrunner.go"
GAME_SERVER="go run ../../runners/grunner.go"
TEST="go run simpletest.go"
CENTRAL_PORT=25340
GAME_SERVER_PORT=15551
NUM_GAME_SERVERS=1

echo "SCRIPT STARTING CENTRAL SERVER ON PORT ${CENTRAL_PORT}"
${CENTRAL_SERVER} -port=${CENTRAL_PORT} -gameservers=${NUM_GAME_SERVERS} &
sleep 2

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    echo "SCRIPT STARTING GAME SERVER ON PORT ${GAME_SERVER_PORT}"
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=localhost:${CENTRAL_PORT} &
    GAME_SERVER_PID[$i]=$!
    sleep 1
done

sleep 5
echo "SCRIPT STARTING TEST"
${TEST}

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    wait ${GAME_SERVER_PID[$i]}
done

exit 0