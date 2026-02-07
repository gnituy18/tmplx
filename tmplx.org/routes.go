package main

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
	"time"
)

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
  const handlerPrefix = "/tx/"
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

            const res = await fetch(handlerPrefix + fun + "?" + searchParams.toString())
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
type state_tx_H_addn struct {
	S_counter int `json:"counter"`
}

func render_tx_H_addn(w io.Writer, key string, states map[string]string, newStates map[string]any, counter int, addNum, addNum_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <p>"))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte("</p> "))

	for i := 0; i < 10; i++ {
		w.Write([]byte("<button tx-key=\"i\" tx-onclick=\""))
		fmt.Fprint(w, addNum)
		w.Write([]byte("?num="))
		if param, err := json.Marshal(i); err != nil {
			log.Panic(err)
		} else {
			w.Write([]byte(url.QueryEscape(string(param))))
		}
		w.Write([]byte("\" tx-swap=\""))
		fmt.Fprint(w, addNum_swap)
		w.Write([]byte("\"> +"))
		w.Write([]byte(html.EscapeString(fmt.Sprint(i))))
		w.Write([]byte(" </button>"))

	}
	w.Write([]byte(" <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_cond struct {
	S_num int `json:"num"`
}

func render_tx_H_cond(w io.Writer, key string, states map[string]string, newStates map[string]any, num int, anon_func_1, anon_func_1_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <button tx-onclick=\"tx_H_cond_anon_func_1\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_1_swap)
	w.Write([]byte("\">change</button> <div> "))
	if num%3 == 0 {
		w.Write([]byte("<p style=\"background: red; color: white\">red</p> "))
	} else if num%3 == 1 {
		w.Write([]byte("<p style=\"background: blue; color: white\">blue</p> "))
	} else {
		w.Write([]byte("<p style=\"background: green; color: white\">green</p> "))

	}
	w.Write([]byte("</div> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_double_H_state struct {
	S_a int `json:"a"`
	S_b int `json:"b"`
}

func render_tx_H_double_H_state(w io.Writer, key string, states map[string]string, newStates map[string]any, a int, b int, init, init_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_double struct {
	S_val int `json:"val"`
}

func render_tx_H_double(w io.Writer, key string, states map[string]string, newStates map[string]any, val int, anon_func_1, anon_func_1_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <p>"))
	w.Write([]byte(html.EscapeString(fmt.Sprint(val))))
	w.Write([]byte("</p> <button tx-onclick=\"tx_H_double_anon_func_1\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_1_swap)
	w.Write([]byte("\">double it!</button> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_example_H_wrapper struct {
}

func render_tx_H_example_H_wrapper(w io.Writer, key string, states map[string]string, newStates map[string]any, render_default_slot func()) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template><div style=\" padding: 2rem; display: flex; justify-content: center; align-items: center; border: solid SlateGray; border-radius: 0.25rem; \"> <div> "))
	if render_default_slot != nil {
		render_default_slot()
	} else {
		w.Write([]byte(" "))

	}
	w.Write([]byte(" </div> </div> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_triangle struct {
	S_counter int `json:"counter"`
}

func render_tx_H_triangle(w io.Writer, key string, states map[string]string, newStates map[string]any, counter int, anon_func_1, anon_func_1_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <div> <span> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte(" </span> <button tx-onclick=\"tx_H_triangle_anon_func_1\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_1_swap)
	w.Write([]byte("\">+</button> </div> "))

	for h := 0; h < counter; h++ {
		w.Write([]byte("<div tx-key=\"h\"> "))

		for s := 0; s < counter-h-1; s++ {
			w.Write([]byte("<span tx-key=\"s\">_</span>"))

		}
		w.Write([]byte(" "))

		for i := 0; i < h*2+1; i++ {
			w.Write([]byte("<span tx-key=\"i\">*</span>"))

		}
		w.Write([]byte(" </div>"))

	}
	w.Write([]byte(" <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_counter struct {
	S_counter int `json:"counter"`
}

func render_tx_H_counter(w io.Writer, key string, states map[string]string, newStates map[string]any, counter int, anon_func_1, anon_func_1_swap string, anon_func_2, anon_func_2_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <button tx-onclick=\"tx_H_counter_anon_func_1\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_1_swap)
	w.Write([]byte("\">-</button> <span> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte(" </span> <button tx-onclick=\"tx_H_counter_anon_func_2\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_2_swap)
	w.Write([]byte("\">+</button> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_current_H_time struct {
	S_t string `json:"t"`
}

func render_tx_H_current_H_time(w io.Writer, key string, states map[string]string, newStates map[string]any, t string, init, init_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <p>"))
	w.Write([]byte(html.EscapeString(fmt.Sprint(t))))
	w.Write([]byte("</p> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_greeting struct {
	S_name string `json:"name"`
}

func render_tx_H_greeting(w io.Writer, key string, states map[string]string, newStates map[string]any, name string, update, update_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <input type=\"text\" tx-value=\"name\" tx-swap=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"value=\""))
	fmt.Fprint(w, name)
	w.Write([]byte("\" placeholder=\"Enter your name\"/> <button tx-onclick=\""))
	fmt.Fprint(w, update)
	w.Write([]byte("\" tx-swap=\""))
	fmt.Fprint(w, update_swap)
	w.Write([]byte("\">Greet</button> "))
	if name != "" {
		w.Write([]byte("<p>Hello, "))
		w.Write([]byte(html.EscapeString(fmt.Sprint(name))))
		w.Write([]byte("</p> "))

	}
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_H_todo struct {
	S_list []string `json:"list"`
	S_item string   `json:"item"`
}

func render_tx_H_todo(w io.Writer, key string, states map[string]string, newStates map[string]any, list []string, item string, add, add_swap string, remove, remove_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <label><input type=\"text\" tx-value=\"item\" tx-swap=\""))
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
		w.Write([]byte("\">"))
		w.Write([]byte(html.EscapeString(fmt.Sprint(l))))
		w.Write([]byte("</li>"))

	}
	w.Write([]byte(" </ol> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state__S_docs struct {
}

func render__S_docs(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>Docs | tmplx</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <nav> <h2>tmplx Docs</h2> <ul> <li><a href=\"#introduction\">Introduction</a></li> <li><a href=\"#installing\">Installing</a></li> <li><a href=\"#quick-start\">Quick Start</a></li> <li><a href=\"#pages-and-routing\">Pages and Routing</a></li> <li><a href=\"#tmplx-script\">tmplx Script</a></li> <li> <a href=\"#expression-interpolation\">Expression Interpolation</a> </li> <li><a href=\"#state\">State</a></li> <li><a href=\"#derived\">Derived</a></li> <li><a href=\"#event-handler\">Event Handler</a></li> <li><a href=\"#init\">init()</a></li> <li><a href=\"#path-parameter\">Path Parameter</a></li> <li> <a href=\"#control-flow\">Control Flow</a> <ul> <li><a href=\"#conditionals\">Conditionals</a></li> <li><a href=\"#loops\">Loops</a></li> </ul> </li> <li><a href=\"#template\">&lt;template&gt;</a></li> <li><a href=\"#input-binding\">🚧 Input Binding</a></li> <li> <a href=\"#component\">Component</a> <ul> <li><a href=\"#props\">Props</a></li> <li><a href=\"#slot\">&lt;slot&gt;</a></li> </ul> </li> <li> Dev Tools <ul> <li><a href=\"#syntax-highlight\">Syntax Highlight</a></li> </ul> </li> </ul> </nav> <main> <h2 id=\"introduction\">Introduction</h2> <p> tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by: </p> <ol> <li> <strong>keeping frontend and backend logic close together</strong> </li> <li> <strong>providing reactive UI updates driven by Go variables</strong> </li> <li><strong>requiring zero new syntax</strong></li> </ol> <p> Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_1"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-todo_1"
					tx_state := &state_tx_H_todo{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_list = []string{}
						tx_state.S_item = ""
					}
					list := tx_state.S_list
					item := tx_state.S_item
					newStates[ckey] = tx_state
					render_tx_H_todo(w, ckey, states, newStates, list, item, "tx_H_todo_add", ckey, "tx_H_todo_remove", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n  var item string = &#34;&#34;\n  \n  func add() {\n    list = append(list, item)\n    item = &#34;&#34;\n  }\n  \n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;label&gt;&lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n&lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n&lt;ol&gt;\n  &lt;li \n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code></pre> <p> You start by creating an HTML file. It can be a page or a reusable component, depending on where you place it. </p> <p> You use the <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;</code> tag to embed Go code and make the page or component dynamic. tmplx uses a subset of Go syntax to provide reactive features like <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, and <a href=\"#event-handler\">event handler</a>. At the same time, because the script is valid Go, you can <strong>implement backend logic</strong>—such as database queries—directly in the template. </p> <p> tmplx compiles the HTML templates and embedded Go code into Go functions that render the HTML on the server and generate HTTP handlers for interactive events. On each interaction, the current state is sent to the server, which computes updates and returns both new HTML and the updated state. The result is server-rendered pages with lightweight client-side swapping (similar to <a href=\"https://htmx.org/\">htmx</a>). The interactivity plumbing is handled automatically by the tmplx compiler and runtime—you just implement the features. </p> <p> Most modern web applications separate the frontend and backend into different languages and teams. tmplx eliminates this split by letting you build the entire interactive application in a single language—Go. With this approach, the mental effort needed to track how data flows from the source to the UI is reduced to a minimum. The fewer transformations you perform on your data, the fewer bugs you introduce. </p> <h2 id=\"installing\">Installing</h2> <p>tmplx requires Go 1.22 or later.</p> <pre><code tx-ignore=\"\">$ go install github.com/gnituy18/tmplx@latest</code></pre> <p> This adds tmplx to your Go bin directory (usually $GOPATH/bin or $HOME/go/bin). Make sure that directory is in your PATH. </p> <p>After installation, verify it works:</p> <pre><code tx-ignore=\"\">$ tmplx --help</code></pre> <h2 id=\"quick-start\">Quick Start</h2> <p>Get a tmplx app running in minutes.</p> <ol> <li> <p><strong>Create a project</strong></p> <pre><code tx-ignore=\"\">$ mkdir hello-tmplx\n$ cd hello-tmplx\n$ go mod init hello-tmplx\n$ mkdir pages</code></pre> </li> <li> <p><strong>Add your first page (pages/index.html)</strong></p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n&lt;head&gt;\n  &lt;meta charset=&#34;UTF-8&#34;&gt;\n  &lt;title&gt;Hello tmplx&lt;/title&gt;\n&lt;/head&gt;\n&lt;body&gt;\n  &lt;script type=&#34;text/tmplx&#34;&gt;\n    var count int\n  &lt;/script&gt;\n\n  &lt;h1&gt;Counter&lt;/h1&gt;\n\n  &lt;button tx-onclick=&#34;count--&#34;&gt;-&lt;/button&gt;\n  &lt;span&gt;{ count }&lt;/span&gt;\n  &lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;\n&lt;/body&gt;\n&lt;/html&gt;</code></pre> </li> <li> <p><strong>Generate the Go code</strong></p> <pre><code tx-ignore=\"\">$ tmplx -out-file tmplx/routes.go</code></pre> </li> <li> <p><strong>Create main.go to serve the app</strong></p> <pre><code tx-ignore=\"\">package main\n\nimport (\n\t&#34;log&#34;\n\t&#34;net/http&#34;\n\n\t&#34;hello-tmplx/tmplx&#34;\n)\n\nfunc main() {\n\tfor _, route := range tmplx.Routes() {\n\t\thttp.Handle(route.Pattern, route.Handler)\n\t}\n\n\tlog.Fatal(http.ListenAndServe(&#34;:8080&#34;, nil))\n}</code></pre> </li> <li> <p><strong>Run the server</strong></p> <pre><code tx-ignore=\"\">$ go run .\n&gt; Listening on :8080</code></pre> </li> </ol> <p> That&#39;s it! Open <a href=\"http://localhost:8080\">http://localhost:8080</a> and you now have a working interactive counter. </p> <h2 id=\"pages-and-routing\">Pages and Routing</h2> <p> A <strong>page</strong> is a standalone HTML file that has its own URL in your web app. </p> <p> All pages are placed in the <strong>pages</strong> directory. Default pages location is <code>./pages</code>. Change it with the <code>-pages</code> flag: </p> <pre><code tx-ignore=\"\">$ tmplx -pages=&#34;/some/other/location&#34;</code></pre> <p> tmplx uses <strong>filesystem-based routing</strong>. The route for a page is the relative path of the HTML file inside the <strong>pages</strong> directory, without the <code>.html</code> extension. For example: </p> <ul> <li><code>pages/index.html</code> → <code>/</code></li> <li><code>pages/about.html</code> → <code>/about</code></li> <li> <code>pages/admin/dashboard.html</code> → <code>/admin/dashboard</code> </li> </ul> <p> When the file is named <code>index.html</code>, the <code>index</code> part is omitted from the route (it serves the directory path). To get a route like <code>/index</code>, place <code>index.html</code> in a subdirectory named <code>index</code>. </p> <ul> <li><code>pages/index/index.html</code> → <code>/index</code></li> </ul> <p> Multiple file paths can map to the same route. Choose the style you prefer. Duplicate routes cause compilation failure. </p> <ul> <li><code>pages/login/index.html</code> → <code>/login</code></li> <li><code>pages/login.html</code> → <code>/login</code></li> </ul> <p> To add URL parameters (path wildcards), use curly braces  in directory or file names inside the pages directory. The name inside  must be a valid Go identifier. </p> <ul> <li> <code tx-ignore=\"\">pages/user/{user_id}.html</code> → <code tx-ignore=\"\">/user/{user_id}</code> </li> <li> <code tx-ignore=\"\">pages/blog/{year}/{slug}.html</code> → <code tx-ignore=\"\">/blog/{year}/{slug}</code> </li> </ul> <p> These patterns are compatible with Go&#39;s <code tx-ignore=\"\">net/http.ServeMux</code> (Go 1.22+). The parameter values are available in page initialisation through <code><a href=\"#path-parameter\">tx:path</a></code> comments. </p> <p> tmplx compiles all pages into a single Go file you can import into your Go project. The pages directory can be outside your project, but keeping it inside is recommended. </p> <h2 id=\"tmplx-script\">tmplx Script</h2> <p> <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> is a special tag that you can add to your page or component to declare <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, <a href=\"#event-handler\">event handler</a>, and the special <a href=\"#init\">init()</a> function to control your UI or add backend logic. </p> <p> Each page or component file can have exactly <strong>one</strong> tmplx script. Multiple scripts cause a compilation error. </p> <p> In pages, place it anywhere inside <code>&lt;head&gt;</code> or <code>&lt;body&gt;</code>. </p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n  &lt;head&gt;\n    ...\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // Go code here\n    &lt;/script&gt;\n    ...\n  &lt;/head&gt;\n  &lt;body&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code> </pre> <p>In components, place it at the root level.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  // Go code here\n&lt;/script&gt;\n...\n...</code></pre> <h2 id=\"expression-interpolation\">Expression Interpolation</h2> <p> Use curly braces <code tx-ignore=\"\">{}</code> to insert <a href=\"https://go.dev/ref/spec#Expressions\">Go expressions</a> into HTML. Expressions are allowed only in: </p> <ul> <li><strong>text nodes</strong></li> <li><strong>attribute values</strong></li> </ul> <p>Placing expressions anywhere else causes a parsing error.</p> <p tx-ignore=\"\">\n        tmplx converts expression results to strings using\n        <code><a href=\"https://pkg.go.dev/fmt#Sprint\">fmt.Sprint</a></code>. The difference is that in <strong>text nodes</strong> the output is\n        <strong>HTML-escaped</strong> to prevent cross-site scripting (XSS)\n        attacks.\n      </p> <p> Expressions run on the server every time the page loads or a component re-renders after an event. Avoid side effects in expressions, such as database queries or heavy computations, because they execute on every render. </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p class=&#39;{ strings.Join([]string{&#34;c1&#34;, &#34;c2&#34;}, &#34; &#34;) }&#39;&gt;\n Hello, { user.GetNameById(0) }!\n&lt;/p&gt;</code> </pre> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p class=&#34;c1 c2&#34;&gt;\n Hello, tmplx!\n&lt;/p&gt;</code></pre> <p tx-ignore=\"\">\n        Add the <code>tx-ignore</code> attribute to an element to disable\n        expression interpolation in that element&#39;s attributes and its direct\n        text children. Descendant elements are still processed normally.\n      </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;{ &#34;not&#34; + &#34; ignored&#34; }&lt;/span&gt;\n&lt;/p&gt;</code> </pre> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;not ignored&lt;/span&gt;\n&lt;/p&gt;</code></pre> <h2 id=\"state\">State</h2> <p> <strong>State</strong> is the mutable data that describes a component&#39;s current condition. </p> <p> Declaring state works like declaring variables in Go&#39;s package scope. If you provide no initial value, the state starts with the zero value for its type. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string\n&lt;/script&gt;</code></pre> <p>To set an initial value, use the <code>=</code> operator.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string = &#34;tmplx&#34;\n&lt;/script&gt;</code></pre> <p>Although the syntax follows valid Go code, these rules apply:</p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong>The type must be explicitly declared and JSON-compatible.</strong> </li> </ol> <p> The 1st rule is enforced by the compiler. The 2nd is not checked at compile time (for now) and will cause a runtime error if violated. </p> <h3>Some invalid state declarations:</h3> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n// ❌ Must explicitly declare the type\nvar str = &#34;&#34;\n\n// ❌ Cannot use the := short declaration\nnum := 1\n\n// ❌ Type must be JSON-marshalable/unmarshalable\nvar f func(int) = func(i int) { ... }\nvar w io.Writer\n\n// ❌ Only one identifier per declaration\nvar a, b int = 10, 20\nvar a, b int = f()\n&lt;/script&gt;</code></pre> <h3>Some valid state declarations:</h3> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n// ✅ Zero value\nvar id int64\n\n// ✅ With initial value\nvar address string = &#34;...&#34;\n\n// ✅ Initialized with a function call (assuming the package is imported)\nvar username string = user.GetNameById(&#34;id&#34;)\n\n// ✅ Complex JSON-compatible types\nvar m map[string]int = map[string]int{&#34;key&#34;: 100}\n&lt;/script&gt;</code></pre> <h2 id=\"derived\">Derived</h2> A <strong>derived</strong> is a <strong>read-only</strong> value that is automatically calculated from states. It updates whenever those states change. <p> Declaring a derived works the same way as declaring package-level variables in Go. When the right-hand side of the declaration <strong>references existing state or other derived values</strong>, it is treated as a derived value. </p> <p> Derived values follow most of the same rules as regular state variables, but with some differences: </p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li><strong>The type must be specified explicitly.</strong></li> <li> <strong>Derived values cannot be modified directly in event handlers, though they may be read.</strong> </li> </ol> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num1 int = 100 // state\n  var num2 int = num1 * 2 // derived\n&lt;/script&gt;\n\n...\n&lt;p&gt;{num1} * 2 = {num2}&lt;/p&gt;</code></pre> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var classStrs []string = []string{&#34;c1&#34;, &#34;c2&#34;, &#34;c3&#34;} // state\n  var class string = strings.Join(classStrs, &#34; &#34;) // derived\n&lt;/script&gt;\n\n...\n&lt;p class=&#34;{class}&#34;&gt; ... &lt;/p&gt;</code></pre> <h2 id=\"event-handler\">Event Handler</h2> <p> Event handlers let you respond to frontend events with backend logic or update state to trigger UI changes. </p> <p> To declare an event handler, define a Go function in the global scope of the <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> block. Bind it to a DOM event by adding an attribute that starts with <code>tx-on</code> followed by the event name (e.g., <code>tx-onclick</code>). </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func add1() {\n    counter += 1\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-onclick=&#34;add1()&#34;&gt;Add 1&lt;/button&gt;</code></pre> <p> In this example, the <code>add1</code> handler runs every time the button is clicked. The <code>counter</code> state increases by 1, and the paragraph updates automatically. </p> <p> It’s not magic. tmplx compiles each event handler into an HTTP endpoint. The runtime JavaScript attaches a lightweight listener that sends the required state to the endpoint, receives the updated HTML fragment, merges the new state, and swaps the affected part of the DOM. It feels like direct backend access from the client, but it’s just a simple API call with targeted DOM swapping. </p> <h3>Arguments</h3> You can add arguments from local variable declared within <code>tx-if</code>, or <code>tx-for</code> with the following rules: <ul> <li> <strong>Argument names cannot match state or derived state names. </strong> </li> <li><strong>Argument types must be JSON-compatible.</strong></li> </ul> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_2"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-addn_1"
					tx_state := &state_tx_H_addn{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_counter = 0
					}
					counter := tx_state.S_counter
					newStates[ckey] = tx_state
					render_tx_H_addn(w, ckey, states, newStates, counter, "tx_H_addn_addNum", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func addNum(num int) {\n    counter += num\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-for=&#34;i := 0; i &lt; 10; i++&#34; tx-key=&#34;i&#34; tx-onclick=&#34;addNum(i)&#34;&gt;\n  +{ i }\n&lt;/button&gt;</code></pre> <h3>Inline Statements</h3> <p> For simple actions, embed Go statements directly in <code>tx-on*</code> attributes to update state. This avoids defining separate handler functions. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_3"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-double_1"
					tx_state := &state_tx_H_double{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_val = 1
					}
					val := tx_state.S_val
					newStates[ckey] = tx_state
					render_tx_H_double(w, ckey, states, newStates, val, "tx_H_double_anon_func_1", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var val int = 1\n&lt;/script&gt;\n\n&lt;p&gt;{ val }&lt;/p&gt;\n&lt;button tx-onclick=&#34;val *= 2&#34;&gt;double it!&lt;/button&gt;</code> </pre> <h2 id=\"init\">init()</h2> <p> <code>init()</code> is a special function that runs automatically <strong>once</strong> when a <strong>page</strong> is first rendered. Components cannot use this function. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_4"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-current-time_1"
					tx_state := &state_tx_H_current_H_time{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					t := tx_state.S_t
					if !tx_old_state_exist {
						t = fmt.Sprint(time.Now().Format(time.RFC3339))
						tx_state.S_t = t
					}
					newStates[ckey] = tx_state
					render_tx_H_current_H_time(w, ckey, states, newStates, t, "tx_H_current_H_time_init", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var t string\n\n  func init() {\n    t = fmt.Sprint(time.Now().Format(time.RFC3339))\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ t }&lt;/p&gt;</code></pre> <p> Another common use case is to initialize one state from another state without turning the second variable into a derived state. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var a int = 1\n  var b int\n\n  func init() {\n    b = a * 2 // b remains a regular state\n  }\n&lt;/script&gt;</code></pre> <h2 id=\"path-parameter\">Path Parameters</h2> <p> You can inject path parameters into states using a <code>//tx:path</code> comment placed directly above the state declaration. This feature works only in <a href=\"#pages-and-routing\">pages</a> and requires the state to be of type <code>string</code>. </p> <p> For example, given a route pattern like <code tx-ignore=\"\">/blog/post/{post_id}</code>, you can access the <code>post_id</code> parameter as follows: </p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html&gt;\n  &lt;head&gt;\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // tx:path post_id\n      var postId string\n\n      var post Post\n      \n      func init() {\n        post = db.GetPost(postId)\n      }\n    &lt;/script&gt;\n  &lt;/head&gt;\n\n  &lt;body&gt;\n    &lt;h1&gt;{ post.Title }&lt;/h1&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code></pre> <p> The value of the <code>post_id</code> path parameter is automatically injected into the <code>postId</code> state during initialization. After that, <code>postId</code> behaves like any other state and can be read or modified as needed. </p> <h2 id=\"control-flow\">Control Flow</h2> <p> tmplx avoids new custom syntax for conditionals and loops because that would increase compiler complexity. Instead, it embeds control flow directly into HTML attributes, similar to Vue.js and <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h3 id=\"conditionals\">Conditionals</h3> <p> To conditionally render elements, use the <code>tx-if</code>, <code>tx-else-if</code>, and <code>tx-else</code> attributes on the desired tags. The values for <code>tx-if</code> and <code>tx-else-if</code> can be any valid Go expression that would fit in an <code>if</code> or <code>else if</code> statement. The <code>tx-else</code> attribute needs no value. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_5"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-cond_1"
					tx_state := &state_tx_H_cond{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					num := tx_state.S_num
					newStates[ckey] = tx_state
					render_tx_H_cond(w, ckey, states, newStates, num, "tx_H_cond_anon_func_1", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;num++&#34;&gt;change&lt;/button&gt;\n&lt;div&gt;\n  &lt;p tx-if=&#34;num % 3 == 0&#34; style=&#34;background: red; color: white&#34;&gt;red&lt;/p&gt;\n  &lt;p tx-else-if=&#34;num % 3 == 1&#34; style=&#34;background: blue; color: white&#34;&gt;blue&lt;/p&gt;\n  &lt;p tx-else style=&#34;background: green; color: white&#34;&gt;green&lt;/p&gt;\n&lt;/div&gt;</code> </pre> <p> You can declare <strong>local variables</strong> and handle errors exactly as you would in regular Go code. Local variables declared in conditionals are available to the element and its descendants, just like in Go. </p> <pre><code tx-ignore=\"\">&lt;p tx-if=&#34;user, err := user.GetUser(); err != nil&#34;&gt;\n  &lt;span tx-if=&#34;err == ErrNotFound&#34;&gt;User not found&lt;/span&gt;\n&lt;/p&gt;\n&lt;p tx-else-if=&#39;user.Name == &#34;&#34;&#39;&gt;user.Name not set&lt;/p&gt;\n&lt;p tx-else&gt;Hi, { user.Name }&lt;/p&gt;</code></pre> <p> A conditional group consists of <strong>consecutive sibling nodes</strong> that share the same parent. Disconnected nodes are not treated as part of the same group. A standalone <code>tx-else-if</code> or <code>tx-else</code> without a preceding <code>tx-if</code> will cause a compilation error. </p> <h3 id=\"loops\">Loops</h3> <p> To repeat elements, use the <code>tx-for</code> attribute. Its value can be any valid Go <code>for</code> statement, including <strong>classic for</strong> or <strong>range for</strong>. </p> <p> Local variables declared in the loop are available to the element and all of its descendants, just like in Go. </p> <p> Always add a <code>tx-key</code> attribute with a unique value for each item. This gives the compiler a unique identifier for the node during updates. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_6"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-triangle_1"
					tx_state := &state_tx_H_triangle{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_counter = 5
					}
					counter := tx_state.S_counter
					newStates[ckey] = tx_state
					render_tx_H_triangle(w, ckey, states, newStates, counter, "tx_H_triangle_anon_func_1", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code> </pre> <pre><code tx-ignore=\"\">&lt;div tx-for=&#34;_, user := range users&#34;&gt;\n  { user.Id }: { user.Name }\n&lt;/div&gt;</code></pre> <h2 id=\"template\">&lt;template&gt;</h2> <p> The <code>&lt;template&gt;</code> tag is a non-rendering container that lets you apply control flow attributes (<code>tx-if</code>, <code>tx-else-if</code>, <code>tx-else</code>, or <code>tx-for</code>) to a group of elements at once. </p> <p> The <code>&lt;template&gt;</code> itself is removed from the outpu.t only its children are rendered (or not, depending on the control flow). </p> <p> You can nest <code>&lt;template&gt;</code> tags and combine them with other control flow attributes on child elements. </p> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var loggedIn bool = true\n&lt;/script&gt;\n\n&lt;template tx-if=&#34;loggedIn&#34;&gt;\n  &lt;p&gt;Welcome back!&lt;/p&gt;\n  &lt;button tx-onclick=&#34;logout()&#34;&gt;Logout&lt;/button&gt;\n&lt;/template&gt;\n\n&lt;template tx-else&gt;\n  &lt;p&gt;Please sign in.&lt;/p&gt;\n  &lt;button tx-onclick=&#34;login()&#34;&gt;Login&lt;/button&gt;\n&lt;/template&gt;</code> </pre> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var posts []Post = []Post{\n    {Title: &#34;First Post&#34;, Body: &#34;Hello world&#34;},\n    {Title: &#34;Second Post&#34;, Body: &#34;tmplx is great&#34;},\n  }\n&lt;/script&gt;\n\n&lt;template tx-for=&#34;i, p := range posts&#34; tx-key=&#34;i&#34;&gt;\n  &lt;article&gt;\n    &lt;h3&gt;{ p.Title }&lt;/h3&gt;\n    &lt;p&gt;{ p.Body }&lt;/p&gt;\n    &lt;hr&gt;\n  &lt;/article&gt;\n&lt;/template&gt;</code> </pre> <h2 id=\"input-binding\">🚧 Input Binding</h2> <p> <strong>This feature is in active development and may change or be removed in the future.</strong> </p> <p> tmplx provides two-way binding (kind of) for <code>&lt;input&gt;</code> elements using the <code>tx-value</code> attribute. The input value is kept in sync with a state variable on the client side. </p> <p> As the user types, the client-side state updates immediately. The displayed value only refreshes after a server round-trip that causes a re-render. Conversely, when another event handler modifies the state, the input value updates automatically on the next render. </p> "))
	{
		ckey := key + "_/docs_tx-example-wrapper_7"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/docs_tx-greeting_1"
					tx_state := &state_tx_H_greeting{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_name = ""
					}
					name := tx_state.S_name
					newStates[ckey] = tx_state
					render_tx_H_greeting(w, ckey, states, newStates, name, "tx_H_greeting_update", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var name string = &#34;&#34;\n\n  func update() {}\n&lt;/script&gt;\n\n&lt;input type=&#34;text&#34; tx-value=&#34;name&#34; placeholder=&#34;Enter your name&#34; /&gt;\n&lt;button tx-onclick=&#34;update()&#34;&gt;Greet&lt;/button&gt;\n\n&lt;p tx-if=&#39;name != &#34;&#34;&#39;&gt;Hello, { name }&lt;/p&gt;</code></pre> <p> Typing updates the client state instantly, but the greeting only changes after clicking the button, which triggers a server round-trip and re-render. </p> <p> This is a current limitation of server-side re-renders. For smooth live preview, combine tmplx with a client-side library like <a href=\"https://alpinejs.dev/\">Alpine.js</a> to handle instant updates locally. </p> <h2 id=\"component\">Component</h2> <p> Components are reusable UI building blocks that encapsulate HTML, state, and behavior. </p> <p> Create a component by placing an <code>.html</code> file in the <code>components</code> directory (default: <code>./components</code>). tmplx automatically registers it as a custom element with the tag name <code>tx-</code> followed by the lowercase kebab-case version of the relative path (without the <code>.html</code> extension). </p> <p>Examples:</p> <ul> <li> <code>components/Button.html</code> → <code>&lt;tx-button&gt;</code> </li> <li> <code>components/user/Card.html</code> → <code>&lt;tx-user-card&gt;</code> </li> <li> <code>components/todo/List.html</code> → <code>&lt;tx-todo-list&gt;</code> </li> </ul> <p> Components can contain their own <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> for local state and logic, and can be used in pages or nested inside other components. </p> <h3 id=\"props\">Props</h3> <p> Pass data to a component via attributes. Inside the component, declare matching state variables with <code>// tx:prop</code> comments to make them reactive to prop changes. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  // tx:prop\n  var title string\n\n  // tx:prop\n  var count int = 0\n&lt;/script&gt;\n\n&lt;h3&gt;{ title }&lt;/h3&gt;\n&lt;span&gt;{ count }&lt;/span&gt;\n&lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;</code></pre> <p>Usage:</p> <pre><code tx-ignore=\"\">&lt;tx-my-component title=&#34;Hello&#34; count=&#34;5&#34;&gt;&lt;/tx-my-component&gt;</code></pre> <h3 id=\"slot\">&lt;slot&gt;</h3> <p> Use <code>&lt;slot&gt;</code> to define insertion points for child content. Name slots for multiple insertion points. </p> <pre><code tx-ignore=\"\">&lt;template&gt;\n  &lt;div class=&#34;card&#34;&gt;\n    &lt;slot name=&#34;header&#34;&gt;Default Header&lt;/slot&gt;\n    &lt;div class=&#34;body&#34;&gt;\n      &lt;slot&gt;Default Body&lt;/slot&gt;\n    &lt;/div&gt;\n    &lt;slot name=&#34;footer&#34;&gt;&lt;/slot&gt;\n  &lt;/div&gt;\n&lt;/template&gt;</code></pre> <p>Usage:</p> <pre><code tx-ignore=\"\">&lt;tx-card&gt;\n  &lt;h2 slot=&#34;header&#34;&gt;Custom Title&lt;/h2&gt;\n  &lt;p&gt;Custom content&lt;/p&gt;\n  &lt;div slot=&#34;footer&#34;&gt;Actions&lt;/div&gt;\n&lt;/tx-card&gt;</code></pre> <p> Unnamed slots receive content without a <code>slot</code> attribute. Fallback content inside <code>&lt;slot&gt;</code> renders when no matching content is provided. </p> <h3 id=\"props\">Props</h3> <p>Pass data to components via attributes. These become initial values for matching state variables.</p> <p>tmplx matches attribute names (kebab-case) to state variable names (camelCase conversion). Only top-level <code>var</code> declarations without initializers can receive props.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var title string     // receives &#34;title&#34; attribute\n  var count int        // receives &#34;count&#34; attribute, defaults to 0\n&lt;/script&gt;\n\n&lt;h3&gt;{ title }&lt;/h3&gt;\n&lt;span&gt;{ count }&lt;/span&gt;\n&lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;</code></pre> <p>Usage:</p> <pre><code tx-ignore=\"\">&lt;tx-counter title=&#34;Clicks&#34; count=&#34;5&#34;&gt;&lt;/tx-counter&gt;</code></pre> <p>Current best way (no special syntax): declare receivable vars without initializer. Local state gets default or explicit initializer. Props override only on initial render.</p> <p>Alternative ideas using pure Go:</p> <ul> <li><code tx-ignore=\"\">var Props struct { Title string; Count int }</code> — compiler could auto-populate matching fields.</li> <li>Reserve prefix like <code>var PropTitle string</code> — convention-based.</li> <li>Comment marker <code>// tx:prop</code> — explicit, zero runtime cost (current draft approach).</li> </ul> <p>Comment marker is cleanest: distinguishes external props from internal state without name conflicts or conventions.</p> <h2 id=\"syntax-highlight\">Syntax Highlight</h2> <a href=\"https://github.com/gnituy18/tmplx.nvim\">Neovim Plugin</a> </main> </body></html>"))
}

type state__S__EX_ struct {
}

func render__S__EX_(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <main> <h1 style=\"text-align: center\">&lt;tmplx&gt;</h1> <h2 style=\"text-align: center; margin-top: 1.5rem\"> Write Go in HTML intuitively </h2> <ul style=\"margin-top: 4rem\"> <li>Full Go backend logic and HTML in the same file</li> <li>Reactive UIs driven by plain Go variables</li> <li>Reusable components written as regular HTML files</li> </ul> <div style=\" display: flex; gap: 2rem; justify-content: center; text-align: center; margin-top: 4rem; \"> <a class=\"btn\" href=\"/docs\">Docs</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> <h2 style=\"text-align: center\">Demos</h2> <h3>Counter</h3> "))
	{
		ckey := key + "_/{$}_tx-example-wrapper_1"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/{$}_tx-counter_1"
					tx_state := &state_tx_H_counter{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					counter := tx_state.S_counter
					newStates[ckey] = tx_state
					render_tx_H_counter(w, ckey, states, newStates, counter, "tx_H_counter_anon_func_1", ckey, "tx_H_counter_anon_func_2", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;counter--&#34;&gt;-&lt;/button&gt;\n&lt;span&gt; { counter } &lt;/span&gt;\n&lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;</code> </pre> <h3>To Do</h3> "))
	{
		ckey := key + "_/{$}_tx-example-wrapper_2"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/{$}_tx-todo_1"
					tx_state := &state_tx_H_todo{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_list = []string{}
						tx_state.S_item = ""
					}
					list := tx_state.S_list
					item := tx_state.S_item
					newStates[ckey] = tx_state
					render_tx_H_todo(w, ckey, states, newStates, list, item, "tx_H_todo_add", ckey, "tx_H_todo_remove", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string = []string{}\n  var item string = &#34;&#34;\n  \n  func add() {\n    list = append(list, item)\n    item = &#34;&#34;\n  }\n  \n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;label&gt;&lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n&lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n&lt;ol&gt;\n  &lt;li \n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code> </pre> <h3>Triangle</h3> "))
	{
		ckey := key + "_/{$}_tx-example-wrapper_3"
		tx_state := &state_tx_H_example_H_wrapper{}
		tx_old_state, tx_old_state_exist := states[ckey]
		if tx_old_state_exist {
			json.Unmarshal([]byte(tx_old_state), tx_state)
		}
		newStates[ckey] = tx_state
		render_tx_H_example_H_wrapper(w, ckey, states, newStates,
			func() {
				w.Write([]byte(" "))
				{
					ckey := key + "_/{$}_tx-triangle_1"
					tx_state := &state_tx_H_triangle{}
					tx_old_state, tx_old_state_exist := states[ckey]
					if tx_old_state_exist {
						json.Unmarshal([]byte(tx_old_state), tx_state)
					}
					if !tx_old_state_exist {
						tx_state.S_counter = 5
					}
					counter := tx_state.S_counter
					newStates[ckey] = tx_state
					render_tx_H_triangle(w, ckey, states, newStates, counter, "tx_H_triangle_anon_func_1", ckey)
				}
				w.Write([]byte(" "))

			},
		)
	}
	w.Write([]byte(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code> </pre> </main> </body></html>"))
}

var txRoutes []TxRoute = []TxRoute{
	{
		Pattern: "GET /docs",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			state := &state__S_docs{}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render__S_docs(&buf, "tx_", map[string]string{}, newStates)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /{$}",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			state := &state__S__EX_{}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render__S__EX_(&buf, "tx_", map[string]string{}, newStates)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /tx/tx_H_greeting_update",
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
			state := &state_tx_H_greeting{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			name := state.S_name
			render_tx_H_greeting(w, txSwap, states, newStates, name, "tx_H_greeting_update", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_greeting{
				S_name: name,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "GET /tx/tx_H_todo_add",
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
			state := &state_tx_H_todo{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			list := state.S_list
			item := state.S_item
			list = append(list, item)
			item = ""
			render_tx_H_todo(w, txSwap, states, newStates, list, item, "tx_H_todo_add", txSwap, "tx_H_todo_remove", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_todo{
				S_list: list,
				S_item: item,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "GET /tx/tx_H_todo_remove",
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
			state := &state_tx_H_todo{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			list := state.S_list
			item := state.S_item
			var i int
			json.Unmarshal([]byte(query.Get("i")), &i)
			list = append(list[0:i], list[i+1:]...)
			render_tx_H_todo(w, txSwap, states, newStates, list, item, "tx_H_todo_add", txSwap, "tx_H_todo_remove", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_todo{
				S_list: list,
				S_item: item,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "GET /tx/tx_H_addn_addNum",
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
			state := &state_tx_H_addn{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			var num int
			json.Unmarshal([]byte(query.Get("num")), &num)
			counter += num
			render_tx_H_addn(w, txSwap, states, newStates, counter, "tx_H_addn_addNum", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_addn{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_H_cond_anon_func_1",
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
			state := &state_tx_H_cond{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			num := state.S_num
			num++
			render_tx_H_cond(w, txSwap, states, newStates, num, "tx_H_cond_anon_func_1", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_cond{
				S_num: num,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_H_double_anon_func_1",
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
			state := &state_tx_H_double{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			val := state.S_val
			val *= 2
			render_tx_H_double(w, txSwap, states, newStates, val, "tx_H_double_anon_func_1", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_double{
				S_val: val,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_H_triangle_anon_func_1",
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
			state := &state_tx_H_triangle{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter++
			render_tx_H_triangle(w, txSwap, states, newStates, counter, "tx_H_triangle_anon_func_1", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_triangle{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_H_counter_anon_func_1",
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
			state := &state_tx_H_counter{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter--
			render_tx_H_counter(w, txSwap, states, newStates, counter, "tx_H_counter_anon_func_1", txSwap, "tx_H_counter_anon_func_2", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_counter{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_H_counter_anon_func_2",
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
			state := &state_tx_H_counter{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter++
			render_tx_H_counter(w, txSwap, states, newStates, counter, "tx_H_counter_anon_func_1", txSwap, "tx_H_counter_anon_func_2", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_H_counter{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
}

func Routes() []TxRoute { return txRoutes }
