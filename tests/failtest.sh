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
CENTRAL_HOSTPORT="localhost:$CENTRAL_PORT"
PASS_RETURN_VAL=7

# Game Server Arguments
NUM_FAULTY_GAME_SERVERS=1
FAULTY_PERCENT=25
LAG_DURATION=15
LAG_PREPARE=true
LAG_ACCEPT=true
LAG_DECIDE=false
MAX_LAG_SLOTNUMBER=15

# Stress Test Arguments
NUM_GAME_CLIENTS=1
NUM_GAME_SERVERS=1
GS_HOSTPORTS_STRING=""
USE_CENTRAL=true
NUM_MOVES=10
NUM_SENDING_GAME_CLIENTS=1
SEND_MOVE_INTERVAL=1000
TIMEOUT=15

# Killer Arguments
NUM_TO_KILL=1
KILL_INTERVAL=10

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
  kill -9 ${CENTRAL_SERVER_PID} &> /dev/null
  wait ${CENTRAL_SERVER_PID} 2> /dev/null
}

function startGameServers {
  GS_HOSTPORTS_STRING=""
  GAME_SERVER_PORT=$(((RANDOM % 10000) + 10000))

  COUNTER=0
  # Start up the faulty servers
  for (( i=0; i < $NUM_FAULTY_GAME_SERVERS; i++ )); do
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=${CENTRAL_HOSTPORT} -faulty=true -faultyPercent=${FAULTY_PERCENT} -lagDuration=${LAG_DURATION} -lagPrepare=${LAG_PREPARE} -lagAccept=${LAG_ACCEPT} -lagDecide=${LAG_DECIDE} -maxLagSlotNumber=${MAX_LAG_SLOTNUMBER} &
    GAME_SERVER_PID[$COUNTER]=$!
    COUNTER=$((COUNTER + 1))
    GAME_SERVER_PORT=$((GAME_SERVER_PORT + 1))
  done
  echo "Started $NUM_FAULTY_GAME_SERVERS faulty game servers"

  REMAINING_SERVERS=$((NUM_GAME_SERVERS - NUM_FAULTY_GAME_SERVERS))

  for (( i=0; i < $REMAINING_SERVERS; i++)); do
    ${GAME_SERVER} -port=${GAME_SERVER_PORT} -central=${CENTRAL_HOSTPORT} &
    GAME_SERVER_PID[$COUNTER]=$!
    COUNTER=$((COUNTER + 1))
    GAME_SERVER_PORT=$((GAME_SERVER_PORT + 1))
  done
  echo "Started $REMAINING_SERVERS good game servers"

  TOTAL_GAME_SERVERS_STARTED=$COUNTER
}

function stopGameServers {
  # Kill the game servers
  for (( i=0; i < $TOTAL_GAME_SERVERS_STARTED; i++))
  do
      kill -9 ${GAME_SERVER_PID[$i]} &> /dev/null
      wait ${GAME_SERVER_PID[$i]} 2> /dev/null
  done
}

function doTest {
  echo "Stress test binary parameters numClients=$NUM_GAME_CLIENTS, numGameServers=$NUM_GAME_SERVERS, numMoves=$NUM_MOVES, numSendingClients=$NUM_SENDING_GAME_CLIENTS, usingCentral=$USE_CENTRAL, sendMoveInterval=$SEND_MOVE_INTERVAL, maxLagSlotNumber=$MAX_LAG_SLOTNUMBER"
  echo "Game server binary parameters, numFaulty=$NUM_FAULTY_GAME_SERVERS, faultyPercent=$FAULTY_PERCENT, lagDuration=$LAG_DURATION sec, lagPrepare=$LAG_PREPARE, lagAccept=$LAG_ACCEPT, lagDecide=$LAG_DECIDE"
  startCentralServer
  startGameServers
  sleep 3 # wait for the game servers to all join

  ${TEST} -numClients=${NUM_GAME_CLIENTS} -csHostPort=${CENTRAL_HOSTPORT} -numMoves=${NUM_MOVES} -numSendingClients=${NUM_SENDING_GAME_CLIENTS} -sendMoveInterval=${SEND_MOVE_INTERVAL} &
  TEST_PID=$!
  TIMEDOUT=false
  sleep ${TIMEOUT} && kill -9 ${TEST_PID} &> /dev/null && TIMEDOUT=true &

  # Start slaughtering game servers
  for (( i=0; i < $NUM_TO_KILL; i++ )); do
    sleep ${KILL_INTERVAL}
    NUM=$((i + 1))
    echo "Killing game server $NUM/$NUM_TO_KILL"
    kill -9 ${GAME_SERVER_PID[$i]} &> /dev/null
  done

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

function testThreeClientThreeGameOneFailure {
  echo "testOneClientThreeGameOneFailure"
  # Game server args
  NUM_FAULTY_GAME_SERVERS=0
  FAULTY_PERCENT=10
  LAG_DURATION=2
  LAG_PREPARE=true
  LAG_ACCEPT=true
  LAG_DECIDE=false
  MAX_LAG_SLOTNUMBER=17

  # Stress binary
  NUM_GAME_CLIENTS=3
  NUM_GAME_SERVERS=3
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=1
  SEND_MOVE_INTERVAL=1000
  USE_CENTRAL=true
  TIMEOUT=60

  # Slaughter
  NUM_TO_KILL=1
  KILL_INTERVAL=5

  doTest
}

function testNineClientThreeGameOneFailure {
  echo "testNineClientThreeGameOneFailure"
  # Game server args
  NUM_FAULTY_GAME_SERVERS=0
  FAULTY_PERCENT=10
  LAG_DURATION=2
  LAG_PREPARE=true
  LAG_ACCEPT=true
  LAG_DECIDE=false
  MAX_LAG_SLOTNUMBER=17

  # Stress binary
  NUM_GAME_CLIENTS=9
  NUM_GAME_SERVERS=3
  NUM_MOVES=10
  NUM_SENDING_GAME_CLIENTS=1
  SEND_MOVE_INTERVAL=1000
  USE_CENTRAL=true
  TIMEOUT=60

  # Slaughter
  NUM_TO_KILL=1
  KILL_INTERVAL=5

  doTest
}

function testTwentyClientFiveGameTwoFailure {
  echo "testTwentyClientFiveGameTwoFailure"
  # Game server args
  NUM_FAULTY_GAME_SERVERS=0
  FAULTY_PERCENT=10
  LAG_DURATION=2
  LAG_PREPARE=true
  LAG_ACCEPT=true
  LAG_DECIDE=false
  MAX_LAG_SLOTNUMBER=17

  # Stress binary
  NUM_GAME_CLIENTS=20
  NUM_GAME_SERVERS=5
  NUM_MOVES=15
  NUM_SENDING_GAME_CLIENTS=1
  SEND_MOVE_INTERVAL=1000
  USE_CENTRAL=true
  TIMEOUT=60

  # Slaughter
  NUM_TO_KILL=2
  KILL_INTERVAL=5

  doTest
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
testThreeClientThreeGameOneFailure
testNineClientThreeGameOneFailure
testTwentyClientFiveGameTwoFailure

echo "Passed (${PASS_COUNT}/$((PASS_COUNT + FAIL_COUNT))) tests"

exit 0