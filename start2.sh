#!/bin/bash

if [ -z $1 ]; then
    echo "Usage: ./start.sh [numGameServers]"
    exit 1
fi

if [ -z $2 ]; then
  HOSTNAME="localhost"
else
  HOSTNAME=$2
fi

# Params
CENTRAL_SERVER_PKG="distributed2048/runners/cservrunner"
GAME_SERVER_PKG="distributed2048/runners/grunner"
CLIENT_SERVER_PKG="distributed2048/runners/crunner"
CENTRAL_HOSTNAME=localhost
CENTRAL_PORT=25340
GAME_SERVER_PORT=15551
NUM_GAME_SERVERS=$1

# Build and install the Central Server binary
go install ${CENTRAL_SERVER_PKG}
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Build and install the Game Server binary
go install ${GAME_SERVER_PKG}
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Build and install the Game Server binary
go install ${CLIENT_SERVER_PKG}
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Commands
CENTRAL_SERVER=$GOPATH/bin/cservrunner
GAME_SERVER=$GOPATH/bin/grunner
CLIENT_SERVER=$GOPATH/bin/crunner

echo "SCRIPT STARTING CENTRAL SERVER ON PORT ${CENTRAL_PORT}"
${CENTRAL_SERVER} -port=${CENTRAL_PORT} -gameservers=${NUM_GAME_SERVERS} &
${CLIENT_SERVER} &
sleep 2

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    # Pick random ports between [10000, 20000).
    GAME_SERVER_PORT=$(((RANDOM % 10000) + 10000))
    echo "SCRIPT STARTING GAME SERVER ON PORT ${GAME_SERVER_PORT}"
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -hostname="localhost" -central=localhost:${CENTRAL_PORT} &
    GAME_SERVER_PID[$i]=$!
    sleep 1
done

echo "SCRIPT STARTED ALL CRAP"

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    wait ${GAME_SERVER_PID[$i]}
done

exit 0