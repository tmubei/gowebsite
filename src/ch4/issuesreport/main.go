package issuesreport

import (
	"log"
	"os"
	"text/template"
	"time"
	"ch4/github"
)

//!+template
const templ = `{{.TotalCount}} issues:
{{range .Items}}
	Number:	{{.Number}}
	User:	{{.User.Login}}
	Title:	{{.Title |	printf "%s.64s"}}
	Age:	{{.CreateAt | daysAgo}}	days
{{end}}`

//!-template

//!+daysAgo
func daysAgo(t time.Time) int {
	return int(time.Since(t).Hours() / 24)
}

//!-daysAge

//!+exec
var report = template.Must(template.New("issuelist").
	Funcs(template.FuncMap{"daysAgo": daysAgo}).Parse(templ))

func main() {
	result, err := github.SearchIssues(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := report.Execute(os.Stdout, result); err != nil {
		log.Fatal(err)
	}
}

//!-exec

func noMust() {
	//!+parse
	report, err := template.New("report").Funcs(template.FuncMap{"daysAgo":daysAgo}).Parse(templ)
	if err != nil {
		log.Fatal(err)
	}
	//!-parse
	result, err := github.SearchIssues(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := report.Execute(os.Stdout, result); err != nil {
		log.Fatal(err)
	}
}

/*
//!+output
$ go build gopl.io/ch4/issuesreport
$ ./issuesreport repo:golang/go is:open json decoder
13 issues:
-----------------------------
Number: 5680
User: eaigner
Title: encoding/json: set key converter on en/decoder
Age: 750 days
-----------------------------
Number: 6050
User: gopherbot
Title: encoding/json: provide tokenizer
Age: 695 days
-----------------------------

...
//!-output
 */