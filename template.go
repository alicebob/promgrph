package promgrph

import (
	"fmt"
	"html/template"
	"net/url"
	"time"
)

func (c *Client) TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"graph": c.templGraph,
		"width": func(n int) Option {
			return func(a *GraphArgs) { a.Width = n }
		},
		"height": func(n int) Option {
			return func(a *GraphArgs) { a.Height = n }
		},
		"period": func(n string) Option {
			return func(a *GraphArgs) {
				p, _ := time.ParseDuration(n) // FIXME: error
				a.Period = p
			}
		},
	}
}

type (
	Option func(*GraphArgs)

	GraphArgs struct {
		Width  int
		Height int
		Period time.Duration
	}
)

// Give it a plain path, "/graph.svg", without query arguments.
func (c *Client) templGraph(path string, opts ...Option) template.HTML {
	args := &GraphArgs{}
	for _, o := range opts {
		o(args)
	}

	var (
		u     = url.Values{}
		attrs = ""
	)
	if args != nil {
		if args.Width > 0 {
			attrs += fmt.Sprintf(` width="%d"`, args.Width)
			u.Set("width", fmt.Sprintf("%d", args.Width))
		}
		if args.Height > 0 {
			attrs += fmt.Sprintf(` height="%d"`, args.Height)
			u.Set("height", fmt.Sprintf("%d", args.Height))
		}
		if args.Period > 0 {
			u.Set("period", fmt.Sprintf("%s", args.Period))
		}
	}
	return template.HTML(fmt.Sprintf(`<img src="%s?%s"%s>`, path, u.Encode(), attrs))
}
