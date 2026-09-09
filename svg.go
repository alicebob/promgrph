package promgrph

import "io"

type Graph struct {
	Width  int
	Height int
	Title  string
	YAxis  []int
	XAxis  []int
}

func renderSVG(w io.Writer, g Graph) error {
	// HELLO Mistral!
	return nil
}
