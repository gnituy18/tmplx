package tmplx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
)

var runtimeScript = `
document.addEventListener('DOMContentLoaded', function() {
  let state = JSON.parse(this.getElementById("tx-state").innerHTML)

  const init = (cn) => {
    for (let attr of cn.attributes) {
      if (attr.name.startsWith('tx-on')) {
        const [fun, params] = attr.value.split("?")
        const searchParams = new URLSearchParams(params)
        const eventName = attr.name.slice(5);
        cn.addEventListener(eventName, async () => {
          const txSwap = cn.getAttribute("tx-swap")
          let pfx = ''
          if (txSwap) {
            pfx = txSwap
          }
          for (let key in state) {
            if (key.startsWith(pfx)) {
              searchParams.append(key, JSON.stringify(state[key]))
            }
          }
          searchParams.append("tx-swap", txSwap)
          const res = await fetch("/tx/" + fun + "?" + searchParams.toString())
          res.text().then(html => {
            if (pfx === '') {
              document.open()
              document.write(html)
              document.close()
              return
            }

            const comp = document.createElement('body')
            comp.innerHTML = html
            const txState = comp.querySelector("#tx-state")
            const newStates = JSON.parse(txState.textContent)
            state = { ...state, ...newStates }
            comp.removeChild(txState)
            const range = document.createRange()
            const start = document.getElementById(txSwap)
            const end = document.getElementById(txSwap + '_e')
            range.setStartBefore(start);
            range.setEndAfter(end);
            range.deleteContents();
            for (let child of comp.childNodes) {
              range.insertNode(child.cloneNode(true))
              range.collapse(false)
            }
          })
        })
      }
    }
  }

  const addHandler = (node) => {
    if (node.nodeType === Node.TEXT_NODE) {
      return
    }

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

    init(walker.root)
    while (walker.nextNode()) {
      init(walker.currentNode)
    }
  }

  new MutationObserver((records) => {
    records.forEach((record) => {
      if (record.type !== 'childList') return
      record.addedNodes.forEach(addHandler)
    })
  }).observe(document.documentElement, { childList: true, subtree: true })
  addHandler(document.documentElement)
});
`

type TmplxHandler struct {
	Url         string
	HandlerFunc http.HandlerFunc
}
type state_tx_h_demo struct {
	S_name     string `json:"name"`
	S_greeting string `json:"greeting"`
	S_counter  int    `json:"counter"`
}

func render_tx_h_demo(w io.Writer, key string, states map[string]string, newStates map[string]any, name string, greeting string, counter int, addOne, addOne_swap string) {
	w.Write([]byte(`<template id="`))
	w.Write([]byte(fmt.Sprint(key)))
	w.Write([]byte(`"></template> <h2>`))
	w.Write([]byte(html.EscapeString(fmt.Sprint(greeting))))
	w.Write([]byte(`</h2> <button tx-onclick="`))
	w.Write([]byte(fmt.Sprint(addOne)))
	w.Write([]byte(`"`))
	if addOne_swap != "tx_" {
		w.Write([]byte(` tx-swap="`))
		w.Write([]byte(fmt.Sprint(addOne_swap)))
		w.Write([]byte(`"`))
	}
	w.Write([]byte(`>Add 1</button> <p>count: `))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte(`</p> <p>count * 10 = `))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter * 10))))
	w.Write([]byte(`</p> <template id="`))
	w.Write([]byte(fmt.Sprint(key + "_e")))
	w.Write([]byte(`"></template>`))
}

type state_index struct {
}

func render_index(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte(`<!DOCTYPE html><html lang="en"><head> <title>a tmplx</title> <meta charset="UTF-8"/> <meta name="viewport" content="width=device-width, initial-scale=1.0"/> <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css"/> <link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css"/> <script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js"></script> <script src="https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/languages/html.min.js"></script> <script>hljs.highlightAll();</script> <style>
    body {
      color: #16161d;
      background: #efeff3;
      display: flex;
      font-size: 1rem;
      line-height: 1.5rem;
    }

    main {
      margin-left: auto;
      margin-right: auto;
      width: 40rem;
    }

    h1 {
      margin-top: 8rem;
      font-family: Verdana;
      font-size: 4rem;
    }

    mark {
      background: Cornsilk;
    }

    .btn {
      background: DimGray;
      color: white;
      text-decoration: none;
      padding: 0.5rem 1rem;
      border-radius: 0.5rem;
      font-size: 1.5rem;

      &:hover {
        background: gray;
      }
    }
  </style> <script id="tx-runtime">`))
	w.Write([]byte(runtimeScript))
	w.Write([]byte(`</script><script type="application/json" id="tx-state">TX_STATE_JSON</script></head> <body> <main> <h1 style="text-align:center">&lt;tmplx&gt;</h1> <p style="font-size:1.5rem;text-align:center"> <strong>Build state-driven web app with Go in HTML.</strong></p> <div style="text-align:center;padding:4rem 0;"> <a class="btn" href="https://github.com/gnituy18/tmplx">Get Started</a> </div> <pre> <code tx-ignore="" class="language-html">&lt;script type=&#34;text/tmplx&#34;&gt;
  // name is a state
  var name string = &#34;tmplx&#34;

  // greeting is a derived
  var greeting string = fmt.Sprintf(&#34;Hello ,%s!&#34;, name)

  var counter int = 0

  // addOne event handler
  func addOne() {
    counter++
  }
&lt;/script&gt;

&lt;h2&gt; { greeting } &lt;/h2&gt;
&lt;button tx-onclick=&#34;addOne()&#34;&gt;Add 1&lt;/button&gt;
&lt;p&gt;count: { counter }&lt;/p&gt;
&lt;p&gt;count * 10 = { counter * 10 }&lt;/p&gt;
      </code> </pre> `))
	{
		ckey := key + "_index_tx-demo_1"
		state := &state_tx_h_demo{}
		if _, ok := states[ckey]; ok {
			json.Unmarshal([]byte(states[ckey]), state)
			newStates[ckey] = state
		} else {
			state.S_name = "tmplx"
			state.S_counter = 0
			newStates[ckey] = state
		}
		name := state.S_name
		greeting := fmt.Sprintf("Hello ,%s!", name)
		counter := state.S_counter
		render_tx_h_demo(w, ckey, states, newStates, name, greeting, counter, "tx_h_demo_addOne", ckey)
	}
	w.Write([]byte(` </main> </body></html>`))
}

var tmplxHandlers []TmplxHandler = []TmplxHandler{
	{
		Url: "/{$}",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			state := &state_index{}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render_index(&buf, "tx_", map[string]string{}, newStates)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Url: "/tx/tx_h_demo_addOne",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			query := r.URL.Query()
			txSwap := query.Get("tx-swap")
			states := map[string]string{}
			for k, v := range query {
				if strings.HasPrefix(k, txSwap) {
					states[k] = v[0]
				}
			}
			newStates := map[string]any{}
			state := &state_tx_h_demo{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			name := state.S_name
			greeting := fmt.Sprintf("Hello ,%s!", name)
			counter := state.S_counter
			counter++
			greeting = fmt.Sprintf("Hello ,%s!", name)
			render_tx_h_demo(w, txSwap, states, newStates, name, greeting, counter, "tx_h_demo_addOne", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_demo{
				S_name:    name,
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
}

func Handlers() []TmplxHandler { return tmplxHandlers }
