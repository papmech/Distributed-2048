<html>
<head>
</head>

<body>
<h1>Distributed 2048</h1>
<p>15-440 Distributed Project 3</p>
<hr>
<h2>Overview</h2>
<ul>
    <li>
        <b>Frontend</b>: 2048 web client mainly done in Javascript.
    </li>
    <li>
        <b>Backend</b>: A cluster of Go servers that stores the game state.
    </li>
    <li>
        <b>Client-server communication</b>: We are using websockets(which are supported in both Go and Javascript) to facilitate communication between the client and servers.
    </li>
    <li>
        <b>Initial connection</b>: We have a "central server" that gives connecting clients the hostport of the specific server that they should connect to.
    </li>
    <li>
        <b>Distributed replication</b>: We are using Paxos as a consensus and backup protocol among the servers.
    </li>
</ul>
<h2>Requirements</h2>
<p>This project uses the excellent go websockets implementation found at <a href="http://godoc.org/code.google.com/p/go.net/websocket">http://godoc.org/code.google.com/p/go.net/websocket</a>. Please run <strong>go get "code.google.com/p/go.net/websocket"</strong> before running anything.</p>

<h2>Mechanics</h2>
<p>
We will first launch the central server and provide a specific number of game servers that we expect to be connected. Next we will launch the game servers that will register with the central server.
</p>

<p>
Users will first go to a website that we are serving. On the client end, the website will do a GET request to the central server, which will give a response detailing the hostport of an available gameserver.
</p>

<p>
The website then facilitates a websocket connection attempt from the client to the game server. At this stage, users can use arrow keys and WSAD to do 'moves', which are sent to the gameserver via the websocket connection.
</p>

<p>
The servers agree on a majority move and update their respective game states, sending the game state back to the client. On the user-end, the website receives the game state and updates the 2048 board that the user sees.
</p>

<h2>Voting protocol</h2>
<p>
    Each server collects the moves sent by all of its clients within a set interval of time. That would be considered as a 'commit'. If this list of moves is non-empty, the server proposes the commit to the cluster via paxos. If it goes through, all the servers are now updated with a list of 'client moves'. This is how we maintain a common gate state among the servers.
</p>
<p>
    We use a simple algorithm to estimate the number of clients. If the commit has <i>x</i> length, we know that there are roughly <i>x</i> active clients connected to that server. Thus we multiply by <i>y</i>, the total number of servers, to get <i>z</i>, target number of move votes. When <i>z</i> votes are collected, we simply count the majority vote and that will be the next move for the common game state. The update shall be propagated to the clients.
</p>

<h2>Failure</h2>
<p>
    We are able to tolerate at most <i>p</i> failures given <i>2p+1</i> total servers courtesy of the paxos protocol. Upon connection failure, a client will repeatedly query the central server until it receives a new game server to connect to. This game server is not guaranteed to be alive. If it is not, the client will again ask the central server for a new server.
</p>

<h2>Testing</h2>
<p>
    We wrote a Go client that runs on the command line specifically for testing purposes. In terms of functionality it is exactly the same as the javascript client. However, we modified it such that it could take a sequence of moves and send them at specified intervals to a gameserver, and is able to return the game state at any point in time. We have a function that simulates running an expected sequence of moves to arrive at some game state. We would then compare the board states to test if the outcome is as expected. We wrote tests with differing number of clients and servers and various scenarios to verify the robustness, correctness, and capacity of the system.
</p>
<p>
    Test files are run exactly as they are without necessary arguments.
    <ul>
        <li>
            <b>simpletest.sh</b>: A single test that serves more of an end-to-end sanity check that everything works as it should.
        </li>
        <li>
            <b>stresstests.sh</b>: Tests with increasing number of clients, moves, and decreasing move intervals.
        </li>
        <li>
            <b>interruptedtests.sh</b>: Introducing random delays (a.k.a. lag) of increasing length into the Paxos instances, so that RPC calls time out and message ordering is totally messed up.
        </li>
        <li>
            <b>failtest.sh</b>: Kill servers and check that game clients reconnect to another server, game state is preserved and replicated correctly via successful Paxos rounds.
        </li>
        <li>
            <b>killall.sh</b>: Not a test file, but useful for killing test processes if user Ctrl+C out of the test.
        </li>
    </ul>
</p>
<h2>Credits</h2>
<p>
The Javascript 2048 client was taken from the 2048 project repository by Gabriele Cirulli and contributors and modified for our purposes.
</p>

</body>
</html>

