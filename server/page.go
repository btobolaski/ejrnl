package server

import (
	"html/template"
)

const header = `{{define "header"}}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<script type="text/javascript" src="/autosize.js"></script>
	<title>{{.Title}}</title>
</head>
<body>
{{end}}`

const footer = `{{define "footer"}}
</body>
</html>
{{end}}`

const indexTemplate = `{{template "header" .}}
<p><a href="/entries/new">new</a></p>
<ul>
	{{ if .Entries }}{{range .Entries}}
	<li><a href="/entries/{{.Id}}/">{{.Date}}</a></li>
	{{end}}{{end}}
</ul>
{{template "footer"}}`

const formTemplate = `{{template "header" .}}
<form action="{{.Target}}" method="post">
	<textarea name="text" style="width: 100%;">{{.Text}}</textarea>
	<button type="submit">Save</button>
</form>
<script type="text/javascript">
autosize(document.querySelector('textarea'));
</script>
{{template "footer"}}`

var indexPage *template.Template
var formPage *template.Template

func init() {
	header, err := template.New("header").Parse(header)
	if err != nil {
		panic(err)
	}
	base, err := header.New("footer").Parse(footer)
	if err != nil {
		panic(err)
	}

	indexPage, err = base.New("index").Parse(indexTemplate)
	if err != nil {
		panic(err)
	}

	formPage, err = base.New("form").Parse(formTemplate)
	if err != nil {
		panic(err)
	}
}
