var CENTRAL_HOSTPORT = "http://localhost:25340/"

function GameManager(size, InputManager, Actuator, StorageManager, ConnManager) {
  this.size           = size; // Size of the grid
  this.inputManager   = new InputManager;
  this.storageManager = new StorageManager;
  this.actuator       = new Actuator;
  this.connManager = new ConnManager;

  this.startTiles     = 2;
    // We want this to "not actually" touch the game state
  this.inputManager.on("move", this.move.bind(this));

  this.inputManager.on("restart", this.restart.bind(this));
  this.inputManager.on("keepPlaying", this.keepPlaying.bind(this));

  this.connManager.on("connectionMade", this.setup.bind(this));
  this.connManager.on("update", this.update.bind(this));


  this.connManager.getConnectionFromCServ()
}

// Restart the game
GameManager.prototype.restart = function () {
//  this.storageManager.clearGameState();
//  this.actuator.continueGame(); // Clear the game won/lost message
//  this.setup();
    alert("You can't restart lol")
};

// Keep playing after winning (allows going over 2048)
GameManager.prototype.keepPlaying = function () {
  this.keepPlaying = true;
  this.actuator.continueGame(); // Clear the game won/lost message
};

// Return true if the game is lost, or has won and the user hasn't kept playing
GameManager.prototype.isGameTerminated = function () {
  return this.over || (this.won && !this.keepPlaying);
};

// Set up the game
// I think we don't need a storage manager lol
GameManager.prototype.setup = function () {
    console.log('setting up the game');
    this.grid        = new Grid(this.size);
    this.score       = 0;
    this.over        = false;
    this.won         = false;
    this.keepPlaying = false;
    this.bestScore   = 0;

  // Update the actuator
  this.actuate();
};

// Set up the initial tiles to start the game with
GameManager.prototype.addStartTiles = function () {
  for (var i = 0; i < this.startTiles; i++) {
    this.addRandomTile();
  }
};

// Sends the updated grid to the actuator
GameManager.prototype.actuate = function () {
  this.actuator.actuate(this.grid, {
    score:      this.score,
    over:       this.over,
    won:        this.won,
    bestScore:  this.bestScore,
    terminated: false,
  });
    console.log("GameManager done actuating");
};

// Represent the current game as an object
GameManager.prototype.serialize = function () {
  return {
    grid:        this.grid.serialize(),
    score:       this.score,
    over:        this.over,
    won:         this.won,
    keepPlaying: this.keepPlaying
  };
};

// Save all tile positions and remove merger info
GameManager.prototype.prepareTiles = function () {
  this.grid.eachCell(function (x, y, tile) {
    if (tile) {
      tile.mergedFrom = null;
      tile.savePosition();
    }
  });
};

// Move a tile and its representation
GameManager.prototype.moveTile = function (tile, cell) {
  this.grid.cells[tile.x][tile.y] = null;
  this.grid.cells[cell.x][cell.y] = tile;
  tile.updatePosition(cell);
};

GameManager.prototype.move = function (dir) {
    console.log("direction of proposed move is" + dir);
    var message = {Direction: dir};
    console.log("sending move: " + dir);
    this.connManager.connection.send(JSON.stringify(message));
};

GameManager.prototype.update = function (data) {
    console.log("updating");
    this.score = data.Score;
    this.grid.cells = this.grid.fromState(data.Grid);
    if (this.score > this.bestScore) {
        this.bestScore = this.score;
    }
    if (data.Won) {
        console.log("won game");
        this.won = true;
        this.grid = data.Grid;
        // We actuate this and then automatically restart
        this.actuate();
        this.grid = new Grid(this.size);
        this.won = false;
        this.over = false;
        this.score = 0;
    } else if (data.Over) {
        console.log("lost game");
        this.over = true;
        this.actuate();
    }
    console.log(this.grid);
    this.actuate();
    return;
};

// Get the vector representing the chosen direction
GameManager.prototype.getVector = function (direction) {
  // Vectors representing tile movement
  var map = {
    0: { x: 0,  y: -1 }, // Up
    1: { x: 1,  y: 0 },  // Right
    2: { x: 0,  y: 1 },  // Down
    3: { x: -1, y: 0 }   // Left
  };

  return map[direction];
};

// Build a list of positions to traverse in the right order
GameManager.prototype.buildTraversals = function (vector) {
  var traversals = { x: [], y: [] };

  for (var pos = 0; pos < this.size; pos++) {
    traversals.x.push(pos);
    traversals.y.push(pos);
  }

  // Always traverse from the farthest cell in the chosen direction
  if (vector.x === 1) traversals.x = traversals.x.reverse();
  if (vector.y === 1) traversals.y = traversals.y.reverse();

  return traversals;
};

GameManager.prototype.findFarthestPosition = function (cell, vector) {
  var previous;

  // Progress towards the vector direction until an obstacle is found
  do {
    previous = cell;
    cell     = { x: previous.x + vector.x, y: previous.y + vector.y };
  } while (this.grid.withinBounds(cell) &&
           this.grid.cellAvailable(cell));

  return {
    farthest: previous,
    next: cell // Used to check if a merge is required
  };
};

GameManager.prototype.movesAvailable = function () {
  return this.grid.cellsAvailable() || this.tileMatchesAvailable();
};

// Check for available matches between tiles (more expensive check)
GameManager.prototype.tileMatchesAvailable = function () {
  var self = this;

  var tile;

  for (var x = 0; x < this.size; x++) {
    for (var y = 0; y < this.size; y++) {
      tile = this.grid.cellContent({ x: x, y: y });

      if (tile) {
        for (var direction = 0; direction < 4; direction++) {
          var vector = self.getVector(direction);
          var cell   = { x: x + vector.x, y: y + vector.y };

          var other  = self.grid.cellContent(cell);

          if (other && other.value === tile.value) {
            return true; // These two tiles can be merged
          }
        }
      }
    }
  }

  return false;
};

GameManager.prototype.positionsEqual = function (first, second) {
  return first.x === second.x && first.y === second.y;
};
