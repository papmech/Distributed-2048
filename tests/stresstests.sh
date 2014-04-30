#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

# Params
CENTRAL_SERVER_PKG="distributed2048/runners/cservrunner"
GAME_SERVER_PKG="distributed2048/runners/grunner"
TEST_PKG="distributed2048/tests/stresstest"
CENTRAL_PORT=25340
PASS_RETURN_VAL=7

# Stress Test Arguments
NUM_GAME_CLIENTS=1
NUM_GAME_SERVERS=1
GS_HOSTPORTS_STRING=""
USE_CENTRAL=true
NUM_MOVES=10
NUM_SENDING_GAME_CLIENTS=1
SEND_MOVE_INTERVAL=1000
TIMEOUT=15

# Commands
CENTRAL_SERVER=$GOPATH/bin/cservrunner
GAME_SERVER=$GOPATH/bin/grunner
TEST=$GOPATH/bin/stresstest

function startCentralServer {
  # echo "[TESTSCRIPT] Starting central server on ${CENTRAL_PORT}"
  ${CENTRAL_SERVER} -port=${CENTRAL_PORT} -gameservers=${NUM_GAME_SERVERS} &
  CENTRAL_SERVER_PID=$!
  sleep 1
}

function stopCentralServer {
  # Kill the central server
  kill -9 ${CENTRAL_SERVER_PID}
  wait ${CENTRAL_SERVER_PID} 2> /dev/null
}

function startGameServers {
  GS_HOSTPORTS_STRING=""
  FIRST_GAME_SERVER_PORT=$(((RANDOM % 10000) + 10000))

  for (( i=0; i < $NUM_GAME_SERVERS; i++)); do
    GAME_SERVER_PORT=$(($FIRST_GAME_SERVER_PORT + $i))
    # echo "[TESTSCRIPT] Starting game server on ${GAME_SERVER_PORT}"
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=localhost:${CENTRAL_PORT} &
    GAME_SERVER_PID[$i]=$!
    if [ $i != 0 ]; then
      GS_HOSTPORTS_STRING="$GS_HOSTPORTS_STRING,"
    fi
    THIS_HOSTPORT="localhost:$GAME_SERVER_PORT"
    GS_HOSTPORTS_STRING="$GS_HOSTPORTS_STRING$THIS_HOSTPORT"
    sleep 1
  done
  # echo $GS_HOSTPORTS_STRING
}

function stopGameServers {
  # Kill the game servers
  for (( i=0; i < $NUM_GAME_SERVERS; i++))
  do
      kill -9 ${GAME_SERVER_PID[$i]}
      wait ${GAME_SERVER_PID[$i]} 2> /dev/null
  done
}

function doStressTest {
  echo "Starting stress test, numClients=$NUM_GAME_CLIENTS, numGameServers=$NUM_GAME_SERVERS, numMoves=$NUM_MOVES, numSendingClients=$NUM_SENDING_GAME_CLIENTS, usingCentral=$USE_CENTRAL, sendMoveInterval=$SEND_MOVE_INTERVAL"
  startCentralServer
  startGameServers
  sleep 3 # wait for the game servers to all join

  if $USE_CENTRAL; then
    GS_HOSTPORTS_STRING=""
    CENTRAL_HOSTPORT="localhost:$CENTRAL_PORT"
  else
    CENTRAL_HOSTPORT=""
  fi

  ${TEST} -numClients=${NUM_GAME_CLIENTS} -gsHostPorts=${GS_HOSTPORTS_STRING} -csHostPort=${CENTRAL_HOSTPORT} -numMoves=${NUM_MOVES} -numSendingClients=${NUM_SENDING_GAME_CLIENTS} -sendMoveInterval=${SEND_MOVE_INTERVAL} &
  TEST_PID=$!
  sleep ${TIMEOUT} && kill -9 ${TEST_PID} &> /dev/null && TIMEDOUT=true &

  wait ${TEST_PID} 2> /dev/null
  if [ "$?" -ne $PASS_RETURN_VAL ]; then
    FAIL_COUNT=$((FAIL_COUNT + 1))
    if $TIMEDOUT; then
      echo "TIMED OUT"
    fi
  else
    PASS_COUNT=$((PASS_COUNT + 1))
  fi
  stopGameServers
  stopCentralServer
}

function testOneClientOneGameYesCentral {
  echo "testOneClientOneGameYesCentral"
  NUM_GAME_CLIENTS=1
  NUM_GAME_SERVERS=1
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=1
  USE_CENTRAL=true
  TIMEOUT=30
  doStressTest
}

function testFiveClientOneGameYesCentral {
  echo "testFiveClientOneGameYesCentral"
  NUM_GAME_CLIENTS=5
  NUM_GAME_SERVERS=1
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=5
  USE_CENTRAL=true
  TIMEOUT=30
  doStressTest
}

function testOneClientFiveGameYesCentral {
  echo "testOneClientFiveGameYesCentral"
  NUM_GAME_CLIENTS=1
  NUM_GAME_SERVERS=5
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=1
  USE_CENTRAL=true
  TIMEOUT=30
  doStressTest
}

function testThreeClientFiveGameYesCentral {
  echo "testThreeClientFiveGameYesCentral"
  NUM_GAME_CLIENTS=3
  NUM_GAME_SERVERS=5
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=3
  USE_CENTRAL=true
  TIMEOUT=30
  doStressTest
}

function testTwentyFiveClientFiveGameYesCentral {
  echo "testTwentyFiveClientFiveGameYesCentral"
  NUM_GAME_CLIENTS=25
  NUM_GAME_SERVERS=5
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=25
  USE_CENTRAL=true
  TIMEOUT=60
  doStressTest
}

function testTwentyFiveClientFiveGameYesCentralHundredMoves {
  echo "testTwentyFiveClientFiveGameYesCentralHundredMoves"
  NUM_GAME_CLIENTS=25
  NUM_GAME_SERVERS=5
  NUM_MOVES=100
  NUM_SENDING_GAME_CLIENTS=25
  USE_CENTRAL=true
  SEND_MOVE_INTERVAL=25
  TIMEOUT=600
  doStressTest
}

function testTwentyFiveClientFiveGameYesCentralThousandMoves {
  echo "testTwentyFiveClientFiveGameYesCentralThousandMoves"
  NUM_GAME_CLIENTS=25
  NUM_GAME_SERVERS=5
  NUM_MOVES=1000
  NUM_SENDING_GAME_CLIENTS=25
  USE_CENTRAL=true
  SEND_MOVE_INTERVAL=25
  TIMEOUT=1200
  doStressTest
}

function testHundredClientTwentyFiveGameYesCentralThousandMoves {
  echo "testHundredClientTwentyFiveGameYesCentralThousandMoves"
  NUM_GAME_CLIENTS=100
  NUM_GAME_SERVERS=25
  NUM_MOVES=1000
  NUM_SENDING_GAME_CLIENTS=100
  USE_CENTRAL=true
  SEND_MOVE_INTERVAL=25
  TIMEOUT=1200
  doStressTest
}

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

# Build and install the Test binary
go install ${TEST_PKG}
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Run tests
PASS_COUNT=0
FAIL_COUNT=0
testOneClientOneGameYesCentral
testFiveClientOneGameYesCentral
testOneClientFiveGameYesCentral
testThreeClientFiveGameYesCentral
testTwentyFiveClientFiveGameYesCentral
testTwentyFiveClientFiveGameYesCentralHundredMoves
testTwentyFiveClientFiveGameYesCentralThousandMoves
testHundredClientTwentyFiveGameYesCentralThousandMoves
echo "Passed (${PASS_COUNT}/$((PASS_COUNT + FAIL_COUNT))) tests"

exit 0