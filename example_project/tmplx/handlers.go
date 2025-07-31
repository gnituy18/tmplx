
package tmplx

import (
	"encoding/json"
	"html/template"
	"net/http"
)

type TmplxHandler struct {
        Url		string
	HandlerFunc 	http.HandlerFunc
}
var tmpl = template.Must(template.New("tmplx_handlers").Parse(`{{define "/{$}"}}
<html><head>



  <title> {{.field_1}} </title>
<script>
document.addEventListener('DOMContentLoaded', function() {
  const state = JSON.parse(this.getElementById("tx-state").innerHTML)
  const addHandler = (node) => {
    const walker = document.createTreeWalker(
      node,
      NodeFilter.SHOW_ELEMENT,
      (n) => {
        for (let attr of n.attributes) {
          if (attr.name.startsWith('tx-on')) {
            return NodeFilter.FILTER_ACCEPT;
          }
        }
        return NodeFilter.FILTER_SKIP
      }
    );
    while (walker.nextNode()) {
      const cn = walker.currentNode;
      for (let attr of cn.attributes) {
        if (attr.name.startsWith('tx-on')) {
          const eName = attr.name.slice(5);
          cn.addEventListener(eName, async () => {
            const states = {}

            for (let key in state) {
              states[key] = JSON.stringify(state[key])
            }
            const res = await fetch("/tx/" + attr.value + "?" + new URLSearchParams(states).toString())
            res.text().then(html => {
              document.open()
              document.write(html)
              document.close()
            })
          })
        }
      }
    }
  }

  new MutationObserver((records) => {
    records.forEach((record) => {
      if (record.type !== 'childList') return
      records.addedNodes()
    })
  }).observe(document.documentElement, { childList: true, subList: true })
  addHandler(document.documentElement)
});
</script><script type="application/json" id="tx-state">{{.state}}</script></head>

<body>
  <h1> {{.field_2}} </h1>
  
  <h2> Counter </h2><h2>
  <p>counter: {{.field_3}}</p>

  <button tx-onclick="index-addOne">Add 1</button>
  <button tx-onclick="index-subOne">Subtract 1</button>

  </h2><h2> Derived </h2><h2>
  <p>counter * 10 = {{.field_4}}</p>


</h2></body></html>
{{end}}
`))
var tmplxHandlers []TmplxHandler = []TmplxHandler{

{
	Url: "/{$}",
	HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		var title string = "Tmplx!"
var h1 string = "Hello, Tmplx!"
var counter int = 0
var counterTimes10 int = counter * 10
counterTimes10 = counter * 10

		tmpl.ExecuteTemplate(w, "/{$}", map[string]any{"field_1": title, "field_2": h1, "field_3": counter, "field_4": counterTimes10, "state": map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10}})
	},
},
{
	Url: "/tx/index-addOne",
	HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
var title string
json.Unmarshal([]byte(query.Get("title")), &title)
var h1 string
json.Unmarshal([]byte(query.Get("h1")), &h1)
var counter int
json.Unmarshal([]byte(query.Get("counter")), &counter)
var counterTimes10 int = counter * 10
counter++
counterTimes10 = counter * 10

		tmpl.ExecuteTemplate(w, "/{$}", map[string]any{"field_4": counterTimes10, "field_1": title, "field_2": h1, "field_3": counter, "state": map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10}})
	},
},
{
	Url: "/tx/index-subOne",
	HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
var title string
json.Unmarshal([]byte(query.Get("title")), &title)
var h1 string
json.Unmarshal([]byte(query.Get("h1")), &h1)
var counter int
json.Unmarshal([]byte(query.Get("counter")), &counter)
var counterTimes10 int = counter * 10
counter--
counterTimes10 = counter * 10

		tmpl.ExecuteTemplate(w, "/{$}", map[string]any{"field_1": title, "field_2": h1, "field_3": counter, "field_4": counterTimes10, "state": map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10}})
	},
},}
func Handlers() []TmplxHandler { return tmplxHandlers }