#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

# Params
CENTRAL_SERVER_PKG="distributed2048/runners/cservrunner"
GAME_SERVER_PKG="distributed2048/runners/grunner"
TEST_PKG="distributed2048/tests/simpletests"
CENTRAL_PORT=25340
GAME_SERVER_PORT=15551
NUM_GAME_SERVERS=1

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

# Build and install the Simple Test binary
go install ${TEST_PKG}
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

CENTRAL_SERVER=$GOPATH/bin/cservrunner
GAME_SERVER=$GOPATH/bin/grunner
TEST=$GOPATH/bin/simpletests

echo "SCRIPT STARTING CENTRAL SERVER ON PORT ${CENTRAL_PORT}"
${CENTRAL_SERVER} -port=${CENTRAL_PORT} -gameservers=${NUM_GAME_SERVERS} &
CENTRAL_SERVER_PID=$!
sleep 2

for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    echo "SCRIPT STARTING GAME SERVER ON PORT ${GAME_SERVER_PORT}"
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=localhost:${CENTRAL_PORT} &
    GAME_SERVER_PID[$i]=$!
    sleep 1
done

sleep 2
echo "SCRIPT STARTING TEST"
${TEST}

# Kill the game servers
for (( i=0; i < $NUM_GAME_SERVERS; i++))
do
    kill -9 ${GAME_SERVER_PID[$i]}
    wait ${GAME_SERVER_PID[$i]} 2> /dev/null
done

# Kill the central server
kill -9 ${CENTRAL_SERVER_PID}
wait ${CENTRAL_SERVER_PID} 2> /dev/null

exit 0