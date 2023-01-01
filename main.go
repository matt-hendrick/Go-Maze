package main

import (
	"container/heap"
	"container/list"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
)

var (
	//go:embed "main.html"
	mapHTML     string
	mapTemplate = template.Must(template.New("maze").Parse(mapHTML))
)

type DataToInject struct {
	BFS   MazeHTMLAndPath
	DFS   MazeHTMLAndPath
	AStar MazeHTMLAndPath
}

type MazeHTMLAndPath struct {
	HTML template.HTML
	Path []string
}

type Node struct {
	x         int
	y         int
	obstacle  bool
	fScore    int // bottom three attributes are only used for A*
	gScore    int
	heapIndex int
}

func (node *Node) Equals(otherNode *Node) bool {
	return node.x == otherNode.x && node.y == otherNode.y
}

func (node *Node) InitializeNode(maze *Maze, y int, x int) {
	node.x = x
	node.y = y
	node.fScore = math.MaxInt32
	node.gScore = math.MaxInt32
}

func (node *Node) ToString() string {
	return fmt.Sprintf("%d,%d", node.y, node.x)
}

type Maze struct {
	matrix               [][]Node
	queueToVisit         *list.List     // used for BFS
	stackToVisit         Stack          // used for DFS
	priorityQueueToVisit *PriorityQueue // used for A*
	visitedSet           Set
	start                Node
	end                  Node
	pathToEnd            []string
}

func NewMaze(size int) *Maze {
	var maze Maze
	maze.matrix = make([][]Node, size)
	for row := range maze.matrix {
		maze.matrix[row] = make([]Node, size)
	}

	maze.visitedSet = *NewSet()
	maze.end.x = rand.Intn(size)
	maze.end.y = rand.Intn(size)
	maze.start.x = rand.Intn(size)
	maze.start.y = rand.Intn(size)
	maze.start.fScore = 0
	maze.start.gScore = calculateDistance(&maze.start, &maze.end)

	for iY, row := range maze.matrix {
		for iX, _ := range row {
			maze.matrix[iY][iX].InitializeNode(&maze, iY, iX)
			maze.AddObstacles(iY, iX)
		}
	}

	return &maze
}

func (maze *Maze) AddObstacles(y int, x int) {
	// don't put an obstacle on the starting or ending cells
	if (y == maze.start.y && x == maze.start.x) || (y == maze.end.y && x == maze.end.x) {
		return
	}
	if rand.Intn(10) < 3 {
		maze.matrix[y][x].obstacle = true
	}
}

type Set struct {
	hashmap map[string]struct{}
}

func NewSet() *Set {
	set := &Set{}
	set.hashmap = make(map[string]struct{})
	return set
}

func (set *Set) Add(value string) {
	set.hashmap[value] = struct{}{}
}

func (set *Set) Remove(value string) {
	delete(set.hashmap, value)
}

func (set *Set) Contains(value string) bool {
	_, exists := set.hashmap[value]
	return exists
}

func (set *Set) Clear() {
	*set = Set{}
	set.hashmap = make(map[string]struct{})
}

type Stack []Node

func (stack *Stack) IsEmpty() bool {
	return len(*stack) == 0
}

func (stack *Stack) Push(node Node) {
	*stack = append(*stack, node)
}

func (stack *Stack) Pop() (Node, bool) {
	if stack.IsEmpty() {
		return Node{}, false
	} else {
		index := len(*stack) - 1   // Get the index of the top most element.
		element := (*stack)[index] // Index into the slice and obtain the element.
		*stack = (*stack)[:index]  // Remove it from the stack by slicing it off.
		return element, true
	}
}

func (stack *Stack) Clear() {
	*stack = (*stack)[:0]
}

func (stack *Stack) Print(endNode *Node) {
	for idx, cell := range *stack {
		fmt.Printf("%d #: %s, distance = %d, fScore = %d, gScore = %d\n", idx, cell.ToString(), calculateDistance(&cell, endNode), cell.fScore, cell.gScore)
	}
}

// PriorityQueue implementation partly based upon standard library documentation example (https://pkg.go.dev/container/heap)
type PriorityQueue []*Node

func NewPriorityQueue(maze *Maze) *PriorityQueue {
	var pq PriorityQueue = make(PriorityQueue, 0)
	pq.Push(&maze.start)
	heap.Init(&pq)
	return &pq
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].gScore > pq[j].gScore
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].heapIndex = i
	pq[j].heapIndex = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	node := x.(*Node)
	node.heapIndex = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil      // avoid memory leak
	node.heapIndex = -1 // for safety
	*pq = old[0 : n-1]
	return node
}

func (pq *PriorityQueue) Clear() {
	*pq = (*pq)[:0]
}

func (pq *PriorityQueue) Update(node *Node) {
	heap.Push(pq, node)
	heap.Fix(pq, node.heapIndex)
}

func main() {
	http.HandleFunc("/", httpHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	if err := mapTemplate.Execute(w, getData()); err != nil {
		log.Fatal(err)
	}
}

// gets data to inject into main.html
func getData() DataToInject {
	var maze *Maze = NewMaze(40)
	var data DataToInject = getDataForEachAlgoType(maze)
	return data
}

func getDataForEachAlgoType(maze *Maze) DataToInject {
	algos := [3]string{"DFS", "BFS", "AStar"}
	var data DataToInject

	for _, algoType := range algos {
		clearMazeState(maze)
		solveMaze(maze, algoType)
		var mazeHtmlAndPath MazeHTMLAndPath
		mazeHtmlAndPath.HTML = generateMazeHTML(algoType, maze)
		mazeHtmlAndPath.Path = maze.pathToEnd

		if algoType == "DFS" {
			data.DFS = mazeHtmlAndPath
		} else if algoType == "BFS" {
			data.BFS = mazeHtmlAndPath
		} else {
			data.AStar = mazeHtmlAndPath
		}
	}

	return data
}

func clearMazeState(maze *Maze) {
	maze.visitedSet.Clear()
	maze.stackToVisit.Clear()
	if maze.queueToVisit != nil {
		maze.queueToVisit.Init()
	}
	if maze.priorityQueueToVisit != nil {
		maze.priorityQueueToVisit.Clear()
	}
	maze.pathToEnd = nil
}

func solveMaze(maze *Maze, algoType string) {
	if algoType == "DFS" {
		DFS(maze)
	} else if algoType == "BFS" {
		BFS(maze)
	} else {
		AStar(maze)
	}
}

func generateMazeHTML(algoType string, maze *Maze) template.HTML {
	var html template.HTML

	for iY, y := range maze.matrix {
		for iX, _ := range y {
			// add opening div for row
			if iX == 0 {
				html += "<div class=row>"
			}

			// add divs for cells
			if iY == maze.start.y && iX == maze.start.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:gold' id='%d,%d-%s'></div>", iY, iX, algoType))
			} else if iY == maze.end.y && iX == maze.end.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:red' id='%d,%d-%s'></div>", iY, iX, algoType))
			} else if maze.matrix[iY][iX].obstacle {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:black' id='%d,%d-%s'></div>", iY, iX, algoType))
			} else {
				html += template.HTML(fmt.Sprintf("<div class='cell' id='%d,%d-%s'></div>", iY, iX, algoType))
			}

			// add closing div for row
			if iX+1 >= len(y) {
				html += "</div>"
			}
		}
	}

	return html
}

func canVisit(y int, x int, nodeString string, maze *Maze) bool {
	// is within bounds
	if x < 0 || y < 0 || y >= len(maze.matrix) || x >= len(maze.matrix[y]) {
		return false
	}
	if maze.visitedSet.Contains(nodeString) {
		return false
	}
	if maze.matrix[y][x].obstacle {
		return false
	}
	return true
}

func visitCell(y int, x int, currNode *Node, currNodeString string, maze *Maze, algoType string) {
	if canVisit(y, x, currNodeString, maze) {
		if algoType == "DFS" {
			maze.stackToVisit.Push(maze.matrix[y][x])
		} else if algoType == "BFS" {
			maze.queueToVisit.PushBack(maze.matrix[y][x])
		} else {
			aStarProcessNeighbor(&maze.matrix[y][x], currNode, maze)
		}
	}
}

func visitNeighboringCells(currNode *Node, currNodeString string, maze *Maze, algoType string) {
	// visit top neighbor
	visitCell(currNode.y-1, currNode.x, currNode, currNodeString, maze, algoType)
	// visit left neighbor
	visitCell(currNode.y, currNode.x-1, currNode, currNodeString, maze, algoType)
	// visit bottom neighbor
	visitCell(currNode.y+1, currNode.x, currNode, currNodeString, maze, algoType)
	// visit right neighbor
	visitCell(currNode.y, currNode.x+1, currNode, currNodeString, maze, algoType)
}

func DFS(maze *Maze) *Maze {
	maze.stackToVisit.Push(maze.start)
	var algoType string = "DFS"
	for len(maze.stackToVisit) > 0 {
		currNode, _ := maze.stackToVisit.Pop()
		currNodeString := currNode.ToString()
		if currNode.Equals(&maze.end) {
			// has reached the end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
			break
		} else if !maze.visitedSet.Contains(currNodeString) {
			visitNeighboringCells(&currNode, currNodeString, maze, algoType)
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
		}
	}
	return maze
}

func BFS(maze *Maze) *Maze {
	maze.queueToVisit = list.New()
	maze.queueToVisit.PushBack(maze.start)
	var algoType string = "BFS"
	for maze.queueToVisit.Len() > 0 {
		front := maze.queueToVisit.Front()
		currNode := front.Value.(Node)
		maze.queueToVisit.Remove(front)
		currNodeString := currNode.ToString()
		if currNode.Equals(&maze.end) {
			// has reached the end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
			break
		} else if !maze.visitedSet.Contains(currNodeString) {
			visitNeighboringCells(&currNode, currNodeString, maze, algoType)
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
		}
	}

	return maze
}

func intAbs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

// using manhattan distance as only 4 directions of movement are allowed through the grid (no diagonals)
func calculateDistance(node1 *Node, node2 *Node) int {
	dx := intAbs(node1.x - node2.x)
	dy := intAbs(node1.y - node2.y)
	return dx + dy
}

// func calculateEuclidianDistance(node1 *Node, node2 *Node) float64 {
// 	return math.Sqrt(math.Pow(float64(node1.x-node2.x), 2) + math.Pow(float64(node1.y-node2.y), 2))
// }

func aStarProcessNeighbor(neighbor *Node, currNode *Node, maze *Maze) {
	fScoreThroughCurrentNode := currNode.fScore + 1
	if fScoreThroughCurrentNode < neighbor.fScore {
		neighbor.fScore = fScoreThroughCurrentNode
	}
	neighbor.gScore = neighbor.fScore + calculateDistance(neighbor, &maze.end)
	maze.priorityQueueToVisit.Update(neighbor)
}

func AStar(maze *Maze) *Maze {
	maze.priorityQueueToVisit = NewPriorityQueue(maze)
	var algoType string = "AStar"
	for maze.priorityQueueToVisit.Len() > 0 {
		currNode := maze.priorityQueueToVisit.Pop().(*Node)
		currNodeString := currNode.ToString()
		if currNode.Equals(&maze.end) {
			// has reached end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
			break
		} else if !maze.visitedSet.Contains(currNodeString) {
			visitNeighboringCells(currNode, currNodeString, maze, algoType)
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString+"-"+algoType)
		}
	}
	return maze
}
