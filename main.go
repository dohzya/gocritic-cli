package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dohzya/gocritic"
	"github.com/russross/blackfriday"
)

func main() {
	md := flag.Bool("md", false, "Use markdown parser")
	i := flag.String("i", "-", "Input file (default: STDIN)")
	o := flag.String("o", "-", "Output file (default: STDOUT)")
	before := flag.Bool("before", false, "Return before only")
	after := flag.Bool("after", false, "Return after only")
	tags := flag.Bool("tags", false, "Keep tags")
	flag.Parse()

	var input io.Reader
	if *i == "" || *i == "-" {
		input = os.Stdin
	} else {
		file, err := os.Open(*i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Can't open %s: %s\n", *i, err.Error())
			return
		}
		input = file
	}
	var output io.Writer
	if *o == "" || *o == "-" {
		output = os.Stdout
	} else {
		file, err := os.Create(*o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Create create %s: %s\n", *o, err.Error())
			return
		}
		output = file
	}

	var filter func(*gocritic.Options)
	if *before == *after {
		filter = gocritic.FilterShowAll
	} else if *before {
		if *tags {
			filter = gocritic.FilterOnlyBefore
		} else {
			filter = gocritic.FilterOnlyRawBefore
		}
	} else {
		if *tags {
			filter = gocritic.FilterOnlyAfter
		} else {
			filter = gocritic.FilterOnlyRawAfter
		}
	}

	if *md {
		bMd := bytes.NewBuffer(make([]byte, 0))
		if _, err := gocritic.Critic(bMd, input, filter); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error during critic parsing: %s\n", err.Error())
			return
		}
		bHTML := blackfriday.MarkdownHtml(bMd.Bytes(), blackfriday.CommonExtensions)
		if _, err := output.Write(bHTML); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error while writing result: %s\n", err.Error())
			return
		}
	} else {
		if _, err := gocritic.Critic(output, input, filter); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error during critic parsing: %s\n", err.Error())
			return
		}
	}
}
