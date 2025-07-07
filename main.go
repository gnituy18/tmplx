package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
)

type State map[string]any

func (s State) String() string {
	bs, _ := json.Marshal(s)
	return string(bs)
}

func main() {

	tmpl := template.Must(template.ParseFiles("index.html"))
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		input := ""
		appName := "Todo App powered by tmplx"
		list := []string{"init"}
		listState, _ := json.Marshal(list)

		state := State{
			"input":     input,
			"appName":   appName,
			"list":      list,
			"listState": string(listState),
		}

		if err := tmpl.Execute(w, state); err != nil {
			log.Panic(err)
		}

	})

	http.HandleFunc("GET /add/{$}", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()

		input := ""
		input = query.Get("input")

		appName := ""
		appName = query.Get("appName")

		list := []string{"init"}
		json.Unmarshal([]byte(query.Get("list")), &list)
		list = append(list, input)

		listState, _ := json.Marshal(list)

		state := State{
			"input":     input,
			"appName":   appName,
			"list":      list,
			"listState": string(listState),
		}

		tmpl := template.Must(template.New("add").Parse(`

<div id="tx-state">
  <data id="tx-state-appName" value="{{ .appName }}"></data>
  <data id="tx-state-input" value="{{ .input }}"></data>
  <data id="tx-state-list" value="{{ .listState }}"></data>
</div>

<h1 class="chunk-a" hx-swap-oob="outerHTML:.chunk-a"> {{ .appName }} </h1>

<div hx-swap-oob="outerHTML:.chunk-b" class="chunk-b">
{{range $i, $item := .list}}
<div hx-swap="outerHTML" hx-get='/delete/?i={{ $i }}' hx-target="#tx-state"> {{ $item }} </div>
{{end}}
</div>
		`))
		tmpl.Execute(w, state)

	})

	http.HandleFunc("GET /delete/{$}", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()

		input := ""
		input = query.Get("input")

		appName := ""
		appName = query.Get("appName")

		list := []string{"init"}
		json.Unmarshal([]byte(query.Get("list")), &list)

		iStr := query.Get("i")
		i, _ := strconv.Atoi(iStr)
		list = slices.Delete(list, i, i+1)

		listState, _ := json.Marshal(list)

		state := State{
			"input":     input,
			"appName":   appName,
			"list":      list,
			"listState": string(listState),
		}

		tmpl := template.Must(template.New("add").Parse(`

<div id="tx-state">
  <data id="tx-state-appName" value="{{ .appName }}"></data>
  <data id="tx-state-input" value="{{ .input }}"></data>
  <data id="tx-state-list" value="{{ .listState }}"></data>
</div>

<h1 class="chunk-a" hx-swap-oob="outerHTML:.chunk-a"> {{ .appName }} </h1>

<div hx-swap-oob="outerHTML:.chunk-b" class="chunk-b">
{{range $i, $item := .list}}
<div hx-swap="outerHTML" hx-get='/delete/?i={{ $i }}' hx-target="#tx-state"> {{ $item }} </div>
{{end}}
</div>
		`))
		tmpl.Execute(w, state)

	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
