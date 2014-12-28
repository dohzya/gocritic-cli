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

var tmplHeader = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Critic Markup Output</title>
	<style>
			#wrapper {
				padding-top: 30px !important;
			}
			#criticnav {
				position: fixed;
				top: 0;
				left: 0;
				width: 100%;
				box-shadow: 0 1px 1px 1px #777;
				margin: 0;
				padding: 0;
				background-color: white;
				font-size: 12px;
			}
			#criticnav ul {
				list-style-type: none;
				width: 90%;
				margin: 0 auto;
				padding: 0;
			}
			#criticnav ul li {
				display: block;
				width: 33%;
				text-align: center;
				padding: 10px 0 5px!important;
				margin: 0 !important;
				line-height: 1em;
				float: left;
				border-left: 1px solid #ccc;
				text-transform: uppercase;
			}
			#criticnav ul li:before {
				content: none !important;
			}
			#criticnav ul li#edited-button {
				border-right: 1px solid #ccc;
			}
			#criticnav ul li.active {
				background-image: -webkit-linear-gradient(top, white, #cccccc)
			}
			.original del {
				text-decoration: none;
			}
			.original ins,
			.original span.popover,
			.original ins.break {
				display: none;
			}
			.edited ins {
				text-decoration: none;
			}
			.edited del,
			.edited span.popover {
				display: none;
			}
			.original mark,
			.edited mark {
				background-color: transparent;
			}
			.markup mark {
				background-color: #fffd38;
				text-decoration: none;
			}
			.markup del {
				background-color: #f6a9a9;
				text-decoration: none;
			}
			.markup ins {
				background-color: #a9f6a9;
				text-decoration: none;
			}
			.markup ins.break {
				display: block;
				line-height: 2px;
				padding: 0 !important;
				margin: 0 !important;
			}
			.markup ins.break span {
				line-height: 1.5em;
			}
			.markup .popover {
				background-color: #4444ff;
				color: #fff;
			}
			.markup .popover .critic.comment {
				display: none;
			}
			.markup .popover:hover span.critic.comment {
				display: block;
				position: absolute;
				width: 200px;
				left: 30%;
				font-size: 0.8em;
				color: #ccc;
				background-color: #333;
				z-index: 10;
				padding: 0.5em 1em;
				border-radius: 0.5em;
			}
	}
	</style>
</head>
<body>
	<header id="criticnav">
		<ul>
			<li id="markup-button">Markup</li>
			<li id="original-button">Original</li>
			<li id="edited-button">Edited</li>
		</ul>
	</header>
	<div id="wrapper">`
var tmplFooter = `</div>
	<script>
		var wrapper = document.getElementById('wrapper');
		var btnMarkup = document.getElementById('markup-button');
		var btnOriginal = document.getElementById('original-button');
		var btnEditor = document.getElementById('edited-button');

		function unstate() {
			btnOriginal.className = '';
			btnEditor.className = '';
			btnMarkup.className = '';
		}

		function original() {
			unstate();
			btnOriginal.className = 'active';
			wrapper.className = 'original';
		}

		function edited() {
			unstate();
			btnEditor.className = 'active';
			wrapper.className = 'edited';
		}

		function markup() {
			unstate();
			btnMarkup.className = 'active';
			wrapper.className = 'markup';
		}

		function init() {
			markup();
			var inss = document.querySelectorAll('ins.break');
			for (var i=0; i < inss.length; i++) {
				var ins = inss[i];
				ins.innerHTML += "<br>";
			}
			var comments = document.querySelectorAll('span.critic.comment');
			for (i=0; i < comments.length; i++) {
				var comment = comments[i];
				var popover = document.createElement('span');
				popover.className = 'popover';
				wrapper.insertOriginal(popover, comment);
				wrapper.removeChild(comment);
				popover.innerHTML = '&#8225;';
				popover.appendChild(comment);
			}
		}

		var o = document.getElementById('original-button');
		var e = document.getElementById('edited-button');
		var m = document.getElementById('markup-button');

		window.onload = init;
		o.onclick = original;
		e.onclick = edited;
		m.onclick = markup;
	</script>
</body>
</html>
`

func main() {
	md := flag.Bool("md", false, "Use markdown parser")
	i := flag.String("i", "-", "Input file (default: STDIN)")
	o := flag.String("o", "-", "Output file (default: STDOUT)")
	original := flag.Bool("original", false, "Render original sources only")
	edited := flag.Bool("edited", false, "Render edited sources only")
	tags := flag.Bool("tags", false, "Keep tags")
	html := flag.Bool("html", false, "Create a full HTML page")
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
	if *original == *edited {
		filter = gocritic.FilterShowAll
	} else if *original {
		if *tags {
			filter = gocritic.FilterOnlyOriginal
		} else {
			filter = gocritic.FilterOnlyRawOriginal
		}
	} else {
		if *tags {
			filter = gocritic.FilterOnlyEdited
		} else {
			filter = gocritic.FilterOnlyRawEdited
		}
	}

	if *html {
		if _, err := output.Write([]byte(tmplHeader)); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error while writing header: %s\n", err.Error())
			return
		}
	}
	if *md {
		bMd := bytes.NewBuffer(make([]byte, 0))
		if _, err := gocritic.Critic(bMd, input, filter); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error during critic parsing: %s\n", err.Error())
			return
		}
		bHTML := blackfriday.MarkdownCommon(bMd.Bytes())
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
	if *html {
		if _, err := output.Write([]byte(tmplFooter)); err != nil {
			fmt.Fprintf(os.Stderr, "[gocritic] Error while writing footer: %s\n", err.Error())
			return
		}
	}
}
