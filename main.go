package main

import (
	"container/list"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

var (
	//go:embed "main.html"
	mapHTML     string
	mapTemplate = template.Must(template.New("maze").Parse(mapHTML))
)

type Node struct {
	x         int
	y         int
	obstacle  bool
	parent    *Node // bottom four attributes are only used for A* atm
	neighbors []*Node
	fScore    float64
	gScore    float64
}

func NewNode(y, x int) Node {
	node := Node{}
	node.y = y
	node.x = x
	return node
}

func (node *Node) ToString() string {
	return fmt.Sprintf("%d,%d", node.y, node.x)
}

type Maze struct {
	matrix [][]Node
	// maybe add a min heap or priority queue for AStar
	queueToVisit *list.List
	stackToVisit Stack
	visitedSet   Set
	start        Node
	end          Node
	pathToEnd    []string
}

func initializeNode(node *Node, maze *Maze, y int, x int) {
	node.x = x
	node.y = y
	node.fScore = math.Inf(1)
	node.gScore = math.Inf(1)
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

	// TODO: Check if there's a better way of initalizing the 2D array without the duplicate nested loops
	for iY, row := range maze.matrix {
		for iX, _ := range row {
			initializeNode(&maze.matrix[iY][iX], &maze, iY, iX)
			addObstacles(&maze, iY, iX)
			// fmt.Printf("Y: %d, X: %d \n", iX, iY)
		}
	}

	return &maze
}

func addObstacles(maze *Maze, y int, x int) {
	// don't put an obstacle on the starting or ending cells
	if (y == maze.start.y && x == maze.start.x) || (y == maze.end.y && x == maze.end.x) {
		return
	}
	if rand.Intn(10) < 3 {
		maze.matrix[y][x].obstacle = true
	}
}

type JSData struct {
	HTML template.HTML
	Path []string
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

func (stack *Stack) Print(endNode *Node) {
	for idx, cell := range *stack {
		fmt.Printf("%d #: %s, distance = %f\n", idx, cell.ToString(), calculateDistance(&cell, endNode))
	}
}

func main() {

	http.HandleFunc("/DFS", dfsHandler)

	http.HandleFunc("/BFS", bfsHandler)

	http.HandleFunc("/AStar", aStarHandler)

	http.HandleFunc("/", aStarHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func dfsHandler(w http.ResponseWriter, r *http.Request) {
	if err := mapTemplate.Execute(w, getData("DFS")); err != nil {
		log.Fatal(err)
	}
}

func bfsHandler(w http.ResponseWriter, r *http.Request) {
	if err := mapTemplate.Execute(w, getData("BFS")); err != nil {
		log.Fatal(err)
	}
}

func aStarHandler(w http.ResponseWriter, r *http.Request) {
	if err := mapTemplate.Execute(w, getData("AStar")); err != nil {
		log.Fatal(err)
	}
}

// gets data to inject into main.html
func getData(algoToUse string) JSData {
	var maze *Maze = createAndSolveMaze(algoToUse)
	var data JSData
	var html template.HTML = ""

	for iY, y := range maze.matrix {
		for iX, _ := range y {
			// add opening div for row
			if iX == 0 {
				html += "<div class=row>"
			}

			// add divs for cells
			if iY == maze.start.y && iX == maze.start.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:gold' id='%d,%d'></div>", iY, iX))
			} else if iY == maze.end.y && iX == maze.end.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:red' id='%d,%d'></div>", iY, iX))
			} else if maze.matrix[iY][iX].obstacle {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:black' id='%d,%d'></div>", iY, iX))
			} else {
				html += template.HTML(fmt.Sprintf("<div class='cell' id='%d,%d'></div>", iY, iX))
			}

			// add closing div for row
			if iX+1 >= len(y) {
				html += "</div>"
			}
		}
	}

	data.HTML = html
	data.Path = maze.pathToEnd
	return data
}

func createAndSolveMaze(algoToUse string) *Maze {
	rand.Seed(time.Now().UnixNano())
	start := time.Now()

	fmt.Println("ALGO TO USE " + algoToUse)

	var maze *Maze = NewMaze(40)
	if algoToUse == "DFS" {
		maze = DFS(maze)
	} else if algoToUse == "AStar" {

		maze = AStar(maze)
	} else {
		maze = BFS(maze)
	}

	duration := time.Since(start)
	fmt.Println(duration.String())
	return maze
}

func canVisit(y int, x int, nodeString string, maze *Maze) bool {
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

func DFS(maze *Maze) *Maze {
	maze.stackToVisit.Push(maze.start)
	count := 0
	for len(maze.stackToVisit) > 0 {
		currNode, _ := maze.stackToVisit.Pop()
		currNodeString := currNode.ToString()
		fmt.Printf("Current node is: %s \n", currNodeString)
		if currNode.x == maze.end.x && currNode.y == maze.end.y {
			// check end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			fmt.Println("Reached the endnode!")
			break
		} else if maze.visitedSet.Contains(currNodeString) {
			// check visited
			fmt.Printf(currNodeString + " already visited!")
		} else {
			// check neighbors
			// check top cell
			if canVisit(currNode.y-1, currNode.x, currNodeString, maze) {
				maze.stackToVisit.Push(maze.matrix[currNode.y-1][currNode.x])
				//fmt.Printf("T - new Node = %d, %d \n", currNode.y-1, currNode.x)
			}
			// check left cell
			if canVisit(currNode.y, currNode.x-1, currNodeString, maze) {
				maze.stackToVisit.Push(maze.matrix[currNode.y][currNode.x-1])
				//fmt.Printf("L - new Node = %d, %d \n", currNode.y, currNode.x-1)
			}
			// check bottom cell
			if canVisit(currNode.y+1, currNode.x, currNodeString, maze) {
				maze.stackToVisit.Push(maze.matrix[currNode.y+1][currNode.x])
				//fmt.Printf("B - new Node = %d, %d \n", currNode.y+1, currNode.x)
			}
			// check right cell
			if canVisit(currNode.y, currNode.x+1, currNodeString, maze) {
				maze.stackToVisit.Push(maze.matrix[currNode.y][currNode.x+1])
				//fmt.Printf("R - new Node = %d, %d \n", currNode.y, currNode.x+1)
			}
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			count++
			fmt.Printf("Iteration # %d \n", count)
		}
	}
	return maze
}

func BFS(maze *Maze) *Maze {
	maze.queueToVisit = list.New()
	maze.queueToVisit.PushBack(maze.start)
	count := 0
	for maze.queueToVisit.Len() > 0 {
		front := maze.queueToVisit.Front()
		currNode := front.Value.(Node)
		maze.queueToVisit.Remove(front)
		currNodeString := currNode.ToString()
		fmt.Printf("Current node is: %s \n", currNodeString)
		if currNode.x == maze.end.x && currNode.y == maze.end.y {
			// check end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			fmt.Println("Reached the endnode!")
			break
		} else if maze.visitedSet.Contains(currNodeString) {
			// check visited
			fmt.Printf(currNodeString + " already visited!")
		} else {
			// check neighbors
			// check top cell
			if canVisit(currNode.y-1, currNode.x, currNodeString, maze) {
				maze.queueToVisit.PushBack(maze.matrix[currNode.y-1][currNode.x])
				//fmt.Printf("T - new Node = %d, %d \n", currNode.y-1, currNode.x)
			}
			// check left cell
			if canVisit(currNode.y, currNode.x-1, currNodeString, maze) {
				maze.queueToVisit.PushBack(maze.matrix[currNode.y][currNode.x-1])
				//fmt.Printf("L - new Node = %d, %d \n", currNode.y, currNode.x-1)
			}
			// check bottom cell
			if canVisit(currNode.y+1, currNode.x, currNodeString, maze) {
				maze.queueToVisit.PushBack(maze.matrix[currNode.y+1][currNode.x])
				//fmt.Printf("B - new Node = %d, %d \n", currNode.y+1, currNode.x)
			}
			// check right cell
			if canVisit(currNode.y, currNode.x+1, currNodeString, maze) {
				maze.queueToVisit.PushBack(maze.matrix[currNode.y][currNode.x+1])
				//fmt.Printf("R - new Node = %d, %d \n", currNode.y, currNode.x+1)
			}
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			count++
			fmt.Printf("Iteration # %d \n", count)
		}
	}

	return maze
}

func calculateDistance(node1 *Node, node2 *Node) float64 {
	return math.Sqrt(math.Pow(float64(node1.x-node2.x), 2) + math.Pow(float64(node1.y-node2.y), 2))
}

func processNeighbor(neighbor *Node, currNode *Node, maze *Maze) {
	fScoreThroughCurrentNode := currNode.fScore + calculateDistance(currNode, neighbor)
	if fScoreThroughCurrentNode < neighbor.fScore {
		neighbor.parent = currNode
		neighbor.fScore = fScoreThroughCurrentNode
	}
	neighbor.gScore = neighbor.fScore + calculateDistance(neighbor, &maze.end)
	maze.stackToVisit.Push(*neighbor)
}

// TODO: Improve A* implementation. ATM, gScore, fScore, and parent attributes are doing nothing
func AStar(maze *Maze) *Maze {
	maze.stackToVisit.Push(maze.start)
	count := 0
	for len(maze.stackToVisit) > 0 {
		sort.Slice(maze.stackToVisit, func(i, j int) bool {
			return calculateDistance(&maze.stackToVisit[i], &maze.end) > calculateDistance(&maze.stackToVisit[j], &maze.end)
		})
		currNode, _ := maze.stackToVisit.Pop()
		currNodeString := currNode.ToString()
		fmt.Printf("Current node is: %s \n", currNodeString)
		if currNode.x == maze.end.x && currNode.y == maze.end.y {
			// check end
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			fmt.Println("Reached the endnode!")
			break
		} else if maze.visitedSet.Contains(currNodeString) {
			// check visited
			fmt.Printf(currNodeString + " already visited!")
		} else {
			// check neighbors
			// check top cell
			if canVisit(currNode.y-1, currNode.x, currNodeString, maze) {
				processNeighbor(&maze.matrix[currNode.y-1][currNode.x], &currNode, maze)
				//fmt.Printf("T - new Node = %d, %d \n", currNode.y-1, currNode.x)
			}
			// check left cell
			if canVisit(currNode.y, currNode.x-1, currNodeString, maze) {
				processNeighbor(&maze.matrix[currNode.y][currNode.x-1], &currNode, maze)
				//fmt.Printf("L - new Node = %d, %d \n", currNode.y, currNode.x-1)
			}
			// check bottom cell
			if canVisit(currNode.y+1, currNode.x, currNodeString, maze) {
				processNeighbor(&maze.matrix[currNode.y+1][currNode.x], &currNode, maze)
				//fmt.Printf("B - new Node = %d, %d \n", currNode.y+1, currNode.x)
			}
			// check right cell
			if canVisit(currNode.y, currNode.x+1, currNodeString, maze) {
				processNeighbor(&maze.matrix[currNode.y][currNode.x+1], &currNode, maze)
				//fmt.Printf("R - new Node = %d, %d \n", currNode.y, currNode.x+1)
			}
			maze.visitedSet.Add(currNodeString)
			maze.pathToEnd = append(maze.pathToEnd, currNodeString)
			count++
			fmt.Printf("Iteration # %d \n", count)
			fmt.Println("In ASTAR")
		}
	}
	return maze
}
