package tmplx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
  let state = JSON.parse(this.getElementById("tx-state").innerHTML)
  let tasks = [];
  let isProcessing = false;

  const init = (cn) => {
    for (let attr of cn.attributes) {
      if (attr.name === 'tx-value' && cn.tagName === 'INPUT') {
        cn.addEventListener('input', (e) => {
          const txSwap = cn.getAttribute("tx-swap")
          const txValue = cn.getAttribute("tx-value")
          if (txValue) {
            state[txSwap][txValue] = e.target.value
            return
          }
        })
      } else if (attr.name.startsWith('tx-on')) {
        const [fun, params] = attr.value.split("?")
        const searchParams = new URLSearchParams(params)
        let eventName = attr.name.slice(5)

        cn.addEventListener(eventName, () => {
          tasks.push(async () => {
            const txSwap = cn.getAttribute("tx-swap")
            searchParams.append("tx-swap", txSwap)


            for (let key in state) {
              if (key.startsWith(txSwap)) {
                searchParams.append(key, JSON.stringify(state[key]))
              }
            }

            const res = await fetch("/tx/" + fun + "?" + searchParams.toString())
            const html = await res.text()

            if (txSwap === 'tx_') {
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
          processQueue()
        })
      }
    }
  }

  async function processQueue() {
    if (isProcessing) return;
    isProcessing = true;
    while (tasks.length > 0) {
      const task = tasks.shift();
      await task();
    }
    isProcessing = false;
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
          if (attr.name.startsWith('tx-on') || attr.name === 'tx-value') {
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
type state_tx_h_todo struct {
	S_list []string `json:"list"`
	S_item string   `json:"item"`
}

func render_tx_h_todo(w io.Writer, key string, states map[string]string, newStates map[string]any, list []string, item string, add, add_swap string, remove, remove_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <label>New <input type=\"text\" tx-value=\"item\" tx-swap=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"value=\""))
	fmt.Fprint(w, item)
	w.Write([]byte("\"/></label> <button tx-onclick=\""))
	fmt.Fprint(w, add)
	w.Write([]byte("\" tx-swap=\""))
	fmt.Fprint(w, add_swap)
	w.Write([]byte("\">Add</button> <ol> "))

	for i, l := range list {
		w.Write([]byte("<li tx-key=\"l\" tx-onclick=\""))
		fmt.Fprint(w, remove)
		w.Write([]byte("?i="))
		if param, err := json.Marshal(i); err != nil {
			log.Panic(err)
		} else {
			w.Write([]byte(url.QueryEscape(string(param))))
		}
		w.Write([]byte("\" tx-swap=\""))
		fmt.Fprint(w, remove_swap)
		w.Write([]byte("\"> "))
		w.Write([]byte(html.EscapeString(fmt.Sprint(l))))
		w.Write([]byte(" </li>"))

	}
	w.Write([]byte(" </ol> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_index struct {
}

func render_index(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title> tmplx </title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/languages/html.min.js\"></script> <script>hljs.highlightAll();</script> <style>\n    body {\n      color: #16161d;\n      background: #efeff3;\n      display: flex;\n    }\n\n    main {\n      margin: 4rem auto;\n      width: 40rem;\n    }\n\n    h1 {\n      font-family: Verdana;\n      font-size: 3rem;\n    }\n\n    h2 {\n      margin-top: 2rem;\n      margin-bottom: 1rem;\n    }\n\n    a {\n      color: RoyalBlue;\n    }\n\n    pre {\n      margin: 0;\n    }\n\n    .btn {\n      text-decoration: none;\n      border: solid;\n      border-radius: 0.25rem;\n      color: #16161d;\n      padding: 0.5rem 2rem;\n      font-weight: bold;\n      background-color: LightGray;\n\n      &:hover {\n        background-color: Silver;\n      }\n    }\n\n    li {\n      cursor: pointer;\n\n      &:hover {\n        text-decoration: line-through\n      }\n    }\n  </style> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <main> <h1 style=\"text-align:center\"> &lt;tmplx&gt; </h1> <p style=\"text-align:center\"> <strong>Build state-driven web app with Go in HTML.</strong></p> <div style=\"text-align:center;margin-top:4rem\"> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">Get Started </a> </div> <h2 style=\"margin-top:4rem\"> Demo: To Do App</h2> "))
	{
		ckey := key + "_index_tx-todo_1"
		state := &state_tx_h_todo{}
		if _, ok := states[ckey]; ok {
			json.Unmarshal([]byte(states[ckey]), state)
			newStates[ckey] = state
		} else {
			state.S_list = []string{}
			state.S_item = ""
			newStates[ckey] = state
		}
		list := state.S_list
		item := state.S_item
		render_tx_h_todo(w, ckey, states, newStates, list, item, "tx_h_todo_add", ckey, "tx_h_todo_remove", ckey)
	}
	w.Write([]byte(" <h2 style=\"margin-top:2rem\"><a href=\"https://github.com/gnituy18/tmplx/blob/main/tmplx.org/components/todo.html\">Source Code:</a></h2> <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n        var list []string = []string{}\n        var item string = &#34;&#34;\n\n        func add() {\n        list = append(list, item)\n        item = &#34;&#34;\n        }\n\n        func remove(i int) {\n        list = append(list[0:i], list[i+1:]...)\n        }\n        &lt;/script&gt;\n\n        &lt;label&gt;New &lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n\n        &lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n\n        &lt;ol&gt;\n        &lt;li tx-for=&#34;i, l := range list&#34; tx-key=&#34;l&#34; tx-onclick=&#34;remove(i)&#34;&gt;\n        { l }\n        &lt;/li&gt;\n        &lt;/ol&gt;</code> </pre> <p>It allows you to build UIs with React-like reactivity purely in Go.</p> <h2 style=\"margin-top:2rem\"><a href=\"https://github.com/gnituy18/tmplx/blob/main/tmplx.org/tmplx/handler.go\">Compiled Code:</a></h2> </main> </body></html>"))
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
		Url: "/tx/tx_h_todo_add",
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
			state := &state_tx_h_todo{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			list := state.S_list
			item := state.S_item
			list = append(list, item)
			item = ""
			render_tx_h_todo(w, txSwap, states, newStates, list, item, "tx_h_todo_add", txSwap, "tx_h_todo_remove", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_todo{
				S_list: list,
				S_item: item,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Url: "/tx/tx_h_todo_remove",
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
			state := &state_tx_h_todo{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			list := state.S_list
			item := state.S_item
			var i int
			json.Unmarshal([]byte(query.Get("i")), &i)
			list = append(list[0:i], list[i+1:]...)
			render_tx_h_todo(w, txSwap, states, newStates, list, item, "tx_h_todo_add", txSwap, "tx_h_todo_remove", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_todo{
				S_list: list,
				S_item: item,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
}

func Handlers() []TmplxHandler { return tmplxHandlers }
