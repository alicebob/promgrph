package main

import (
	"fmt"
	"net/http"

	"github.com/alicebob/promgrph"
)

var (
	prom   = "http://localhost:9090"
	listen = "localhost:9091"
)

func main() {
	c := promgrph.NewClient(prom)

	m := http.NewServeMux()
	m.Handle("GET /graph.svg", c.MakeSVGHandler("up"))
	m.HandleFunc("GET /", indexHandler)
	fmt.Printf("at: %s\n", listen)
	if err := (&http.Server{Handler: m, Addr: listen}).ListenAndServe(); err != nil {
		fmt.Printf("httpd: %s\n", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
<html>
<head>
</head>
<body>
	<img src="./graph.svg">
</body>
</html>
`))
}
