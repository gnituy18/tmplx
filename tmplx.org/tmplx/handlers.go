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

func render_index(w io.Writer, state string) {
	w.Write([]byte(`<!DOCTYPE html><html><head>


  
  <script id="tx-runtime">`))
	w.Write([]byte(runtimeScript))
	w.Write([]byte(`</script><script type="application/json" id="tx-state">`))
	w.Write([]byte(fmt.Sprint(state)))
	w.Write([]byte(`</script></head>
  <body>
    <h1> gg </h1>
  

</body></html>`))
}
func Handlers() []TmplxHandler { return tmplxHandlers }

var tmplxHandlers []TmplxHandler = []TmplxHandler{

	{
		Url: "/{$}",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {

			stateBytes, _ := json.Marshal(map[string]any{})
			state := string(stateBytes)
			render_index(w, state)

		},
	},
}
