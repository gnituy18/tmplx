package tmplx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TmplxHandler struct {
	Url         string
	HandlerFunc http.HandlerFunc
}

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
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
`

func render_index(w io.Writer, state string, title string, h1 string, counter int, counterTimes10 int) {
	w.Write([]byte(`<html><head>




  <title> `))
	w.Write([]byte(fmt.Sprint(title)))
	w.Write([]byte(` </title>
<script id="tx-runtime">`))
	w.Write([]byte(runtimeScript))
	w.Write([]byte(`</script><script type="application/json" id="tx-state">`))
	w.Write([]byte(fmt.Sprint(state)))
	w.Write([]byte(`</script></head>

<body>
  <h1> `))
	w.Write([]byte(fmt.Sprint(h1)))
	w.Write([]byte(` </h1>

  <p>counter: `))
	w.Write([]byte(fmt.Sprint(counter)))
	w.Write([]byte(`</p>

  <button tx-onclick="index_addOne">Add 1</button>
  <button tx-onclick="index_subOne">Subtract 1</button>

  <p>counter * 10 = `))
	w.Write([]byte(fmt.Sprint(counterTimes10)))
	w.Write([]byte(`</p>

  <a href="/second-page">second page</a>



</body></html>`))
}
func render_second_d_page(w io.Writer, state string) {
	w.Write([]byte(`<html><head>
  <title> `))
	w.Write([]byte(fmt.Sprint(1 + 2)))
	w.Write([]byte(` </title>
<script id="tx-runtime">`))
	w.Write([]byte(runtimeScript))
	w.Write([]byte(`</script><script type="application/json" id="tx-state">`))
	w.Write([]byte(fmt.Sprint(state)))
	w.Write([]byte(`</script></head>

<body>
  <h1> `))
	w.Write([]byte(fmt.Sprint(fmt.Sprintf("a + b = %d", 1+2))))
	w.Write([]byte(` </h1>


</body></html>`))
}
func Handlers() []TmplxHandler { return tmplxHandlers }

var tmplxHandlers []TmplxHandler = []TmplxHandler{

	{
		Url: "/{$}",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			var title string = "Tmplx"
			var h1 string = "Hello, Tmplx!"
			var counter int = 0
			var counterTimes10 int = counter * 10
			counterTimes10 = counter * 10

			stateBytes, _ := json.Marshal(map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10})
			state := string(stateBytes)
			render_index(w, state, title, h1, counter, counterTimes10)

		},
	},
	{
		Url: "/tx/index_addOne",
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

			stateBytes, _ := json.Marshal(map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10})
			state := string(stateBytes)
			render_index(w, state, title, h1, counter, counterTimes10)

		},
	},
	{
		Url: "/tx/index_subOne",
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

			stateBytes, _ := json.Marshal(map[string]any{"title": title, "h1": h1, "counter": counter, "counterTimes10": counterTimes10})
			state := string(stateBytes)
			render_index(w, state, title, h1, counter, counterTimes10)

		},
	},
	{
		Url: "/second-page",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {

			stateBytes, _ := json.Marshal(map[string]any{})
			state := string(stateBytes)
			render_second_d_page(w, state)

		},
	},
}
