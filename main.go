package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

var (
	//go:embed "main.html"
	mapHTML     string
	mapTemplate = template.Must(template.New("maze").Parse(mapHTML))
)

// type Maze struct {
// 	matrix [][]uint8
// }

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := mapTemplate.Execute(w, getHTML()); err != nil {
			log.Fatal(err)
		}
	})

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func getHTML() template.HTML {
	maze := [40][40]uint8{}

	var data template.HTML = ""

	for iY, y := range maze {
		for iX, _ := range y {
			if iX == 0 {
				fmt.Println("add begin row", iX)
				data += "<div class=row>"
			}

			fmt.Println("add cell", iX)
			if iY == 0 && iX == 0 {
				data += template.HTML(fmt.Sprintf("<div class='cell' style='background-color:red' id='%d, %d'></div>", iY, iX))
			} else {
				data += template.HTML(fmt.Sprintf("<div class='cell' id='%d, %d'></div>", iY, iX))
			}

			if iX+1 >= len(y) {
				fmt.Println("close row", iX, iY)
				data += "</div>"
			}
		}
	}

	return data
}
