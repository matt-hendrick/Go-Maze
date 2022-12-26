package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
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

func (point *Point) ToString() string {
	return fmt.Sprintf("%d,%d", point.y, point.x)
}

type Maze struct {
	matrix       [][]uint8
	visitedStack Stack
	visitedSet   Set
	start        Point
	end          Point
	pathToEnd    []string
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
	start := time.Now()

	var mazeToRun Maze = generateMaze(40, 40)
	var completedMaze Maze = DFS(mazeToRun)

	duration := time.Since(start)
	fmt.Println(duration.String())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := mapTemplate.Execute(w, getHTML(completedMaze)); err != nil {
			log.Fatal(err)
		}
	})

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func generateMaze(height, width int) Maze {
	if height < 5 || width < 5 {
		panic("Maze too small for current implementation")
	}
	var maze Maze
	maze.matrix = make([][]uint8, height)
	for row := range maze.matrix {
		maze.matrix[row] = make([]uint8, width)
	}
	maze.visitedSet = *NewSet()
	maze.start.x = 0
	maze.start.y = 0
	maze.end.x = 20
	maze.end.y = 13
	return maze
}

func DFS(maze Maze) Maze {
	maze.visitedStack.Push(maze.start)
	count := 0
	for len(maze.visitedStack) > 0 {
		currPoint, _ := maze.visitedStack.Pop()
		currPointString := currPoint.ToString()
		fmt.Printf("Current point is: %s \n", currPointString)
		if currPoint == maze.end {
			// check end
			fmt.Println("Reached the endpoint!")
			break
		} else if maze.visitedSet.Contains(currPoint) {
			// check visited
			fmt.Printf(currPointString + " already visited!")
		} else {
			// check next
			// right cell is valid
			rightNeighbor := Point{currPoint.y, currPoint.x + 1}
			if currPoint.x+1 < len(maze.matrix) && !maze.visitedSet.Contains(rightNeighbor) {
				maze.visitedStack.Push(rightNeighbor)
				fmt.Printf("R - new X = %d \n", currPoint.x+1)
			}
			// left cell is valid
			leftNeighbor := Point{currPoint.y, currPoint.x - 1}
			if currPoint.x-1 >= 0 && !maze.visitedSet.Contains(leftNeighbor) {
				maze.visitedStack.Push(leftNeighbor)
				fmt.Printf("L - new X = %d \n", currPoint.x+1)
			}
			// top cell is valid
			topNeighbor := Point{currPoint.y - 1, currPoint.x}
			if currPoint.y-1 >= 0 && !maze.visitedSet.Contains(topNeighbor) {
				maze.visitedStack.Push(topNeighbor)
				fmt.Printf("T - new Y = %d \n", currPoint.x+1)
			}
			// bottom cell is valid
			bottomNeighbor := Point{currPoint.y + 1, currPoint.x}
			if currPoint.y+1 < len(maze.matrix) && !maze.visitedSet.Contains(bottomNeighbor) {
				maze.visitedStack.Push(bottomNeighbor)
				fmt.Printf("B - new Y = %d \n", currPoint.y+1)
			}
			maze.visitedSet.Add(currPoint)
			maze.pathToEnd = append(maze.pathToEnd, currPointString)
			count++
			fmt.Printf("Iteration # %d \n", count)
			//maze.visitedStack.Print()
		}
	}
	return maze
}

func BFS() {
	return
}

// gets HTML to inject into main.html
func getHTML(maze Maze) JSData {

	var data JSData
	var html template.HTML = ""

	for iY, y := range maze.matrix {
		for iX, _ := range y {
			if iX == 0 {
				fmt.Println("add beginning of row", iX)
				html += "<div class=row>"
			}

			fmt.Println("add cell", iX)
			if iY == maze.start.y && iX == maze.start.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:gold' id='%d, %d'></div>", iY, iX))
			} else if iY == maze.end.y && iX == maze.end.x {
				html += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:red' id='%d, %d'></div>", iY, iX))
			} else {
				html += template.HTML(fmt.Sprintf("<div class='cell' id='%d, %d'></div>", iY, iX))
			}

			if iX+1 >= len(y) {
				fmt.Println("add closing of row", iX, iY)
				html += "</div>"
			}
		}
	}

	data.HTML = html
	data.Path = maze.pathToEnd
	return data
}
