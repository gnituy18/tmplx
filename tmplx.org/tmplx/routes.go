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

type TxRoute struct {
	Pattern string
	Handler http.HandlerFunc
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

type state_docs struct {
}

func render_docs(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>Tmplx Guide</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <nav> <ul> <li><a href=\"#quick-start\">Quick Start</a></li> <li><a href=\"#essentials\">Essentials</a></li> </ul> </nav> <main> <h2 id=\"quick-start\">Quick Start</h2> <h3>1. Install tmplx</h3> <pre><code tx-ignore=\"\">$ go install github.com/gnituy18/tmplx@latest</code></pre> <h3>2. Create a new project</h3> <pre><code tx-ignore=\"\">$ mkdir my-tmplx-app\n$ cd my-tmplx-app\n$ go mod init my-tmplx-app\n\n$ mkdir pages\n$ touch pages/index.html\n\n$ mkdir tmplx   # generated code will go here</code></pre> <h3>3. Write your first page (pages/index.html)</h3> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n  &lt;head&gt;\n    &lt;title&gt;My Tmplx App&lt;/title&gt;\n\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      var name string = &#34;World&#34; // state\n      var greeting string = fmt.Sprintf(&#34;Hello, %s!&#34;, name) // derived\n    &lt;/script&gt;\n  &lt;/head&gt;\n\n  &lt;body&gt;\n    &lt;h1&gt;{greeting}&lt;/h1&gt;\n  &lt;/body&gt;\n&lt;/html&gt;</code></pre> <h3>4. Generate routes</h3> <pre><code tx-ignore=\"\">$ tmplx -out-file=tmplx/routes.go</code></pre> <p>This creates <code>tmplx/routes.go</code> with compiled handlers and render functions.</p> <h3>5. Add main.go</h3> <pre><code tx-ignore=\"\" class=\"language-go\">package main\n\nimport (\n\t&#34;log&#34;\n\t&#34;net/http&#34;\n\n\t&#34;my-tmplx-app/tmplx&#34;\n)\n\nfunc main() {\n\tfor _, r := range tmplx.Routes() {\n\t\thttp.HandleFunc(r.Pattern, r.Handler)\n\t}\n\tlog.Fatal(http.ListenAndServe(&#34;:8080&#34;, nil))\n}</code></pre> <h3>6. Run the app</h3> <pre><code tx-ignore=\"\">$ go run .</code></pre> <p>Open <a href=\"http://localhost:8080/\" target=\"_blank\">http://localhost:8080/</a>. You should see “Hello, World!” rendered server-side using Go logic.</p> <h2 id=\"essentials\">Essentials</h2> </main> </body></html>"))
}

type state_index struct {
}

func render_index(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <main> <h1 style=\"text-align: center\">&lt;tmplx&gt;</h1> <p style=\"text-align: center\"> <strong>Build state-driven web app with Go in HTML.</strong> </p> <div style=\"display:flex; gap:1rem; justify-content: center; text-align:center; margin-top:4rem\"> <a class=\"btn\" href=\"/docs\">Docs</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> <h2 style=\"margin-top: 4rem\">Demo: To Do App</h2> "))
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
	w.Write([]byte(" <h2 style=\"margin-top: 2rem\"> <a href=\"https://github.com/gnituy18/tmplx/blob/main/tmplx.org/components/todo.html\">Source Code:</a> </h2> <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\nvar list []string = []string{}\nvar item string = &#34;&#34;\n\nfunc add() {\n  list = append(list, item)\n  item = &#34;&#34;\n}\n\nfunc remove(i int) {\n  list = append(list[0:i], list[i+1:]...)\n}\n&lt;/script&gt;\n\n&lt;label&gt;New &lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n&lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n&lt;ol&gt;\n  &lt;li tx-for=&#34;i, l := range list&#34; tx-key=&#34;l&#34; tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code> </pre> <p>It allows you to build UIs with React-like reactivity purely in Go.</p> <h2 style=\"margin-top: 2rem\"> <a href=\"https://github.com/gnituy18/tmplx/blob/main/tmplx.org/tmplx/handler.go\">Compiled Code:</a> </h2> </main> </body></html>"))
}

var txRoutes []TxRoute = []TxRoute{
	{
		Pattern: "GET /docs",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			state := &state_docs{}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render_docs(&buf, "tx_", map[string]string{}, newStates)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /{$}",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
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
		Pattern: "GET /tx/tx_h_todo_add",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			query := tx_r.URL.Query()
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
		Pattern: "GET /tx/tx_h_todo_remove",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			query := tx_r.URL.Query()
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

func Routes() []TxRoute { return txRoutes }
