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
type state_tx_h_counter struct {
	S_counter int `json:"counter"`
}

func render_tx_h_counter(w io.Writer, key string, states map[string]string, newStates map[string]any, counter int, anon_func_1, anon_func_1_swap string, anon_func_2, anon_func_2_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <button tx-onclick=\"tx_h_counter_anon_func_1\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_1_swap)
	w.Write([]byte("\">-</button> <span> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte(" </span> <button tx-onclick=\"tx_h_counter_anon_func_2\"\" tx-swap=\""))
	fmt.Fprint(w, anon_func_2_swap)
	w.Write([]byte("\">+</button> <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_tx_h_todo struct {
	S_list []string `json:"list"`
	S_item string   `json:"item"`
}

func render_tx_h_todo(w io.Writer, key string, states map[string]string, newStates map[string]any, list []string, item string, add, add_swap string, remove, remove_swap string) {
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

type state_tx_h_triangle struct {
	S_counter int `json:"counter"`
}

func render_tx_h_triangle(w io.Writer, key string, states map[string]string, newStates map[string]any, counter int, anon_func_1, anon_func_1_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template> <div> <span> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte(" </span> <button tx-onclick=\"tx_h_triangle_anon_func_1\"\" tx-swap=\""))
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

type state_docs struct {
}

func render_docs(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>Docs | tmplx</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <nav> <h2>tmplx Docs</h2> <ul> <li><a href=\"#introduction\">Introduction</a></li> <li><a href=\"#installing\">Installing</a></li> <li><a href=\"#pages-and-routing\">Pages and Routing</a></li> <li><a href=\"#tmplx-script\">tmplx Script</a></li> <li> <a href=\"#expression-interpolation\">Expression Interpolation</a> </li> <li><a href=\"#state\">State</a></li> <li><a href=\"#derived-state\">Derived State</a></li> <li><a href=\"#event-handler\">Event Handler</a></li> <li><a href=\"#init\">init()</a></li> <li> <h3>Dev Tools</h3> <ul> <li><a href=\"#syntax-highlight\">Syntax Highlight</a></li> </ul> </li> </ul> </nav> <main> <h2 id=\"introduction\">Introduction</h2> <p> tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by: </p> <ol> <li>keeping frontend and backend logic close together</li> <li>providing reactive UI updates driven by Go variables</li> <li>requiring zero new syntax</li> </ol> <p> Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive. </p> <div style=\" padding: 2rem; display: flex; justify-content: center; align-items: center; border: solid SlateGray; border-radius: 0.25rem; \"> <div> "))
	{
		ckey := key + "_docs_tx-todo_1"
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
	w.Write([]byte(" </div> </div> <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n  var item string = &#34;&#34;\n  \n  func add() {\n    list = append(list, item)\n    item = &#34;&#34;\n  }\n  \n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;label&gt;&lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n&lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n&lt;ol&gt;\n  &lt;li \n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code></pre> <p> You start by creating an HTML file. It can be a page or a reusable component, depending on where you place it. tmplx respects the HTML standard. There&#39;s no new syntax for </p> <p> You use the <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;</code> tag to embed valid Go code and make the page or component dynamic. tmplx uses a subset of Go syntax to provide reactive features like local state, derived values, and event handlers. At the same time, because the script is valid Go, you can implement any backend logic—such as database queries—directly in the template. </p> <p> tmplx compiles the HTML templates and embedded Go code into Go functions that render the HTML on the server and generate HTTP handlers for interactive events. On each interaction, the current state is sent to the server, which computes updates and returns both new HTML and the updated state. The result is server-rendered pages with lightweight client-side swapping (similar to <a href=\"https://htmx.org/\">htmx</a>). The interactivity plumbing is handled automatically by the tmplx compiler and runtime—you just implement the features. </p> <p> Most modern web applications separate the frontend and backend into different languages and teams. tmplx eliminates this split by letting you build the entire interactive application in a single language—Go. With this approach, the mental effort needed to track how data flows from the source to the UI is reduced to a minimum. The fewer transformations you perform on your data, the fewer bugs you introduce. </p> <h2 id=\"pages-and-routing\">Pages and Routing</h2> <p> A <strong>page</strong> is a standalone HTML file that has its own URL in your web app. </p> <p>All pages are placed in the <strong>pages</strong> directory.</p> <p> tmplx uses <strong>filesystem-based routing</strong>. The route for a page is the relative path of the HTML file inside the <strong>pages</strong> directory, without the <code>.html</code> extension. For example: </p> <ul> <li><code>pages/index.html</code> → <code>/</code></li> <li><code>pages/about.html</code> → <code>/about</code></li> <li> <code>pages/admin/dashboard.html</code> → <code>/admin/dashboard</code> </li> </ul> <p> When the file is named <code>index.html</code>, the <code>index</code> part is omitted from the route (it serves the directory path). To get a route like <code>/index</code>, place <code>index.html</code> in a subdirectory named <code>index</code>. </p> <ul> <li><code>pages/index/index.html</code> → <code>/index</code></li> </ul> <p> Multiple file paths can map to the same route. Choose the style you prefer. Duplicate routes cause compilation failure. </p> <ul> <li><code>pages/login/index.html</code> → <code>/login</code></li> <li><code>pages/login.html</code> → <code>/login</code></li> </ul> <p> To add URL parameters (path wildcards), use curly braces  in directory or file names inside the pages directory. The name inside  must be a valid Go identifier. </p> <ul> <li> <code tx-ignore=\"\">pages/user/{user_id}.html</code> → <code tx-ignore=\"\">/user/{user_id}</code> </li> <li> <code tx-ignore=\"\">pages/blog/{year}/{slug}.html</code> → <code tx-ignore=\"\">/blog/{year}/{slug}</code> </li> </ul> <p> These patterns are compatible with Go&#39;s <code tx-ignore=\"\">net/http.ServeMux</code> (Go 1.22+). The parameter values are available in page initialisation through tmplx comments. </p> <p> tmplx compiles all pages into a single Go file you can import into your Go project. The pages directory can be outside your project, but keeping it inside is recommended. </p> <p> Default pages location: <code>./pages</code>. Change it with the <code>-pages</code> flag: </p> <pre><code tx-ignore=\"\">$ tmplx -pages=&#34;/some/other/location&#34;</code></pre> <h2 id=\"tmplx-script\">tmplx Script</h2> <p> <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> is a special tag that you can add to your page or component to declare <a href=\"#state\">state</a>, <a href=\"#derived-state\">derived state</a>, <a href=\"#event-handlers\">event handlers</a>, and the special <a href=\"#init\">init()</a> function to control your UI or add backend logic. </p> <p> Each page or component file can have exactly <strong>one</strong> tmplx script. Multiple scripts cause a compilation error. </p> <p> In pages, place it anywhere inside <code>&lt;head&gt;</code> or <code>&lt;body&gt;</code>. </p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n  &lt;head&gt;\n    ...\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // Go code here\n    &lt;/script&gt;\n    ...\n  &lt;/head&gt;\n  &lt;body&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code> </pre> <p>In components, place it at the root level.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  // Go code here\n&lt;/script&gt;\n...\n...</code></pre> <h2 id=\"expression-interpolation\">Expression Interpolation</h2> <p tx-ignore=\"\">\n        Embed Go expressions in HTML using {} for dynamic content. You can only\n        place Go expressions in text nodes or attribute values; other placements\n        cause parsing errors. text nodes output is HTML-escaped; attribute it is\n        not escaped.\n      </p> <p> Every expression is run once when the page is load or every component is rerender, so bemindful with every code you put in the interpolation like accessing database or run heavying loading functions. </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p class=&#39;{ strings.Join([]string{&#34;c1&#34;, &#34;c2&#34;}, &#34; &#34;) }&#39;&gt;\n Hello, { user.GetNameById(&#34;id&#34;) }!\n&lt;/p&gt;</code></pre> <pre> <code>&lt;p class=&#34;c1 c2&#34;&gt; Hello, tmplx! &lt;/p&gt;</code> </pre> <p> You can add tx-ignore to disable Go expression interpolation for that specific node&#39;s attribute values and its text children, but not the element children. </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt; { &#34;not&#34; + &#34;ignore&#34; } &lt;/span&gt;\n&lt;/p&gt;</code> </pre> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt; not ignored &lt;/span&gt;\n&lt;/p&gt;</code></pre> <h2>Syntax Highlight</h2> <a href=\"https://github.com/gnituy18/tmplx.nvim\">Neovim Plugin</a> </main> </body></html>"))
}

type state_index struct {
}

func render_index(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<!DOCTYPE html><html lang=\"en\"><head> <title>tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <main> <h1 style=\"text-align: center\">&lt;tmplx&gt;</h1> <h2 style=\"text-align: center; margin-top: 1.5rem\"> Interactive Web App with just HTML and Go </h2> <ul style=\"margin-top: 4rem\"> <li>Reactive UIs driven by plain Go variables</li> <li>Reusable components written as regular HTML files</li> <li>Full Go backend logic and HTML in the same file</li> </ul> <div style=\" display: flex; gap: 2rem; justify-content: center; text-align: center; margin-top: 4rem; \"> <a class=\"btn\" href=\"/docs\">Docs</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> <h2 style=\"text-align: center\">Demos</h2> <h3>Counter</h3> <div style=\" padding: 2rem; display: flex; justify-content: center; align-items: center; border: solid SlateGray; border-radius: 0.25rem; \"> <div> "))
	{
		ckey := key + "_index_tx-counter_1"
		state := &state_tx_h_counter{}
		if _, ok := states[ckey]; ok {
			json.Unmarshal([]byte(states[ckey]), state)
			newStates[ckey] = state
		} else {

			newStates[ckey] = state
		}
		counter := state.S_counter
		render_tx_h_counter(w, ckey, states, newStates, counter, "tx_h_counter_anon_func_1", ckey, "tx_h_counter_anon_func_2", ckey)
	}
	w.Write([]byte(" </div> </div> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;counter--&#34;&gt;-&lt;/button&gt;\n&lt;span&gt; { counter } &lt;/span&gt;\n&lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;</code> </pre> <h3>To Do</h3> <div style=\" padding: 2rem; display: flex; justify-content: center; align-items: center; border: solid SlateGray; border-radius: 0.25rem; \"> <div> "))
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
	w.Write([]byte(" </div> </div> <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string = []string{}\n  var item string = &#34;&#34;\n  \n  func add() {\n    list = append(list, item)\n    item = &#34;&#34;\n  }\n  \n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;label&gt;&lt;input type=&#34;text&#34; tx-value=&#34;item&#34;&gt;&lt;/label&gt;\n&lt;button tx-onclick=&#34;add()&#34;&gt;Add&lt;/button&gt;\n&lt;ol&gt;\n  &lt;li \n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code> </pre> <h3>Triangle</h3> <div style=\" padding: 2rem; display: flex; justify-content: center; align-items: center; border: solid SlateGray; border-radius: 0.25rem; \"> <div> "))
	{
		ckey := key + "_index_tx-triangle_1"
		state := &state_tx_h_triangle{}
		if _, ok := states[ckey]; ok {
			json.Unmarshal([]byte(states[ckey]), state)
			newStates[ckey] = state
		} else {
			state.S_counter = 5
			newStates[ckey] = state
		}
		counter := state.S_counter
		render_tx_h_triangle(w, ckey, states, newStates, counter, "tx_h_triangle_anon_func_1", ckey)
	}
	w.Write([]byte(" </div> </div> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code> </pre> </main> </body></html>"))
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
		Pattern: " /tx/tx_h_counter_anon_func_1",
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
			state := &state_tx_h_counter{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter--
			render_tx_h_counter(w, txSwap, states, newStates, counter, "tx_h_counter_anon_func_1", txSwap, "tx_h_counter_anon_func_2", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_counter{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: " /tx/tx_h_counter_anon_func_2",
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
			state := &state_tx_h_counter{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter++
			render_tx_h_counter(w, txSwap, states, newStates, counter, "tx_h_counter_anon_func_1", txSwap, "tx_h_counter_anon_func_2", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_counter{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
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
	{
		Pattern: " /tx/tx_h_triangle_anon_func_1",
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
			state := &state_tx_h_triangle{}
			json.Unmarshal([]byte(states[txSwap]), &state)
			counter := state.S_counter
			counter++
			render_tx_h_triangle(w, txSwap, states, newStates, counter, "tx_h_triangle_anon_func_1", txSwap)
			w.Write([]byte("<script id=\"tx-state\" type=\"application/json\">"))
			newStates[txSwap] = &state_tx_h_triangle{
				S_counter: counter,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(stateBytes)
			w.Write([]byte("</script>"))
		},
	},
}

func Routes() []TxRoute { return txRoutes }
