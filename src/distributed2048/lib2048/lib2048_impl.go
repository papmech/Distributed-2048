// ported from https://github.com/olafurw/ML2048

package lib2048

import (
	"distributed2048/libsimplerand"
	"fmt"
)

type Grid [BoardLen][BoardLen]int

type game struct {
	grid  Grid
	score int
	r     *libsimplerand.SimpleRand
}

func NewGame2048() Game2048 {
	g := &game{
		score: 0,
		r:     libsimplerand.NewSimpleRand(15440),
	}
	g.reset()
	g.newRound(InitialTileCount)
	return g
}

func (g *game) MakeMove(dir Direction) {
	if !g.canMove() {
		return
	}
	a := g.move(dir)
	b := g.merge(dir)
	c := g.move(dir)

	if a || b || c {
		g.newRound(EachTurnNewTileCount)
	}

}

func (g *game) GetScore() int {
	return g.score
}

func (g *game) GetBoard() Grid {
	return g.grid
}

func (g *game) GetRand() *libsimplerand.SimpleRand {
	return g.r
}

func (g *game) IsGameOver() bool {
	return g.IsGameWon() || !g.canMove()
}

func (g *game) IsGameWon() bool {
	return g.getLargest() == 2048
}

func (g *game) String() string {
	result := ""
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			result += fmt.Sprintf("%d\t", g.grid[row][col])
		}
		result += "\n"
	}
	result += fmt.Sprintf("Score: %d\n", g.score)
	return result
}

func (g *game) Equals(other Game2048) bool {
	// Check that the grids are equal
	otherGrid := other.GetBoard()
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			if g.grid[row][col] != otherGrid[row][col] {
				return false
			}
		}
	}

	// Check that the scores are equal
	if g.GetScore() != other.GetScore() {
		return false
	}
	return true
}

func (g *game) SetGrid(grid Grid) {
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			g.grid[row][col] = grid[row][col]
		}
	}
}

func (g *game) SetScore(score int) {
	g.score = score
}

func (g *game) CloneFrom(other Game2048) {
	otherGrid := other.GetBoard()
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			g.grid[row][col] = otherGrid[row][col]
		}
	}
	g.score = other.GetScore()
	g.r = other.GetRand()
}

func (g *game) newRound(numNewTiles int) {
	for i := 0; i < numNewTiles; i++ {
		y, x := g.randomEmptyPos()
		if x == -1 && y == -1 {
			return // Board full
		}
		g.grid[y][x] = FirstTileValue
		if g.shouldInitialValueBeDouble() {
			g.grid[y][x] = FirstTileValue * 2
		}
	}
}

// Sets all values to 0
func (g *game) reset() {
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			g.grid[row][col] = 0
		}
	}
}

func (g *game) isOutside(row, col int) bool {
	return row < 0 || col < 0 || row >= BoardLen || col >= BoardLen
}

func (g *game) canMove() bool {
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			// 1 empty space?
			if g.grid[row][col] == 0 {
				return true
			}

			value := g.grid[row][col]
			// NORTH
			if g.get(row-1, col) == value {
				return true
			}
			// SOUTH
			if g.get(row+1, col) == value {
				return true
			}
			// EAST
			if g.get(row, col+1) == value {
				return true
			}
			// WEST
			if g.get(row, col-1) == value {
				return true
			}
		}
	}
	return false
}

func (g *game) get(row, col int) int {
	if g.isOutside(row, col) {
		return -1
	}
	return g.grid[row][col]
}

func (g *game) hasEmpty() bool {
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			if g.grid[row][col] == 0 {
				return true
			}
		}
	}
	return false
}

func (g *game) set(row, col, value int) {
	if g.isOutside(row, col) {
		return
	}
	g.grid[row][col] = value
}

func (g *game) getLargest() int {
	largest := 0
	for row := 0; row < BoardLen; row++ {
		for col := 0; col < BoardLen; col++ {
			if g.grid[row][col] > largest {
				largest = g.grid[row][col]
			}
		}
	}
	return largest
}

func (g *game) move(dir Direction) bool {
	hasMovement := false
	switch dir {
	case Down:
		for x := 0; x < BoardLen; x++ {
			for y := BoardLen - 1; y >= 0; y-- {
				// Don't move empty spaces
				if g.grid[y][x] == 0 {
					continue
				}

				newY, nextY := y, y+1
				for !g.isOutside(nextY, x) && g.grid[nextY][x] == 0 {
					newY = nextY
					nextY++
				}

				if newY != y {
					hasMovement = true
				}

				value := g.grid[y][x]
				g.grid[y][x] = 0
				g.grid[newY][x] = value
			}
		}
	case Up:
		for x := 0; x < BoardLen; x++ {
			for y := 0; y < BoardLen; y++ {
				// Don't move empty spaces
				if g.grid[y][x] == 0 {
					continue
				}

				newY, nextY := y, y-1
				for !g.isOutside(nextY, x) && g.grid[nextY][x] == 0 {
					newY = nextY
					nextY--
				}

				if newY != y {
					hasMovement = true
				}

				value := g.grid[y][x]
				g.grid[y][x] = 0
				g.grid[newY][x] = value
			}
		}
	case Left:
		for x := 0; x < BoardLen; x++ {
			for y := 0; y < BoardLen; y++ {
				// Don't move empty spaces
				if g.grid[y][x] == 0 {
					continue
				}

				newX, nextX := x, x-1
				for !g.isOutside(y, nextX) && g.grid[y][nextX] == 0 {
					newX = nextX
					nextX--
				}
				if newX != x {
					hasMovement = true
				}

				value := g.grid[y][x]
				g.grid[y][x] = 0
				g.grid[y][newX] = value
			}
		}
	case Right:
		for x := BoardLen - 1; x >= 0; x-- {
			for y := 0; y < BoardLen; y++ {
				// Don't move empty spaces
				if g.grid[y][x] == 0 {
					continue
				}

				newX, nextX := x, x+1
				for !g.isOutside(y, nextX) && g.grid[y][nextX] == 0 {
					newX = nextX
					nextX++
				}
				if newX != x {
					hasMovement = true
				}

				value := g.grid[y][x]
				g.grid[y][x] = 0
				g.grid[y][newX] = value
			}
		}
	}
	return hasMovement
}

// Goes through and merges in that direction
func (g *game) merge(dir Direction) bool {
	merging := false

	switch dir {
	case Down:
		for x := 0; x < BoardLen; x++ {
			for y := BoardLen - 1; y >= 0; y-- {
				// Empty slots dont merge
				if g.grid[y][x] == 0 {
					continue
				}

				nextY, value := y+1, g.grid[y][x]
				merge_value := g.get(nextY, x)

				if value == merge_value {
					new_value := value + value

					g.grid[y][x] = 0
					g.grid[nextY][x] = new_value

					merging = true

					g.score += new_value
				}
			}
		}

		return merging
	case Up:
		for x := 0; x < BoardLen; x++ {
			for y := 0; y < BoardLen; y++ {
				// Empty slots dont merge
				if g.grid[y][x] == 0 {
					continue
				}

				nextY := y - 1
				value := g.grid[y][x]
				merge_value := g.get(nextY, x)

				if value == merge_value {
					new_value := value + value

					g.grid[y][x] = 0
					g.grid[nextY][x] = new_value

					merging = true

					g.score += new_value
				}
			}
		}

		return merging
	case Left:
		for x := 0; x < BoardLen; x++ {
			for y := 0; y < BoardLen; y++ {
				// Empty slots dont merge
				if g.grid[y][x] == 0 {
					continue
				}

				nextX := x - 1
				value := g.grid[y][x]
				merge_value := g.get(y, nextX)

				if value == merge_value {
					new_value := value + value

					g.grid[y][x] = 0
					g.grid[y][nextX] = new_value

					merging = true

					g.score += new_value
				}
			}
		}

		return merging
	case Right:
		for x := BoardLen - 1; x >= 0; x-- {
			for y := 0; y < BoardLen; y++ {
				// Empty slots dont merge
				if g.grid[y][x] == 0 {
					continue
				}

				nextX := x + 1
				value := g.grid[y][x]
				merge_value := g.get(y, nextX)

				if value == merge_value {
					new_value := value + value

					g.grid[y][x] = 0
					g.grid[y][nextX] = new_value

					merging = true

					g.score += new_value
				}
			}
		}

		return merging
	}

	return merging
}

// Returns row, col
func (g *game) randomEmptyPos() (int, int) {
	if !g.hasEmpty() {
		return -1, -1
	}
	x, y := g.randPos(), g.randPos()
	for g.grid[y][x] != 0 {
		x, y = g.randPos(), g.randPos()
	}
	return y, x
}

func (g *game) randPos() int {
	return g.r.Int() % BoardLen
}

func (g *game) shouldInitialValueBeDouble() bool {
	return g.r.Int()%100 < InitialTileDoublePercent
}
