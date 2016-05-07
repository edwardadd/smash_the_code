package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	//"time"
)

const (
	GRID_WIDTH  int = 6
	GRID_HEIGHT int = 12
	SAMPLES     int = 2000
	DEPTH       int = 8
	EMPTY_SPACE     = '.' - 48
)

var ErrNoMoreSpace = errors.New("No more space!")
var ErrAlreadyExplored = errors.New("Leaf already explored")

var g_data [22][2]int = [22][2]int{
	{2, 0}, {2, 1}, {2, 2}, {2, 3},
	{4, 0}, {4, 1}, {4, 2}, {4, 3},
	{3, 0}, {3, 1}, {3, 2}, {3, 3},
	{1, 0}, {1, 1}, {1, 2}, {1, 3},
	{0, 0}, {0, 1}, {0, 3},
	{5, 1}, {5, 2}, {5, 3},
}

var colourString [5]string = [5]string{
	"Blue",
	"Green",
	"Pink",
	"Red",
	"Yellow",
}

var g_invalidate bool = false

var g_chainDepression = 4

type Grid [GRID_WIDTH * GRID_HEIGHT]int

type Node struct {
	choice       int
	grid         Grid
	score        int
	chainCount   int
	turn         int
	turnExplored int
	position     int
	rotation     int

	nodes   [22]*Node
	parent  *Node
	err     error
	message string
	invalid bool
}

type Stats struct {
	nextHistogram         [5]int
	playerGridHistogram   [5]int
	maximumPossibleChains [5]int
}

type Game struct {
	nextColours [8][2]int
	playerGrid  Grid
	cpuGrid     Grid
	turn        int
	stats       Stats
	node        *Node
}

// ============================================================================

func main() {
	var game Game = Game{}

	game.initialise()
	game.gameLoop()
}

// Game =======================================================================

func (game *Game) initialise() {
	game.turn = 0
}

func (game *Game) gameLoop() {

	game.node = &Node{
		turn:   game.turn,
		parent: nil,
	}

	rand.Seed(643)

	for {
		//parseStart := time.Now()
		parseNextBlocks(&game.nextColours)
		parseGrid(&game.playerGrid)
		parseGrid(&game.cpuGrid)
		//parseElapsed := time.Since(parseStart)
		//fmt.Fprintln(os.Stderr, "Parse Time: ", parseElapsed)

		if game.turn == 0 {
			game.node.grid = game.playerGrid
		}

		// fmt.Fprintln(os.Stderr, "Debug messages...")

		if game.node.chainCount > 0 {
			g_chainDepression--

			if g_chainDepression == 0 {
				g_chainDepression = 4
			}
		}

		//game.node.grid.print("Previous Grid")

		// game.analyseNextColours()
		g_invalidate = skullAttack(&game.node.grid, &game.playerGrid)
		if g_invalidate {
			g_chainDepression = 2
			game.node.grid = game.playerGrid

			for i := 0; i < 22; i++ {
				if game.node.nodes[i] != nil {
					game.node.nodes[i].invalid = true
				}
			}
		}

		//game.playerGrid.print("Current Grid")

		//start := time.Now()
		for i := 0; i < SAMPLES; i++ {
			explore(betterChoice(), game.node, game.turn, DEPTH, &game.nextColours, 0)
		}
		//elapsed := time.Since(start)
		//fmt.Fprintln(os.Stderr, "Timer: ", elapsed)

		//printTree(game.node, 1)

		//find choice with greatest score

		bestNode, nodeCount := game.chooseBestNode()

		if bestNode != nil {

			if bestNode.score == 0 {
				// randomly pick one
				num := rand.Intn(nodeCount - 1)
				for i := 0; i < 22; i++ {
					if game.node.nodes[num] == nil {
						num++
					} else {
						bestNode = game.node.nodes[num]
						bestNode.message = "Luck of the draw"
						break
					}
				}
			}

			game.node = bestNode
			bestNode.parent = nil

			fmt.Fprintln(os.Stderr, "bestNode score", bestNode.score)
			// bestNode.grid.print()

			//for i, node := range bestNode.nodes {
			//    fmt.Fprintf(os.Stderr, "%d> child score %d\n", i, node.score)
			//}
			output(bestNode.position, bestNode.rotation, bestNode.message)
		} else {
			// No more good moves... so ganme over!
			fmt.Println("0 0 It's game over, man! IT'S GAME OVER!")
		}
		game.turn++
	}
}

func (game *Game) chooseBestNode() (*Node, int) {
	var bestNode *Node = nil
	var nodeCount int

	for n := 0; n < 22; n++ {
		var node *Node = game.node.nodes[n]

		if node == nil || node.err != nil || node.invalid {
			continue
		}

		nodeCount++

		if bestNode == nil {
			bestNode = node
		}

		if node.score > bestNode.score {
			bestNode = node
		}
	}

	return bestNode, nodeCount
}

func (game *Game) analyseNextColours() {
	// fmt.Fprintln(os.Stderr, "analyseNextColours")
	// fmt.Fprintln(os.Stderr, "NC", game.nextColours)
	// fmt.Fprintln(os.Stderr, "PGrid", game.playerGrid)

	// histogram of colours
	var histogram [5]int

	for _, colours := range game.nextColours {
		for _, colour := range colours {
			histogram[colour-1]++
		}
	}

	game.stats.nextHistogram = histogram

	histogram = [5]int{0, 0, 0, 0, 0}
	for _, colour := range game.playerGrid {
		if colour > 0 {
			histogram[colour-1]++
		}
	}

	game.stats.playerGridHistogram = histogram

	for i, colourCount := range game.stats.nextHistogram {
		histogram[i] += colourCount
	}

	for i, colourCount := range histogram {
		game.stats.maximumPossibleChains[i] = colourCount / 4 // rough...

		// fmt.Fprintln(os.Stderr, "Possible Chains ", game.stats.maximumPossibleChains)
	}

}

func skullAttack(oldGrid *Grid, newGrid *Grid) bool {
	//oldGrid.print("Last predicted:")
	//newGrid.print("Current state:")

	var count int = 0

	for i := 0; i < GRID_WIDTH*GRID_HEIGHT; i++ {
		if oldGrid[i] != newGrid[i] {
			count++
		}
	}

	return count > 5
}

func parseNextBlocks(nextColours *[8][2]int) {
	for i := 0; i < 8; i++ {
		// colorA: color of the first block
		// colorB: color of the attached block
		var colorA, colorB int
		fmt.Scan(&colorA, &colorB)
		nextColours[i][0] = colorA
		nextColours[i][1] = colorB
	}
}

func parseGrid(grid *Grid) {
	for i := 0; i < 12; i++ {
		var row string
		fmt.Scan(&row)

		for j := 0; j < 6; j++ {
			value := int(row[j])
			grid[i*GRID_WIDTH+j] = value - 48
		}
	}
}

func choiceToAction(choice int) (int, int) {
	// 6 positions and 4 possible rotations
	// except at the edges where there are 3 rotations
	// 4 * 4 + 2 * 3 = 22

	return g_data[choice][0], g_data[choice][1]
}

func (grid *Grid) applyGravity() {
	// fmt.Fprintf(os.Stderr, "applyGravity\n")

	// grid.print()

	// not efficient or helping the cacheline
	for x := 0; x < GRID_WIDTH; x++ {
		var lastFilled int = GRID_HEIGHT - 1
		for y := GRID_HEIGHT - 1; y >= 0; y-- {
			index := x + y*GRID_WIDTH
			if grid[index] > EMPTY_SPACE {
				// check above
				if lastFilled-y > 0 {
					// drop
					// fmt.Fprintf(os.Stderr, "replace %d, %d with %d, %d\n", x, y, x, lastFilled)
					grid[x+lastFilled*GRID_WIDTH] = grid[index]
					grid[index] = EMPTY_SPACE
				}

				lastFilled--
			}
		}
	}

	// fmt.Fprintf(os.Stderr, "applied\n")
	// grid.print()
	// fmt.Fprintf(os.Stderr, "applyGravity Done\n")
}

func countConnectedBlocks(grid Grid, x int, y int, visited *Grid) (foundColour int, count int) {
	// fmt.Fprintf(os.Stderr, "findConnectedBlocks at %d, %d\n", x, y)
	initialIndex := x + y*GRID_WIDTH
	if grid[initialIndex] == EMPTY_SPACE || grid[initialIndex] == 0 {
		return EMPTY_SPACE, 0
	}

	//grid.print()

	var stack [GRID_WIDTH * GRID_HEIGHT]int
	var si, ci int = 0, 0
	var colour int = grid[initialIndex]

	stack[si] = initialIndex
	si++

	for {
		index := stack[ci]
		visited[index] = 1

		if index/GRID_HEIGHT > 0 && index-GRID_WIDTH >= 0 {
			block := grid[index-GRID_WIDTH]
			if block == colour {
				stack[si] = index - GRID_WIDTH
				si++
			}
		}

		if index/GRID_HEIGHT < GRID_HEIGHT-1 && index+GRID_WIDTH < GRID_WIDTH*GRID_HEIGHT {
			block := grid[index+GRID_WIDTH]
			if block == colour {
				stack[si] = index + GRID_WIDTH
				si++
			}
		}

		if index%GRID_WIDTH > 0 && index-1 >= 0 {
			block := grid[index-1]
			if block == colour {
				stack[si] = index - 1
				si++
			}
		}

		if index%GRID_WIDTH < GRID_WIDTH-1 && index+1 < GRID_WIDTH*GRID_HEIGHT {
			block := grid[index+1]
			if block == colour {
				stack[si] = index + 1
				si++
			}
		}

		ci++

		if ci >= si {
			break
		}
	}

	return colour, ci
}

func findConnectedBlocks(originalGrid *Grid, x int, y int, visited *Grid) (int, int, int) {
	// fmt.Fprintf(os.Stderr, "findConnectedBlocks at %d, %d\n", x, y)
	initialIndex := x + y*GRID_WIDTH
	if originalGrid[initialIndex] == EMPTY_SPACE || originalGrid[initialIndex] == 0 {
		return EMPTY_SPACE, 0, 0
	}

	//grid.print()

	var grid Grid = *originalGrid
	var mappedToVisit map[int]bool = map[int]bool{}
	var stack [GRID_WIDTH * GRID_HEIGHT]int
	var si, ci int = 0, 0
	var colour int = grid[initialIndex]
	var skullCount int = 0

	stack[si] = initialIndex
	si++

	for {
		index := stack[ci]
		grid[index] = EMPTY_SPACE
		visited[index] = 1
		mappedToVisit[index] = true

		if index-GRID_WIDTH >= 0 && !mappedToVisit[index-GRID_WIDTH] {
			block := grid[index-GRID_WIDTH]
			if block == colour {
				stack[si] = index - GRID_WIDTH
				si++
				mappedToVisit[index-GRID_WIDTH] = true
			}
		}

		if index+GRID_WIDTH < GRID_WIDTH*GRID_HEIGHT && !mappedToVisit[index+GRID_WIDTH] {
			block := grid[index+GRID_WIDTH]
			if block == colour {
				stack[si] = index + GRID_WIDTH
				si++
				mappedToVisit[index+GRID_WIDTH] = true
			}
		}

		if (index-1)%GRID_WIDTH < GRID_WIDTH-1 && index-1 >= 0 && !mappedToVisit[index-1] {
			block := grid[index-1]
			if block == colour {
				stack[si] = index - 1
				si++
				mappedToVisit[index-1] = true
			}
		}

		if (index+1)%GRID_WIDTH > 0 && index+1 < GRID_WIDTH*GRID_HEIGHT && !mappedToVisit[index+1] {
			block := grid[index+1]
			if block == colour {
				stack[si] = index + 1
				si++
				mappedToVisit[index+1] = true
			}
		}

		// fmt.Fprintf(os.Stderr, "si %d, ci %d\n", si, ci)

		ci++

		if ci >= si {
			break
		}
	}

	if ci > 3 {

		// Remove skulls
		ci = 0
		for {
			index := stack[ci]

			if index-GRID_WIDTH >= 0 {
				block := grid[index-GRID_WIDTH]
				if block == 0 {
					grid[index-GRID_WIDTH] = EMPTY_SPACE
					visited[index-GRID_WIDTH] = 1
					skullCount++
				}
			}

			if index+GRID_WIDTH < GRID_WIDTH*GRID_HEIGHT {
				block := grid[index+GRID_WIDTH]
				if block == 0 {
					grid[index+GRID_WIDTH] = EMPTY_SPACE
					visited[index+GRID_WIDTH] = 1
					skullCount++
				}
			}

			if (index-1)%GRID_WIDTH < GRID_WIDTH-1 && index-1 >= 0 {
				block := grid[index-1]
				if block == 0 {
					grid[index-1] = EMPTY_SPACE
					visited[index-1] = 1
					skullCount++
				}
			}

			if (index+1)%GRID_WIDTH > 0 && index+1 < GRID_WIDTH*GRID_HEIGHT {
				block := grid[index+1]
				if block == 0 {
					grid[index+1] = EMPTY_SPACE
					visited[index+1] = 1
					skullCount++
				}
			}

			// fmt.Fprintf(os.Stderr, "si %d, ci %d\n", si, ci)

			ci++

			if ci >= si {
				break
			}
		}

		*originalGrid = grid
	}

	return colour, ci, skullCount
}

func groupBonus(blocks int) int {
	var score int = blocks - 4
	if score < 0 {
		return 0
	}

	if score > 8 {
		return 8
	}

	return score
}

func colourBonus(colours int) int {
	var score int = 1
	if colours == 0 {
		return 0
	}

	for i := 1; i < colours && i < 5; i++ {
		score *= 2
	}
	return score
}

func chainPowerForStep(step int) int {
	// CP is the chain power, starting at 0 for the first step.
	// It is worth 8 for the second step and for each following step it is worth twice as much as the previous step.

	var chainPower int = 0
	for i := 0; i < step; i++ {
		if i == 1 {
			chainPower = 8
		} else if i > 1 {
			chainPower *= 2
		}
	}

	return chainPower
}

func (grid *Grid) print(title string) {
	fmt.Fprintf(os.Stderr, "%s\n{\n", title)
	for y := 0; y < GRID_HEIGHT; y++ {
		fmt.Fprintf(os.Stderr, "    ")
		for x := 0; x < GRID_WIDTH; x++ {
			fmt.Fprintf(os.Stderr, "%02d, ", grid[x+y*GRID_WIDTH])
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, "}\n")
}

func (node *Node) print() {
	fmt.Fprintf(os.Stderr, "T:%02d C:%02d S:%02d p/r %d,%d\n", node.turn, node.choice, node.score, node.position, node.rotation)
}

func printTree(node *Node, depth int) {
	if depth == 0 {
		return
	}

	for n, node := range node.nodes {
		if node == nil {
			continue
		}

		for tab := 0; tab < node.turn; tab++ {
			fmt.Fprintf(os.Stderr, "-")
		}

		fmt.Fprintf(os.Stderr, " %d] ", n)
		node.print()

		printTree(node, depth-1)
	}
}

func output(position int, rotation int, message string) {
	if message == "" {
		fmt.Printf("%d %d\n", position, rotation)
	} else {
		fmt.Printf("%d %d %s\n", position, rotation, message)
	}
}

func betterChoice() int {
	// choice := float32(rand.Intn(22)) / 22

	// rng := (1 - choice * choice) * 21

	// fmt.Fprintf(os.Stderr, "%f %f\n", rng, choice)
	// return int(rng)
	return rand.Intn(22)
}

///////////////////////////////////////////////////////////////////////////////
/////////////////////////////////// FREQUENTLY TWEAKED ////////////////////////
///////////////////////////////////////////////////////////////////////////////

func explore(choice int, node *Node, currentTurn int, maxDepth int, nextBlocks *[8][2]int, exploreType int) error {
	// fmt.Fprintf(os.Stderr, "Explore Turn %d - %d\n", currentTurn, node.turn)
	var newNode *Node = nil

	if node.nodes[choice] != nil {
		// fmt.Fprintf(os.Stderr, "Already explored Turn %d - %d, score %d\n", currentTurn, node.turn, node.nodes[choice].score)
		newNode = node.nodes[choice]

		if newNode.invalid || newNode.turnExplored != currentTurn {
			newNode.grid = node.grid
			newNode.score = 0
			newNode.message = ""
			if newNode.invalid {
				newNode.message = "Damn those skulls!"
			}
			newNode.turnExplored = currentTurn

			err := simulate(newNode, currentTurn, nextBlocks)
			newNode.err = err

			if err != nil {
				return err
			}
		}
	} else {

		position, rotation := choiceToAction(choice)

		newNode = &Node{
			choice:       choice,
			grid:         node.grid,
			turn:         node.turn + 1,
			turnExplored: currentTurn,
			position:     position,
			rotation:     rotation,
			score:        0,
			parent:       node,
			err:          nil,
		}

		node.nodes[choice] = newNode

		err := simulate(newNode, currentTurn, nextBlocks)
		newNode.err = err

		if err != nil {
			return err
		}
	}

	switch exploreType {

	case 0:
		if newNode.turn-currentTurn < maxDepth {
			//if newNode.turn - currentTurn > 7 {
			//	nextBlocks = &[8][2]int{
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//		{rand.Intn(4) + 1, rand.Intn(4) + 1},
			//	}
			//}

			for {
				err := explore(betterChoice(), newNode, currentTurn, maxDepth, nextBlocks, exploreType)
				if err == nil {
					break
				}
			}
		}
	case 1:
		if newNode.turn-currentTurn < maxDepth {
			for i := 0; i < 22; i++ {
				explore(i, newNode, currentTurn, maxDepth, nextBlocks, exploreType)
			}
		}
	}

	return nil
}

func simulate(node *Node, currentTurn int, nextBlocks *[8][2]int) error {
	var leftX, rightX int

	if node.invalid {
		for i := 0; i < 22; i++ {
			if node.nodes[i] != nil {
				node.nodes[i].invalid = true
			}
		}
	}

	// Fill the grid
	next := ((node.turn - currentTurn) - 1) % 8
	// fmt.Fprintf(os.Stderr, "SImulate Turn %d - %d - %d %d\n", currentTurn, node.turn, node.position, node.rotation)

	colour := nextBlocks[next]

	if node.rotation == 1 || node.rotation == 3 {
		leftX = node.position
		rightX = node.position
	} else {
		if node.rotation == 0 {
			leftX = node.position
			rightX = node.position + 1
		} else {
			leftX = node.position - 1
			rightX = node.position
		}
	}

	var highestPositions [GRID_WIDTH]int = highPosition(node.grid)

	if node.rotation == 0 || node.rotation == 2 {
		if highestPositions[leftX] == EMPTY_SPACE || highestPositions[rightX] == EMPTY_SPACE {
			return ErrNoMoreSpace
		}
	} else {
		if highestPositions[leftX] < 1 {
			return ErrNoMoreSpace
		}
	}

	leftY, rightY := positionBlockInGridWithY(&node.grid, leftX, rightX, node.rotation, colour[0], colour[1], highestPositions[leftX], highestPositions[rightX])

	// node.grid.print()

	// Check for clearable blocks
	// If found then update grid and check again

	var finalScore int = 0
	var tempGrid Grid = node.grid
	var chainCount int = 0
	var actualScore int = 0

	var blocksAboveThree int = 0
	var averageNeighbouringBlockCount int = 0
	var groupCount int = 0
	var averageChainBlock int
	var skullCountCleared int = 0

	var aVisited Grid

	//check for clearing at recently dropped position
	_, count0, sk0 := findConnectedBlocks(&tempGrid, leftX, leftY, &aVisited)
	_, count1, sk1 := findConnectedBlocks(&tempGrid, rightX, rightY, &aVisited)

	// actualScore += 20 * count0 + 20 * count1
	if count0 > 0 {
		groupCount++
	}

	if count1 > 0 {
		groupCount++
	}

	const blocksMakeClear int = 4
	if count0 >= blocksMakeClear {
		averageChainBlock += count0
		skullCountCleared += sk0
	}

	if count1 >= blocksMakeClear {
		averageChainBlock += count1
		skullCountCleared += sk1
	}

	searchFurther := count0 >= blocksMakeClear || count1 >= blocksMakeClear

	averageNeighbouringBlockCount += count0 + count1

	// fmt.Fprintf(os.Stderr, "initial chain count %d, blocks %d %d\n", chainCount, count0, count1)

	if searchFurther {
		for {
			//find connected blocks and continue
			var visited Grid
			var chainThisStep bool

			if chainCount == 0 {
				visited = aVisited
				chainThisStep = true
				chainCount = 0
			}

			for i := 0; i < GRID_WIDTH*GRID_HEIGHT; i++ {
				var x int = i % GRID_WIDTH
				var y int = i / GRID_WIDTH
				if tempGrid[i] > 0 && visited[i] == 0 {
					c, blockCount, sk := findConnectedBlocks(&tempGrid, x, y, &visited)
					if blockCount >= 4 {
						chainThisStep = true
						node.message = fmt.Sprintf(" Can Clear %s", colourString[c-1])
						averageChainBlock += blockCount
						skullCountCleared += sk
					}
					if blockCount >= 3 {
						blocksAboveThree++
					}
					if blockCount > 0 {
						groupCount++
					}
					averageNeighbouringBlockCount += blockCount
				}
			}

			if !chainThisStep {
				break
			}

			chainCount++
			tempGrid.applyGravity()
		}

		tempGrid.applyGravity()
	}

	// actualScore = (10 * count0) * (chainPower + colourBonus + groupBonus)

	// Higher the score the lower the average line
	var heightBonus int = 0
	for i := 0; i < GRID_WIDTH; i++ {
		heightBonus += highestPositions[i]
	}
	heightBonus = int(float32(heightBonus) / 6.0)

	// average Groupd size
	if groupCount > 0 {
		averageNeighbouringBlockCount = averageNeighbouringBlockCount / groupCount

		averageChainBlock = averageChainBlock / groupCount
	} else {
		averageNeighbouringBlockCount = 0

		averageChainBlock = 0
	}

	// if chainCount > 0 {
	//     fmt.Fprintf(os.Stderr, "Now %d Turn %d, chainCount %d \n", currentTurn, node.turn, chainCount)
	//     node.grid.print("Initial")
	//     tempGrid.print("altered")
	// }
	chainExpected := g_chainDepression
	chainScore := 1 - (chainCount-chainExpected)*(chainCount-chainExpected)
	actualScore = chainScore*100*(8-next) + averageChainBlock*chainCount*10 + skullCountCleared*100*chainCount // + blocksAboveThree
	finalScore = actualScore + heightBonus*30 + (averageNeighbouringBlockCount - 3)

	// Update node
	node.score = finalScore
	node.chainCount = chainCount
	node.grid = tempGrid
	node.invalid = false
	if node.chainCount > 0 {
		node.message = fmt.Sprintf("Go! Go! Gadget Chain x%d", node.chainCount)
	}

	// Back propagate the score
	node.backPropagateScore()

	return nil
}

func highPosition(grid Grid) [GRID_WIDTH]int {
	var positions [GRID_WIDTH]int
	for x := 0; x < GRID_WIDTH; x++ {
		positions[x] = EMPTY_SPACE

		for y := GRID_HEIGHT - 1; y >= 0; y-- {
			index := x + y*GRID_WIDTH

			if grid[index] == EMPTY_SPACE {
				positions[x] = y
				break
			}
		}
	}
	return positions
}

func positionBlockInGridWithY(grid *Grid, leftX int, rightX int, rotation int, colourA int, colourB int, leftY int, rightY int) (int, int) {

	if rotation == 0 {
		indexA := leftX + leftY*GRID_WIDTH
		indexB := rightX + rightY*GRID_WIDTH
		grid[indexA] = colourA
		grid[indexB] = colourB

		return leftY, rightY
	} else if rotation == 2 {
		indexB := leftX + leftY*GRID_WIDTH
		indexA := rightX + rightY*GRID_WIDTH
		grid[indexA] = colourA
		grid[indexB] = colourB

		return rightY, leftY
	} else if rotation == 1 {
		indexB := leftX + (leftY-1)*GRID_WIDTH
		indexA := rightX + leftY*GRID_WIDTH
		grid[indexA] = colourA
		grid[indexB] = colourB

		return leftY, leftY - 1
	} else {
		indexA := leftX + (leftY-1)*GRID_WIDTH
		indexB := rightX + leftY*GRID_WIDTH
		grid[indexA] = colourA
		grid[indexB] = colourB

		return leftY - 1, leftY
	}
}

func (grid *Grid) findChainCount() ([8][2]int, int) {
	var chainCount int
	var clearInfo [8][2]int

	for {
		//find connected blocks and continue
		var thisCount int = 0
		var visited Grid

		for i := 0; i < GRID_WIDTH*GRID_HEIGHT; i++ {
			var x int = i % GRID_WIDTH
			var y int = i / GRID_WIDTH
			if grid[i] > 0 && visited[i] == 0 {
				c, blockCount, _ := findConnectedBlocks(grid, x, y, &visited)
				if blockCount >= 4 {
					clearInfo[chainCount+thisCount] = [2]int{c, blockCount}
					thisCount++
				}
			}
		}

		if thisCount == 0 {
			break
		}

		chainCount += thisCount
		grid.applyGravity()
	}

	return clearInfo, chainCount
}

func (node *Node) backPropagateScore() {
	var parent *Node = node.parent
	for {
		if parent == nil {
			break
		}

		if parent.score < node.score {
			parent.score = node.score
		}
		parent = parent.parent
	}
}
