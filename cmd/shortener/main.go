package main

import (
	"fmt"
	"github.com/romanp1989/go-shortener/internal/app"
	"html/template"
	"os"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// BuildData хранит данные о текущей версии, дате, хэш коммите
type BuildData struct {
	BuildVersion string
	BuildDate    string
	BuildCommit  string
}

// Template содержит шаблон для вывода информации о сборке
const Template = `Build version: {{if .BuildVersion}} {{.BuildVersion}} {{else}} N/A {{end}}
Build date: {{if .BuildDate}} {{.BuildDate}} {{else}} N/A {{end}}
Build commit: {{if .BuildCommit}} {{.BuildCommit}} {{else}} N/A {{end}}
`

// main Main function for launch application
func main() {
	printBuildInfo()
	app.RunServer()
}

func printBuildInfo() {
	bData := BuildData{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}
	templ := template.Must(template.New("buildTags").Parse(Template))
	err := templ.Execute(os.Stdout, bData)
	if err != nil {
		fmt.Printf("error: %v\n\n", err)
		return
	}
}
