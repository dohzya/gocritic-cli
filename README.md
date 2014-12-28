gocritic-cli
============

gocritic-cli is a CLI for [gocritic](https://github.com/dohzya/gocritic), a [Go](https://golang.org) library for [CriticMarkup](http://criticmarkup.com).

The main purpose of this tool is to render CriticMarkup, but Markdown is supported as well, thanks to the [Blackfriday](https://github.com/russross/blackfriday) library.

Use
---

Basic use:

```bash
gocritic-cli <input-file> -o <output-file>
```

As filter:

```bash
cmd | gocritic-cli | othercmd
```

Render markdown

```bash
gocritic-cli -md
```

Render only original source

```bash
gocritic-cli -original
```

Render only edited source

```bash
gocritic-cli -edited
```

Render original + `<del>` tags and comments

```bash
gocritic-cli -original -tags
```

Render edited + `<ins>` tags and comments

```bash
gocritic-cli -edited -tags
```

Render a full HTML page which allows to display critic/original/edited sources:

```bash
gocritic-cli -html
```
