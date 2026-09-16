package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/alicebob/promgrph"
)

var (
	prom   = "http://localhost:9090"
	listen = "localhost:9091"
)

func main() {
	c := promgrph.NewClient(prom)

	m := http.NewServeMux()
	m.Handle("GET /graph.svg", c.MakeSVGHandler(
		"up",
		promgrph.GraphOpts{
			Title:   "<b>up</b>",
			Stacked: true,
			Legend:  "Up!<b>b",
		},
	))
	m.Handle("GET /alloc.svg", c.MakeSVGHandler(
		"go_memstats_alloc_bytes",
		promgrph.GraphOpts{
			Title:  "alloc",
			Legend: ">{{.instance}}<",
		},
	))
	m.Handle("GET /free.svg", c.MakeSVGHandler(
		"rate(go_memstats_frees_total[5m])",
		promgrph.GraphOpts{
			Title: "frees",
		},
	))
	m.HandleFunc("GET /", indexHandler(c))
	fmt.Printf("at: %s\n", listen)
	if err := (&http.Server{Handler: m, Addr: listen}).ListenAndServe(); err != nil {
		fmt.Printf("httpd: %s\n", err)
	}
}

func indexHandler(c *promgrph.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		template.Must(template.New("index").Funcs(c.TemplateFuncMap()).Parse(`
<html>
<head>
</head>
<body>
	{{graph "/graph.svg" (width 600)}}<br>
	{{graph "/graph.svg" (width 600) (period "24h")}}<br>
	{{graph "/alloc.svg" (width 800) (height 300) }}<br>
	{{graph "/free.svg" (width 400) }}<br>
</body>
</html>
`)).Execute(w, nil)
	}
}
