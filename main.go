package main

import (
	"container/list"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	//go:embed "main.html"
	mapHTML     string
	mapTemplate = template.Must(template.New("maze").Parse(mapHTML))
)

type Point struct {
	x int
	y int
}

func NewPoint(y, x int) Point {
	point := Point{}
	point.y = y
	point.x = x
	return point
}

func (point *Point) ToString() string {
	return fmt.Sprintf("%d,%d", point.y, point.x)
}

type Maze struct {
	matrix       [][]bool
	queueToVisit *list.List
	stackToVisit Stack
	visitedSet   Set
	start        Point
	end          Point
	pathToEnd    []string
}

func NewMaze(size int) *Maze {
	var maze Maze
	maze.matrix = make([][]bool, size)
	for row := range maze.matrix {
		maze.matrix[row] = make([]bool, size)
	}
	maze.visitedSet = *NewSet()
	maze.start.x = rand.Intn(size)
	maze.start.y = rand.Intn(size)
	maze.end.x = rand.Intn(size)
	maze.end.y = rand.Intn(size)
	addObstacles(&maze)
	return &maze
}

func addObstacles(maze *Maze) {
	for iY, row := range maze.matrix {
		for iX, _ := range row {
			// don't put an obstacle on the starting or ending cells
			if (iY == maze.start.y && iX == maze.start.x) || (iY == maze.end.y && iX == maze.end.x) {
				continue
			} else if rand.Intn(10) < 3 {
				maze.matrix[iY][iX] = true
			}
		}
	}
}

type JSData struct {
	HTML template.HTML
	Path []string
}

type Set struct {
	hashmap map[Point]struct{}
}

func NewSet() *Set {
	set := &Set{}
	set.hashmap = make(map[Point]struct{})
	return set
}

func (set *Set) Add(value Point) {
	set.hashmap[value] = struct{}{}
}

func (set *Set) Remove(value Point) {
	delete(set.hashmap, value)
}

func (set *Set) Contains(value Point) bool {
	_, exists := set.hashmap[value]
	return exists
}

type Stack []Point

func (stack *Stack) IsEmpty() bool {
	return len(*stack) == 0
}

func (stack *Stack) Push(point Point) {
	*stack = append(*stack, point)
}

func (stack *Stack) Pop() (Point, bool) {
	if stack.IsEmpty() {
		return Point{}, false
	} else {
		index := len(*stack) - 1   // Get the index of the top most element.
		element := (*stack)[index] // Index into the slice and obtain the element.
		*stack = (*stack)[:index]  // Remove it from the stack by slicing it off.
		return element, true
	}
}

func (stack *Stack) Print() {
	for idx, cell := range *stack {
		fmt.Printf("%d #: %s \n", idx, cell.ToString())
	}
}

func main() {

	http.HandleFunc("/DFS", dfsHandler)

	http.HandleFunc("/BFS", bfsHandler)

	http.HandleFunc("/", dfsHandler)

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
			} else if maze.matrix[iY][iX] {
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

	var maze *Maze = NewMaze(40)
	if algoToUse == "DFS" {
		maze = DFS(maze)
	} else {
		maze = BFS(maze)
	}

	duration := time.Since(start)
	fmt.Println(duration.String())
	return maze
}

func canVisit(point *Point, maze *Maze) bool {
	if point.x < 0 || point.y < 0 || point.y >= len(maze.matrix) || point.x >= len(maze.matrix[point.y]) {
		return false
	}
	if maze.visitedSet.Contains(*point) {
		return false
	}
	// if is obstacle
	if maze.matrix[point.y][point.x] {
		return false
	}
	return true
}

func DFS(maze *Maze) *Maze {
	maze.stackToVisit.Push(maze.start)
	count := 0
	for len(maze.stackToVisit) > 0 {
		currPoint, _ := maze.stackToVisit.Pop()
		currPointString := currPoint.ToString()
		fmt.Printf("Current point is: %s \n", currPointString)
		if currPoint == maze.end {
			// check end
			maze.pathToEnd = append(maze.pathToEnd, currPointString)
			fmt.Println("Reached the endpoint!")
			break
		} else if maze.visitedSet.Contains(currPoint) {
			// check visited
			fmt.Printf(currPointString + " already visited!")
		} else {
			// check neighbors
			// check top cell
			topNeighbor := NewPoint(currPoint.y-1, currPoint.x)
			if canVisit(&topNeighbor, maze) {
				maze.stackToVisit.Push(topNeighbor)
				//fmt.Printf("T - new Point = %d, %d \n", currPoint.y-1, currPoint.x)
			}
			// check left cell
			leftNeighbor := NewPoint(currPoint.y, currPoint.x-1)
			if canVisit(&leftNeighbor, maze) {
				maze.stackToVisit.Push(leftNeighbor)
				//fmt.Printf("L - new Point = %d, %d \n", currPoint.y, currPoint.x-1)
			}
			// check bottom cell
			bottomNeighbor := NewPoint(currPoint.y+1, currPoint.x)
			if canVisit(&bottomNeighbor, maze) {
				maze.stackToVisit.Push(bottomNeighbor)
				//fmt.Printf("B - new Point = %d, %d \n", currPoint.y+1, currPoint.x)
			}
			// check right cell
			rightNeighbor := NewPoint(currPoint.y, currPoint.x+1)
			if canVisit(&rightNeighbor, maze) {
				maze.stackToVisit.Push(rightNeighbor)
				//fmt.Printf("R - new Point = %d, %d \n", currPoint.y, currPoint.x+1)
			}
			maze.visitedSet.Add(currPoint)
			maze.pathToEnd = append(maze.pathToEnd, currPointString)
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
		currPoint := front.Value.(Point)
		maze.queueToVisit.Remove(front)
		currPointString := currPoint.ToString()
		fmt.Printf("Current point is: %s \n", currPointString)
		if currPoint == maze.end {
			// check end
			maze.pathToEnd = append(maze.pathToEnd, currPointString)
			fmt.Println("Reached the endpoint!")
			break
		} else if maze.visitedSet.Contains(currPoint) {
			// check visited
			fmt.Printf(currPointString + " already visited!")
		} else {
			// check neighbors
			// check top cell
			topNeighbor := NewPoint(currPoint.y-1, currPoint.x)
			if canVisit(&topNeighbor, maze) {
				maze.queueToVisit.PushBack(topNeighbor)
				//fmt.Printf("T - new Point = %d, %d \n", currPoint.y-1, currPoint.x)
			}
			// check left cell
			leftNeighbor := NewPoint(currPoint.y, currPoint.x-1)
			if canVisit(&leftNeighbor, maze) {
				maze.queueToVisit.PushBack(leftNeighbor)
				//fmt.Printf("L - new Point = %d, %d \n", currPoint.y, currPoint.x-1)
			}
			// check bottom cell
			bottomNeighbor := NewPoint(currPoint.y+1, currPoint.x)
			if canVisit(&bottomNeighbor, maze) {
				maze.queueToVisit.PushBack(bottomNeighbor)
				//fmt.Printf("B - new Point = %d, %d \n", currPoint.y+1, currPoint.x)
			}
			// check right cell
			rightNeighbor := NewPoint(currPoint.y, currPoint.x+1)
			if canVisit(&rightNeighbor, maze) {
				maze.queueToVisit.PushBack(rightNeighbor)
				//fmt.Printf("R - new Point = %d, %d \n", currPoint.y, currPoint.x+1)
			}
			maze.visitedSet.Add(currPoint)
			maze.pathToEnd = append(maze.pathToEnd, currPointString)
			count++
			fmt.Printf("Iteration # %d \n", count)
		}
	}

	return maze
}
