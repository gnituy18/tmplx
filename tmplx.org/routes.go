package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"time"
)

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
  let state = JSON.parse(this.getElementById("tx-saved").innerHTML)
  let tasks = [];
  let isProcessing = false;

  const findComment = (text) => {
    const walker = document.createTreeWalker(document.documentElement, NodeFilter.SHOW_COMMENT)
    while (walker.nextNode()) {
      if (walker.currentNode.nodeValue === text) return walker.currentNode
    }
  }

  const send = async (cn, fun, params) => {
    const txSwap = cn.getAttribute("tx-swap") ?? ""
    if (txSwap !== "") {
      params.append("tx-swap", txSwap)
    }

    const txParent = cn.getAttribute("tx-pid")
    for (let key in state) {
      if (key.startsWith(txSwap)) {
        params.append(key, JSON.stringify(state[key]))
      }
    }
    if (txParent !== null && state[txParent] !== undefined) {
      params.append(txParent, JSON.stringify(state[txParent]))
    }

    for (let attr of cn.attributes) {
      if (attr.name === 'tx-loc' || attr.name === 'tx-pid') {
        params.append(attr.name, attr.value)
      }
    }

    const res = await fetch("/tx/" + fun, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString() })
    const html = await res.text()

    if (txSwap === '') {
      document.open()
      document.write(html)
      document.close()
      return
    }

    const comp = document.createElement('body')
    comp.innerHTML = html
    const txState = comp.querySelector("#tx-saved")
    const newStates = JSON.parse(txState.textContent)
    state = { ...state, ...newStates }
    comp.removeChild(txState)
    const range = document.createRange()
    const start = findComment('tx:' + txSwap)
    const end = findComment('tx:' + txSwap + '_e')
    range.setStartBefore(start);
    range.setEndAfter(end);
    range.deleteContents();
    for (let child of comp.childNodes) {
      range.insertNode(child.cloneNode(true))
      range.collapse(false)
    }
  }

  const init = (cn) => {
    for (let attr of cn.attributes) {
      if (attr.name.startsWith('tx-on')) {
        const [fun, params] = attr.value.split("?")
        cn.addEventListener(attr.name.slice(5), () => {
          tasks.push(() => send(cn, fun, new URLSearchParams(params)))
          processQueue()
        })
      } else if (attr.name === 'tx-action') {
        const fun = attr.value
        cn.addEventListener('submit', (e) => {
          e.preventDefault()
          const params = new URLSearchParams()
          for (const el of cn.elements) {
            if (!el.name) continue
            if (el.type === 'radio' && !el.checked) continue
            let v
            if (el.type === 'checkbox') v = el.checked ? 'true' : 'false'
            else if (el.type === 'number' || el.type === 'range') v = el.value === '' ? 'null' : el.value
            else v = JSON.stringify(el.value)
            params.append(el.name, v)
          }
          tasks.push(() => send(cn, fun, params))
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
    if (node.nodeType !== Node.ELEMENT_NODE) {
      return
    }

    const walker = document.createTreeWalker(
      node,
      NodeFilter.SHOW_ELEMENT,
      (n) => {
        for (let attr of n.attributes) {
          if (attr.name.startsWith('tx-on') || attr.name === 'tx-action') {
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

type tx_H_addn struct {
	S_counter int `json:"counter"`
}

func render_tx_H_addn(tx_w *bytes.Buffer, tx_id string, counter int, addNum, addNum_swap string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(counter)))
	tx_w.WriteString("</p> ")

	for i := 0; i < 10; i++ {
		tx_w.WriteString("<button tx-key=\"i\" tx-onclick=\"")
		fmt.Fprint(tx_w, addNum)
		tx_w.WriteString("?num=")
		if param, err := json.Marshal(i); err != nil {
			log.Panic(err)
		} else {
			tx_w.WriteString(url.QueryEscape(string(param)))
		}
		tx_w.WriteString("\" tx-swap=\"")
		fmt.Fprint(tx_w, addNum_swap)
		tx_w.WriteString("\"> +")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(i)))
		tx_w.WriteString(" </button>")

	}
	tx_w.WriteString(" <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_cond struct {
	S_num int `json:"num"`
}

func render_tx_H_cond(tx_w *bytes.Buffer, tx_id string, num int) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <button tx-onclick=\"tx-cond:af-1\" tx-swap=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\">change</button> <div> ")
	if num%3 == 0 {
		tx_w.WriteString("<p style=\"background: red; color: white\">red</p> ")
	} else if num%3 == 1 {
		tx_w.WriteString("<p style=\"background: blue; color: white\">blue</p> ")
	} else {
		tx_w.WriteString("<p style=\"background: green; color: white\">green</p> ")

	}
	tx_w.WriteString("</div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_counter struct {
	S_counter int `json:"counter"`
}

func render_tx_H_counter(tx_w *bytes.Buffer, tx_id string, counter int) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <button tx-onclick=\"tx-counter:af-1\" tx-swap=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\">-</button> <span> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(counter)))
	tx_w.WriteString(" </span> <button tx-onclick=\"tx-counter:af-2\" tx-swap=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\">+</button> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_current_H_time struct {
	S_t string `json:"t"`
}

func render_tx_H_current_H_time(tx_w *bytes.Buffer, tx_id string, t string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(t)))
	tx_w.WriteString("</p> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_double struct {
	S_val int `json:"val"`
}

func render_tx_H_double(tx_w *bytes.Buffer, tx_id string, val int) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(val)))
	tx_w.WriteString("</p> <button tx-onclick=\"tx-double:af-1\" tx-swap=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\">double it!</button> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_double_H_state struct {
	S_a int `json:"a"`
	S_b int `json:"b"`
}

func render_tx_H_double_H_state(tx_w *bytes.Buffer, tx_id string, a int, b int) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <div> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(b * a)))
	tx_w.WriteString(" </div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_example_H_wrapper struct {
}

func render_tx_H_example_H_wrapper(tx_w *bytes.Buffer, tx_id string, tx_pid, tx_loc string, tx_render_fill_ func()) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--><div style=\"margin-top: 0.5rem;\n    margin-bottom: 0.5rem;\n    padding: 2rem;\n    display: flex;\n    justify-content: center;\n    align-items: center;\n    border: solid SlateGray;\n    border-radius: 0.25rem;\"> <div> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	}
	tx_w.WriteString(" </div> </div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}
func render_comp_fill_tx_H_example_H_wrapper(tx_w *bytes.Buffer, tx_loc string, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	switch tx_loc {
	case "/{$}_1_":
		tx_saved := &_S__EX_{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S__EX__tx_H_example_H_wrapper_1_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/{$}_2_":
		tx_saved := &_S__EX_{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S__EX__tx_H_example_H_wrapper_2_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/{$}_3_":
		tx_saved := &_S__EX_{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S__EX__tx_H_example_H_wrapper_3_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_1_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_1_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_2_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_2_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_3_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_3_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_4_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_4_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_5_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_5_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_6_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_6_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	case "/docs_7_":
		tx_saved := &_S_docs{}
		json.Unmarshal([]byte(tx_curr_saved[tx_id]), tx_saved)
		render_fill__S_docs_tx_H_example_H_wrapper_7_(tx_w, tx_id, tx_curr_saved, tx_next_saved)
	}
}

type tx_H_greeting struct {
	S_greeting string `json:"greeting"`
}

func render_tx_H_greeting(tx_w *bytes.Buffer, tx_id string, greeting string, greet, greet_swap string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <form tx-action=\"")
	fmt.Fprint(tx_w, greet)
	tx_w.WriteString("\" tx-swap=\"")
	fmt.Fprint(tx_w, greet_swap)
	tx_w.WriteString("\"> <input name=\"name\" type=\"text\" required=\"\"/> <button type=\"submit\">Greet</button> </form> ")
	if greeting != "" {
		tx_w.WriteString("<p>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(greeting)))
		tx_w.WriteString("</p> ")

	}
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_todo struct {
	S_list []string `json:"list"`
}

func render_tx_H_todo(tx_w *bytes.Buffer, tx_id string, list []string, add, add_swap string, remove, remove_swap string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <form tx-action=\"")
	fmt.Fprint(tx_w, add)
	tx_w.WriteString("\" tx-swap=\"")
	fmt.Fprint(tx_w, add_swap)
	tx_w.WriteString("\"> <label><input name=\"item\" type=\"text\" required=\"\"/></label> <button type=\"submit\">Add</button> </form> <ol> ")

	for i, l := range list {
		tx_w.WriteString("<li tx-key=\"l\" tx-onclick=\"")
		fmt.Fprint(tx_w, remove)
		tx_w.WriteString("?i=")
		if param, err := json.Marshal(i); err != nil {
			log.Panic(err)
		} else {
			tx_w.WriteString(url.QueryEscape(string(param)))
		}
		tx_w.WriteString("\" tx-swap=\"")
		fmt.Fprint(tx_w, remove_swap)
		tx_w.WriteString("\"> ")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(l)))
		tx_w.WriteString(" </li>")

	}
	tx_w.WriteString(" </ol> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_triangle struct {
	S_counter int `json:"counter"`
}

func render_tx_H_triangle(tx_w *bytes.Buffer, tx_id string, counter int) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <div> <span> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(counter)))
	tx_w.WriteString(" </span> <button tx-onclick=\"tx-triangle:af-1\" tx-swap=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\">+</button> </div> ")

	for h := 0; h < counter; h++ {
		tx_w.WriteString("<div tx-key=\"h\"> ")

		for s := 0; s < counter-h-1; s++ {
			tx_w.WriteString("<span tx-key=\"s\">_</span>")

		}
		tx_w.WriteString(" ")

		for i := 0; i < h*2+1; i++ {
			tx_w.WriteString("<span tx-key=\"i\">*</span>")

		}
		tx_w.WriteString(" </div>")

	}
	tx_w.WriteString(" <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type _S_docs struct {
}

func render__S_docs(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Docs | tmplx</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script type=\"application/json\" id=\"tx-saved\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	tx_w2.WriteString(runtimeScript)
	tx_w2.WriteString("</script></head> <body> <nav> <h2>tmplx Docs</h2> <ul> <li><a href=\"#introduction\">Introduction</a></li> <li><a href=\"#installing\">Installing</a></li> <li><a href=\"#quick-start\">Quick Start</a></li> <li><a href=\"#pages-and-routing\">Pages and Routing</a></li> <li><a href=\"#tmplx-script\">tmplx Script</a></li> <li> <a href=\"#expression-interpolation\">Expression Interpolation</a> </li> <li><a href=\"#state\">State</a></li> <li><a href=\"#derived\">Derived</a></li> <li><a href=\"#event-handler\">Event Handler</a></li> <li><a href=\"#init\">init()</a></li> <li><a href=\"#path-parameter\">Path Parameter</a></li> <li> <a href=\"#control-flow\">Control Flow</a> <ul> <li><a href=\"#conditionals\">Conditionals</a></li> <li><a href=\"#loops\">Loops</a></li> </ul> </li> <li><a href=\"#template\">&lt;template&gt;</a></li> <li><a href=\"#forms\">Forms</a></li> <li> <a href=\"#component\">Component</a> <ul> <li><a href=\"#props\">Props</a></li> <li><a href=\"#slot\">&lt;slot&gt;</a></li> </ul> </li> <li> Dev Tools <ul> <li><a href=\"#syntax-highlight\">Syntax Highlight</a></li> </ul> </li> </ul> </nav> <main> <h2 id=\"introduction\">Introduction</h2> <p> tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by: </p> <ol> <li> <strong>keeping frontend and backend logic close together</strong> </li> <li> <strong>providing reactive UI updates driven by Go variables</strong> </li> <li><strong>requiring zero new syntax</strong></li> </ol> <p> Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive. </p> ")
	{
		tx_cid := "tx-example-wrapper-1"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_1",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_1_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n\n  func add(item string) {\n    list = append(list, item)\n  }\n\n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;add&#34;&gt;\n  &lt;label&gt;&lt;input name=&#34;item&#34; type=&#34;text&#34; required&gt;&lt;/label&gt;\n  &lt;button type=&#34;submit&#34;&gt;Add&lt;/button&gt;\n&lt;/form&gt;\n&lt;ol&gt;\n  &lt;li\n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code></pre> <p> You start by creating an HTML file. It can be a page or a reusable component, depending on where you place it. </p> <p> You use the <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;</code> tag to embed Go code and make the page or component dynamic. tmplx uses a subset of Go syntax to provide reactive features like <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, and <a href=\"#event-handler\">event handler</a>. At the same time, because the script is valid Go, you can <strong>implement backend logic</strong>—such as database queries—directly in the template. </p> <p> tmplx compiles the HTML templates and embedded Go code into Go functions that render the HTML on the server and generate HTTP handlers for interactive events. On each interaction, the current state is sent to the server, which computes updates and returns both new HTML and the updated state. The result is server-rendered pages with lightweight client-side swapping (similar to <a href=\"https://htmx.org/\">htmx</a>). The interactivity plumbing is handled automatically by the tmplx compiler and runtime—you just implement the features. </p> <p> Most modern web applications separate the frontend and backend into different languages and teams. tmplx eliminates this split by letting you build the entire interactive application in a single language—Go. With this approach, the mental effort needed to track how data flows from the source to the UI is reduced to a minimum. The fewer transformations you perform on your data, the fewer bugs you introduce. </p> <h2 id=\"installing\">Installing</h2> <p>tmplx requires Go 1.24 or later.</p> <pre><code tx-ignore=\"\">$ go install github.com/gnituy18/tmplx@latest</code></pre> <p> This adds tmplx to your Go bin directory (usually $GOPATH/bin or $HOME/go/bin). Make sure that directory is in your PATH. </p> <p>After installation, verify it works:</p> <pre><code tx-ignore=\"\">$ tmplx --help</code></pre> <h2 id=\"quick-start\">Quick Start</h2> <p>Get a tmplx app running in minutes.</p> <ol> <li> <p><strong>Create a project</strong></p> <pre><code tx-ignore=\"\">$ mkdir hello-tmplx\n$ cd hello-tmplx\n$ go mod init hello-tmplx\n$ mkdir pages</code></pre> </li> <li> <p><strong>Add your first page (pages/index.html)</strong></p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n&lt;head&gt;\n  &lt;meta charset=&#34;UTF-8&#34;&gt;\n  &lt;title&gt;Hello tmplx&lt;/title&gt;\n&lt;/head&gt;\n&lt;body&gt;\n  &lt;script type=&#34;text/tmplx&#34;&gt;\n    var count int\n  &lt;/script&gt;\n\n  &lt;h1&gt;Counter&lt;/h1&gt;\n\n  &lt;button tx-onclick=&#34;count--&#34;&gt;-&lt;/button&gt;\n  &lt;span&gt;{ count }&lt;/span&gt;\n  &lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;\n&lt;/body&gt;\n&lt;/html&gt;</code></pre> </li> <li> <p><strong>Generate the Go code</strong></p> <pre><code tx-ignore=\"\">$ tmplx -output-file tmplx/routes.go</code></pre> </li> <li> <p><strong>Create main.go to serve the app</strong></p> <pre><code tx-ignore=\"\">package main\n\nimport (\n\t&#34;log&#34;\n\t&#34;net/http&#34;\n\n\t&#34;hello-tmplx/tmplx&#34;\n)\n\nfunc main() {\n\tfor _, route := range tmplx.Routes() {\n\t\thttp.Handle(route.Pattern, route.Handler)\n\t}\n\n\tlog.Fatal(http.ListenAndServe(&#34;:8080&#34;, nil))\n}</code></pre> </li> <li> <p><strong>Run the server</strong></p> <pre><code tx-ignore=\"\">$ go run .\n&gt; Listening on :8080</code></pre> </li> </ol> <p> That&#39;s it! Open <a href=\"http://localhost:8080\">http://localhost:8080</a> and you now have a working interactive counter. </p> <h2 id=\"pages-and-routing\">Pages and Routing</h2> <p> A <strong>page</strong> is a standalone HTML file that has its own URL in your web app. </p> <p> All pages are placed in the <strong>pages</strong> directory. Default pages location is <code>./pages</code>. Change it with the <code>-pages-dir</code> flag: </p> <pre><code tx-ignore=\"\">$ tmplx -pages-dir=&#34;/some/other/location&#34;</code></pre> <p> tmplx uses <strong>filesystem-based routing</strong>. The route for a page is the relative path of the HTML file inside the <strong>pages</strong> directory, without the <code>.html</code> extension. For example: </p> <ul> <li><code>pages/index.html</code> → <code>/</code></li> <li><code>pages/about.html</code> → <code>/about</code></li> <li> <code>pages/admin/dashboard.html</code> → <code>/admin/dashboard</code> </li> </ul> <p> When the file is named <code>index.html</code>, the <code>index</code> part is omitted from the route (it serves the directory path). To get a route like <code>/index</code>, place <code>index.html</code> in a subdirectory named <code>index</code>. </p> <ul> <li><code>pages/index/index.html</code> → <code>/index</code></li> </ul> <p> Multiple file paths can map to the same route. Choose the style you prefer. Duplicate routes cause compilation failure. </p> <ul> <li><code>pages/login/index.html</code> → <code>/login</code></li> <li><code>pages/login.html</code> → <code>/login</code></li> </ul> <p> To add URL parameters (path wildcards), use curly braces  in directory or file names inside the pages directory. The name inside  must be a valid Go identifier. </p> <ul> <li> <code tx-ignore=\"\">pages/user/{user_id}.html</code> → <code tx-ignore=\"\">/user/{user_id}</code> </li> <li> <code tx-ignore=\"\">pages/blog/{year}/{slug}.html</code> → <code tx-ignore=\"\">/blog/{year}/{slug}</code> </li> </ul> <p> These patterns are compatible with Go&#39;s <code tx-ignore=\"\">net/http.ServeMux</code> (Go 1.22+). The parameter values are available in page initialisation through <code><a href=\"#path-parameter\">tx:path</a></code> comments. </p> <p> tmplx compiles all pages into a single Go file you can import into your Go project. The pages directory can be outside your project, but keeping it inside is recommended. </p> <h2 id=\"tmplx-script\">tmplx Script</h2> <p> <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> is a special tag that you can add to your page or component to declare <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, <a href=\"#event-handler\">event handler</a>, and the special <a href=\"#init\">init()</a> function to control your UI or add backend logic. </p> <p> Each page or component file can have exactly <strong>one</strong> tmplx script. Multiple scripts cause a compilation error. </p> <p> In pages, place it anywhere inside <code>&lt;head&gt;</code> or <code>&lt;body&gt;</code>. </p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n  &lt;head&gt;\n    ...\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // Go code here\n    &lt;/script&gt;\n    ...\n  &lt;/head&gt;\n  &lt;body&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code> </pre> <p>In components, place it at the root level.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  // Go code here\n&lt;/script&gt;\n...\n...</code></pre> <h2 id=\"expression-interpolation\">Expression Interpolation</h2> <p> Use curly braces <code tx-ignore=\"\">{}</code> to insert <a href=\"https://go.dev/ref/spec#Expressions\">Go expressions</a> into HTML. Expressions are allowed only in: </p> <ul> <li><strong>text nodes</strong></li> <li><strong>attribute values</strong></li> </ul> <p>Placing expressions anywhere else causes a parsing error.</p> <p tx-ignore=\"\">\n        tmplx converts expression results to strings using\n        <code><a href=\"https://pkg.go.dev/fmt#Sprint\">fmt.Sprint</a></code>. The difference is that in <strong>text nodes</strong> the output is\n        <strong>HTML-escaped</strong> to prevent cross-site scripting (XSS)\n        attacks.\n      </p> <p> Expressions run on the server every time the page loads or a component re-renders after an event. Avoid side effects in expressions, such as database queries or heavy computations, because they execute on every render. </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p class=&#39;{ strings.Join([]string{&#34;c1&#34;, &#34;c2&#34;}, &#34; &#34;) }&#39;&gt;\n Hello, { user.GetNameById(0) }!\n&lt;/p&gt;</code> </pre> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p class=&#34;c1 c2&#34;&gt;\n Hello, tmplx!\n&lt;/p&gt;</code></pre> <p tx-ignore=\"\">\n        Add the <code>tx-ignore</code> attribute to an element to disable\n        expression interpolation in that element&#39;s attributes and its direct\n        text children. Descendant elements are still processed normally.\n      </p> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;{ &#34;not&#34; + &#34; ignored&#34; }&lt;/span&gt;\n&lt;/p&gt;</code> </pre> <pre><code tx-ignore=\"\" class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;not ignored&lt;/span&gt;\n&lt;/p&gt;</code></pre> <h2 id=\"state\">State</h2> <p> <strong>State</strong> is the mutable data that describes a component&#39;s current condition. </p> <p> Declaring state works like declaring variables in Go&#39;s package scope. If you provide no initial value, the state starts with the zero value for its type. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string\n&lt;/script&gt;</code></pre> <p>To set an initial value, use the <code>=</code> operator.</p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string = &#34;tmplx&#34;\n&lt;/script&gt;</code></pre> <p>Although the syntax follows valid Go code, these rules apply:</p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong>The type must be explicitly declared and JSON-compatible.</strong> </li> </ol> <p> The 1st rule is enforced by the compiler. The 2nd is not checked at compile time (for now) and will cause a runtime error if violated. </p> <h3>Some invalid state declarations:</h3> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n// ❌ Must explicitly declare the type\nvar str = &#34;&#34;\n\n// ❌ Cannot use the := short declaration\nnum := 1\n\n// ❌ Type must be JSON-marshalable/unmarshalable\nvar f func(int) = func(i int) { ... }\nvar w io.Writer\n\n// ❌ Only one identifier per declaration\nvar a, b int = 10, 20\nvar a, b int = f()\n&lt;/script&gt;</code></pre> <h3>Some valid state declarations:</h3> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n// ✅ Zero value\nvar id int64\n\n// ✅ With initial value\nvar address string = &#34;...&#34;\n\n// ✅ Initialized with a function call (assuming the package is imported)\nvar username string = user.GetNameById(&#34;id&#34;)\n\n// ✅ Complex JSON-compatible types\nvar m map[string]int = map[string]int{&#34;key&#34;: 100}\n&lt;/script&gt;</code></pre> <h2 id=\"derived\">Derived</h2> A <strong>derived</strong> is a <strong>read-only</strong> value that is automatically calculated from states. It updates whenever those states change. <p> Declaring a derived works the same way as declaring package-level variables in Go. When the right-hand side of the declaration <strong>references existing state or other derived values</strong>, it is treated as a derived value. </p> <p> Derived values follow most of the same rules as regular state variables, but with some differences: </p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li><strong>The type must be specified explicitly.</strong></li> <li> <strong>Derived values cannot be modified directly in event handlers, though they may be read.</strong> </li> </ol> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num1 int = 100 // state\n  var num2 int = num1 * 2 // derived\n&lt;/script&gt;\n\n...\n&lt;p&gt;{num1} * 2 = {num2}&lt;/p&gt;</code></pre> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var classStrs []string = []string{&#34;c1&#34;, &#34;c2&#34;, &#34;c3&#34;} // state\n  var class string = strings.Join(classStrs, &#34; &#34;) // derived\n&lt;/script&gt;\n\n...\n&lt;p class=&#34;{class}&#34;&gt; ... &lt;/p&gt;</code></pre> <h2 id=\"event-handler\">Event Handler</h2> <p> Event handlers let you respond to frontend events with backend logic or update state to trigger UI changes. </p> <p> To declare an event handler, define a Go function in the global scope of the <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> block. Bind it to a DOM event by adding an attribute that starts with <code>tx-on</code> followed by the event name (e.g., <code>tx-onclick</code>). </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func add1() {\n    counter += 1\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-onclick=&#34;add1()&#34;&gt;Add 1&lt;/button&gt;</code></pre> <p> In this example, the <code>add1</code> handler runs every time the button is clicked. The <code>counter</code> state increases by 1, and the paragraph updates automatically. </p> <p> It’s not magic. tmplx compiles each event handler into an HTTP endpoint. The runtime JavaScript attaches a lightweight listener that sends the required state to the endpoint, receives the updated HTML fragment, merges the new state, and swaps the affected part of the DOM. It feels like direct backend access from the client, but it’s just a simple API call with targeted DOM swapping. </p> <h3>Arguments</h3> You can add arguments from local variable declared within <code>tx-if</code>, or <code>tx-for</code> with the following rules: <ul> <li> <strong>Argument names cannot match state or derived state names. </strong> </li> <li><strong>Argument types must be JSON-compatible.</strong></li> </ul> ")
	{
		tx_cid := "tx-example-wrapper-2"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_2",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_2_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func addNum(num int) {\n    counter += num\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-for=&#34;i := 0; i &lt; 10; i++&#34; tx-key=&#34;i&#34; tx-onclick=&#34;addNum(i)&#34;&gt;\n  +{ i }\n&lt;/button&gt;</code></pre> <h3>Inline Statements</h3> <p> For simple actions, embed Go statements directly in <code>tx-on*</code> attributes to update state. This avoids defining separate handler functions. </p> ")
	{
		tx_cid := "tx-example-wrapper-3"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_3",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_3_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var val int = 1\n&lt;/script&gt;\n\n&lt;p&gt;{ val }&lt;/p&gt;\n&lt;button tx-onclick=&#34;val *= 2&#34;&gt;double it!&lt;/button&gt;</code> </pre> <h2 id=\"init\">init()</h2> <p> <code>init()</code> is a special function that runs automatically the first time a page or component is rendered. For pages, it runs on every GET request. For components, it runs when the component has no saved state yet (for example, the first time it appears on the page, or the first time a new <code>tx-for</code> iteration produces it). After that, subsequent renders reuse the saved state and skip <code>init()</code>. </p> ")
	{
		tx_cid := "tx-example-wrapper-4"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_4",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_4_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var t string\n\n  func init() {\n    t = fmt.Sprint(time.Now().Format(time.RFC3339))\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ t }&lt;/p&gt;</code></pre> <p> Another common use case is to initialize one state from another state without turning the second variable into a derived state. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var a int = 1\n  var b int\n\n  func init() {\n    b = a * 2 // b remains a regular state\n  }\n&lt;/script&gt;</code></pre> <h2 id=\"path-parameter\">Path Parameters</h2> <p> You can inject path parameters into states using a <code>//tx:path</code> comment placed directly above the state declaration. This feature works only in <a href=\"#pages-and-routing\">pages</a> and requires the state to be of type <code>string</code>. </p> <p> For example, given a route pattern like <code tx-ignore=\"\">/blog/post/{post_id}</code>, you can access the <code>post_id</code> parameter as follows: </p> <pre><code tx-ignore=\"\">&lt;!DOCTYPE html&gt;\n&lt;html&gt;\n  &lt;head&gt;\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // tx:path post_id\n      var postId string\n\n      var post Post\n      \n      func init() {\n        post = db.GetPost(postId)\n      }\n    &lt;/script&gt;\n  &lt;/head&gt;\n\n  &lt;body&gt;\n    &lt;h1&gt;{ post.Title }&lt;/h1&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code></pre> <p> The value of the <code>post_id</code> path parameter is automatically injected into the <code>postId</code> state during initialization. After that, <code>postId</code> behaves like any other state and can be read or modified as needed. </p> <h2 id=\"control-flow\">Control Flow</h2> <p> tmplx avoids new custom syntax for conditionals and loops because that would increase compiler complexity. Instead, it embeds control flow directly into HTML attributes, similar to Vue.js and <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h3 id=\"conditionals\">Conditionals</h3> <p> To conditionally render elements, use the <code>tx-if</code>, <code>tx-else-if</code>, and <code>tx-else</code> attributes on the desired tags. The values for <code>tx-if</code> and <code>tx-else-if</code> can be any valid Go expression that would fit in an <code>if</code> or <code>else if</code> statement. The <code>tx-else</code> attribute needs no value. </p> ")
	{
		tx_cid := "tx-example-wrapper-5"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_5",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_5_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;num++&#34;&gt;change&lt;/button&gt;\n&lt;div&gt;\n  &lt;p tx-if=&#34;num % 3 == 0&#34; style=&#34;background: red; color: white&#34;&gt;red&lt;/p&gt;\n  &lt;p tx-else-if=&#34;num % 3 == 1&#34; style=&#34;background: blue; color: white&#34;&gt;blue&lt;/p&gt;\n  &lt;p tx-else style=&#34;background: green; color: white&#34;&gt;green&lt;/p&gt;\n&lt;/div&gt;</code> </pre> <p> You can declare <strong>local variables</strong> and handle errors exactly as you would in regular Go code. Local variables declared in conditionals are available to the element and its descendants, just like in Go. </p> <pre><code tx-ignore=\"\">&lt;p tx-if=&#34;user, err := user.GetUser(); err != nil&#34;&gt;\n  &lt;span tx-if=&#34;err == ErrNotFound&#34;&gt;User not found&lt;/span&gt;\n&lt;/p&gt;\n&lt;p tx-else-if=&#39;user.Name == &#34;&#34;&#39;&gt;user.Name not set&lt;/p&gt;\n&lt;p tx-else&gt;Hi, { user.Name }&lt;/p&gt;</code></pre> <p> A conditional group consists of <strong>consecutive sibling nodes</strong> that share the same parent. Disconnected nodes are not treated as part of the same group. A standalone <code>tx-else-if</code> or <code>tx-else</code> without a preceding <code>tx-if</code> will cause a compilation error. </p> <h3 id=\"loops\">Loops</h3> <p> To repeat elements, use the <code>tx-for</code> attribute. Its value can be any valid Go <code>for</code> statement, including <strong>classic for</strong> or <strong>range for</strong>. </p> <p> Local variables declared in the loop are available to the element and all of its descendants, just like in Go. </p> <p> Always add a <code>tx-key</code> attribute with a unique value for each item. This gives the compiler a unique identifier for the node during updates. </p> ")
	{
		tx_cid := "tx-example-wrapper-6"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_6",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_6_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code> </pre> <pre><code tx-ignore=\"\">&lt;div tx-for=&#34;_, user := range users&#34;&gt;\n  { user.Id }: { user.Name }\n&lt;/div&gt;</code></pre> <h2 id=\"template\">&lt;template&gt;</h2> <p> The <code>&lt;template&gt;</code> tag is a non-rendering container that lets you apply control flow attributes (<code>tx-if</code>, <code>tx-else-if</code>, <code>tx-else</code>, or <code>tx-for</code>) to a group of elements at once. </p> <p> The <code>&lt;template&gt;</code> itself is removed from the output; only its children are rendered (or not, depending on the control flow). </p> <p> You can nest <code>&lt;template&gt;</code> tags and combine them with other control flow attributes on child elements. </p> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var loggedIn bool = true\n&lt;/script&gt;\n\n&lt;template tx-if=&#34;loggedIn&#34;&gt;\n  &lt;p&gt;Welcome back!&lt;/p&gt;\n  &lt;button tx-onclick=&#34;logout()&#34;&gt;Logout&lt;/button&gt;\n&lt;/template&gt;\n\n&lt;template tx-else&gt;\n  &lt;p&gt;Please sign in.&lt;/p&gt;\n  &lt;button tx-onclick=&#34;login()&#34;&gt;Login&lt;/button&gt;\n&lt;/template&gt;</code> </pre> <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var posts []Post = []Post{\n    {Title: &#34;First Post&#34;, Body: &#34;Hello world&#34;},\n    {Title: &#34;Second Post&#34;, Body: &#34;tmplx is great&#34;},\n  }\n&lt;/script&gt;\n\n&lt;template tx-for=&#34;i, p := range posts&#34; tx-key=&#34;i&#34;&gt;\n  &lt;article&gt;\n    &lt;h3&gt;{ p.Title }&lt;/h3&gt;\n    &lt;p&gt;{ p.Body }&lt;/p&gt;\n    &lt;hr&gt;\n  &lt;/article&gt;\n&lt;/template&gt;</code> </pre> <h2 id=\"forms\">Forms</h2> <p> Attach a handler to a <code>&lt;form&gt;</code> with <code>tx-action</code>. When the form is submitted, tmplx cancels the default submission, collects every named form element, and calls the handler on the server. </p> <p> The value of <code>tx-action</code> must be the name of a function declared in the tmplx script. Each form element&#39;s <code>name</code> attribute must match a parameter name on that function; unnamed elements are ignored. </p> ")
	{
		tx_cid := "tx-example-wrapper-7"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/docs_7",
			func() { render_fill__S_docs_tx_H_example_H_wrapper_7_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var greeting string\n\n  func greet(name string) {\n    greeting = &#34;Hello, &#34; + name\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;greet&#34;&gt;\n  &lt;input name=&#34;name&#34; type=&#34;text&#34; required /&gt;\n  &lt;button type=&#34;submit&#34;&gt;Greet&lt;/button&gt;\n&lt;/form&gt;\n\n&lt;p tx-if=&#39;greeting != &#34;&#34;&#39;&gt;{ greeting }&lt;/p&gt;</code></pre> <p> Values are JSON-decoded into each parameter&#39;s Go type, so the parameter type is what determines how the string is parsed. The runtime serializes form elements by input type: </p> <ul> <li> <code>text</code>, <code>email</code>, <code>password</code>, <code>textarea</code>, <code>select</code>, etc.—sent as a JSON string. Decode into <code>string</code>. </li> <li> <code>number</code>, <code>range</code>—sent as the raw numeric value, or <code>null</code> when empty. Decode into a numeric type or pointer. </li> <li> <code>checkbox</code>—sent as <code>true</code> or <code>false</code>. Decode into <code>bool</code>. </li> <li> <code>radio</code>—only the checked radio in a group is sent (using its shared <code>name</code>). Decode into <code>string</code>. </li> </ul> <p> Because submission goes through a full server round-trip, use native HTML validation (<code>required</code>, <code>minlength</code>, <code>pattern</code>, ...) to catch client-side errors before the request is sent. For richer live-updating inputs, combine tmplx with a client-side library like <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h2 id=\"component\">Component</h2> <p> Components are reusable UI building blocks that encapsulate HTML, state, and behavior. </p> <p> Create a component by placing an <code>.html</code> file in the <code>components</code> directory (default: <code>./components</code>). tmplx automatically registers it as a custom element with the tag name <code>tx-</code> followed by the lowercase kebab-case version of the relative path (without the <code>.html</code> extension). </p> <p>Examples:</p> <ul> <li> <code>components/Button.html</code> → <code>&lt;tx-button&gt;</code> </li> <li> <code>components/user/Card.html</code> → <code>&lt;tx-user-card&gt;</code> </li> <li> <code>components/todo/List.html</code> → <code>&lt;tx-todo-list&gt;</code> </li> </ul> <p> Components can contain their own <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> for local state and logic, and can be used in pages or nested inside other components. </p> <h3 id=\"props\">Props</h3> <p> Pass data to a component via attributes. Inside the component, declare matching state variables with <code>// tx:prop</code> comments to make them reactive to prop changes. </p> <pre><code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  // tx:prop\n  var title string\n\n  // tx:prop\n  var count int = 0\n&lt;/script&gt;\n\n&lt;h3&gt;{ title }&lt;/h3&gt;\n&lt;span&gt;{ count }&lt;/span&gt;\n&lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;</code></pre> <p> Prop attribute values are evaluated as <strong>Go expressions</strong>, not as plain strings. Pass a literal by writing the Go literal directly, or pass a state/derived/prop variable from the parent by name: </p> <pre><code tx-ignore=\"\">&lt;!-- Go string literal (note the inner quotes) and numeric literal --&gt;\n&lt;tx-my-component title=&#39;&#34;Hello&#34;&#39; count=&#34;5&#34;&gt;&lt;/tx-my-component&gt;\n\n&lt;!-- Pass parent state/derived/prop variables --&gt;\n&lt;tx-my-component title=&#34;heading&#34; count=&#34;itemCount&#34;&gt;&lt;/tx-my-component&gt;</code></pre> <h3 id=\"slot\">&lt;slot&gt;</h3> <p> Use <code>&lt;slot&gt;</code> to define insertion points for child content. Name slots for multiple insertion points. </p> <pre><code tx-ignore=\"\">&lt;div class=&#34;card&#34;&gt;\n  &lt;slot name=&#34;header&#34;&gt;Default Header&lt;/slot&gt;\n  &lt;div class=&#34;body&#34;&gt;\n    &lt;slot&gt;Default Body&lt;/slot&gt;\n  &lt;/div&gt;\n  &lt;slot name=&#34;footer&#34;&gt;&lt;/slot&gt;\n&lt;/div&gt;</code></pre> <p>Usage:</p> <pre><code tx-ignore=\"\">&lt;tx-card&gt;\n  &lt;h2 slot=&#34;header&#34;&gt;Custom Title&lt;/h2&gt;\n  &lt;p&gt;Custom content&lt;/p&gt;\n  &lt;div slot=&#34;footer&#34;&gt;Actions&lt;/div&gt;\n&lt;/tx-card&gt;</code></pre> <p> Unnamed slots receive content without a <code>slot</code> attribute. Fallback content inside <code>&lt;slot&gt;</code> renders when no matching content is provided. </p> <h2 id=\"syntax-highlight\">Syntax Highlight</h2> <a href=\"https://github.com/gnituy18/tmplx.nvim\">Neovim Plugin</a> </main> </body></html>")
}
func render_fill__S_docs_tx_H_example_H_wrapper_1_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-todo-1"
		tx_saved := &tx_H_todo{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_todo(tx_w, tx_cid, tx_saved.S_list, "tx-todo:add", tx_cid, "tx-todo:remove", tx_cid)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_2_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-addn-1"
		tx_saved := &tx_H_addn{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		} else {
			tx_saved.S_counter = 0
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_addn(tx_w, tx_cid, tx_saved.S_counter, "tx-addn:addNum", tx_cid)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_3_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-double-1"
		tx_saved := &tx_H_double{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		} else {
			tx_saved.S_val = 1
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_double(tx_w, tx_cid, tx_saved.S_val)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_4_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-current-time-1"
		tx_saved := &tx_H_current_H_time{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		} else {
			tx_saved.S_t = fmt.Sprint(time.Now().Format(time.RFC3339))
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_current_H_time(tx_w, tx_cid, tx_saved.S_t)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_5_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-cond-1"
		tx_saved := &tx_H_cond{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_cond(tx_w, tx_cid, tx_saved.S_num)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_6_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-triangle-1"
		tx_saved := &tx_H_triangle{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		} else {
			tx_saved.S_counter = 5
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_triangle(tx_w, tx_cid, tx_saved.S_counter)
	}
	tx_w.WriteString(" ")
}
func render_fill__S_docs_tx_H_example_H_wrapper_7_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-greeting-1"
		tx_saved := &tx_H_greeting{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_greeting(tx_w, tx_cid, tx_saved.S_greeting, "tx-greeting:greet", tx_cid)
	}
	tx_w.WriteString(" ")
}

type _S_examples_S__EX_ struct {
}

func render__S_examples_S__EX_(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<html><head> <title>tmplx fixture</title> <script type=\"application/json\" id=\"tx-saved\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	tx_w2.WriteString(runtimeScript)
	tx_w2.WriteString("</script></head> <body> <h1>tmplx fixture</h1> <ul> <li><a href=\"/state\">state</a> — state variables, initial values, interpolation</li> </ul> </body></html>")
}

type _S_examples_S_state struct {
	S_count int    `json:"count"`
	S_label string `json:"label"`
	S_flag  bool   `json:"flag"`
}

func render__S_examples_S_state(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer, count int, label string, flag bool) {
	tx_w1.WriteString("<html><head>  <title>state</title> <script type=\"application/json\" id=\"tx-saved\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	tx_w2.WriteString(runtimeScript)
	tx_w2.WriteString("</script></head> <body> <h1>state</h1> <p>int state with initial value: <b id=\"count\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(count)))
	tx_w2.WriteString("</b> (expect: 42)</p> <p>string state with initial value: <b id=\"label\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(label)))
	tx_w2.WriteString("</b> (expect: hello)</p> <p>bool state with initial value: <b id=\"flag\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(flag)))
	tx_w2.WriteString("</b> (expect: true)</p> </body></html>")
}

type _S__EX_ struct {
}

func render__S__EX_(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script type=\"application/json\" id=\"tx-saved\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	tx_w2.WriteString(runtimeScript)
	tx_w2.WriteString("</script></head> <body> <main> <h1 style=\"text-align: center\">&lt;tmplx&gt;</h1> <h2 style=\"text-align: center; margin-top: 1.5rem\"> Write Go in HTML intuitively </h2> <ul style=\"margin-top: 4rem\"> <li>Full Go backend logic and HTML in the same file</li> <li>Reactive UIs driven by plain Go variables</li> <li>Reusable components written as regular HTML files</li> </ul> <div style=\"display: flex;\n          gap: 2rem;\n          justify-content: center;\n          text-align: center;\n          margin-top: 4rem;\"> <a class=\"btn\" href=\"/docs\">Docs</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> <h2 style=\"text-align: center\">Demos</h2> <h3>Counter</h3> ")
	{
		tx_cid := "tx-example-wrapper-1"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/{$}_1",
			func() { render_fill__S__EX__tx_H_example_H_wrapper_1_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;counter--&#34;&gt;-&lt;/button&gt;\n&lt;span&gt; { counter } &lt;/span&gt;\n&lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;</code> </pre> <h3>To Do</h3> ")
	{
		tx_cid := "tx-example-wrapper-2"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/{$}_2",
			func() { render_fill__S__EX__tx_H_example_H_wrapper_2_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\" class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n\n  func add(item string) {\n    list = append(list, item)\n  }\n\n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;add&#34;&gt;\n  &lt;label&gt;&lt;input name=&#34;item&#34; type=&#34;text&#34; required&gt;&lt;/label&gt;\n  &lt;button type=&#34;submit&#34;&gt;Add&lt;/button&gt;\n&lt;/form&gt;\n&lt;ol&gt;\n  &lt;li\n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code> </pre> <h3>Triangle</h3> ")
	{
		tx_cid := "tx-example-wrapper-3"
		tx_next_saved[tx_cid] = &tx_H_example_H_wrapper{}
		render_tx_H_example_H_wrapper(tx_w2, tx_cid, "page", "/{$}_3",
			func() { render_fill__S__EX__tx_H_example_H_wrapper_3_(tx_w2, tx_cid, tx_curr_saved, tx_next_saved) },
		)
	}
	tx_w2.WriteString(" <pre> <code tx-ignore=\"\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code> </pre> </main> </body></html>")
}
func render_fill__S__EX__tx_H_example_H_wrapper_1_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-counter-1"
		tx_saved := &tx_H_counter{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_counter(tx_w, tx_cid, tx_saved.S_counter)
	}
	tx_w.WriteString(" ")
}
func render_fill__S__EX__tx_H_example_H_wrapper_2_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-todo-1"
		tx_saved := &tx_H_todo{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_todo(tx_w, tx_cid, tx_saved.S_list, "tx-todo:add", tx_cid, "tx-todo:remove", tx_cid)
	}
	tx_w.WriteString(" ")
}
func render_fill__S__EX__tx_H_example_H_wrapper_3_(tx_w *bytes.Buffer, tx_id string, tx_curr_saved map[string]string, tx_next_saved map[string]any) {
	tx_w.WriteString(" ")
	{
		tx_cid := tx_id + "@tx-triangle-1"
		tx_saved := &tx_H_triangle{}
		tx_curr_saved_str, tx_curr_saved_exist := tx_curr_saved[tx_cid]
		if tx_curr_saved_exist {
			json.Unmarshal([]byte(tx_curr_saved_str), tx_saved)
		} else {
			tx_saved.S_counter = 5
		}
		tx_next_saved[tx_cid] = tx_saved
		render_tx_H_triangle(tx_w, tx_cid, tx_saved.S_counter)
	}
	tx_w.WriteString(" ")
}

type TxRoute struct {
	Pattern string
	Handler http.HandlerFunc
}

var txRoutes []TxRoute = []TxRoute{
	{
		Pattern: "GET /docs",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_saved := &_S_docs{}
			tx_next_saved := map[string]any{"page": tx_saved}
			var tx_buf1, tx_buf2 bytes.Buffer
			render__S_docs(&tx_buf1, &tx_buf2, map[string]string{}, tx_next_saved)
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_savedBytes)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /examples/{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_saved := &_S_examples_S__EX_{}
			tx_next_saved := map[string]any{"page": tx_saved}
			var tx_buf1, tx_buf2 bytes.Buffer
			render__S_examples_S__EX_(&tx_buf1, &tx_buf2)
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_savedBytes)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /examples/state",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_saved := &_S_examples_S_state{}
			tx_saved.S_count = 42
			tx_saved.S_label = "hello"
			tx_saved.S_flag = true
			tx_next_saved := map[string]any{"page": tx_saved}
			var tx_buf1, tx_buf2 bytes.Buffer
			render__S_examples_S_state(&tx_buf1, &tx_buf2, tx_saved.S_count, tx_saved.S_label, tx_saved.S_flag)
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_savedBytes)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_saved := &_S__EX_{}
			tx_next_saved := map[string]any{"page": tx_saved}
			var tx_buf1, tx_buf2 bytes.Buffer
			render__S__EX_(&tx_buf1, &tx_buf2, map[string]string{}, tx_next_saved)
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_savedBytes)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "POST /tx/tx-addn:addNum",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_addn{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			var num int
			json.Unmarshal([]byte(tx_r.PostFormValue("num")), &num)
			tx_saved.S_counter += num
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_addn(&tx_buf, tx_id, tx_saved.S_counter, "tx-addn:addNum", tx_id)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-cond:af-1",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_cond{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			tx_saved.S_num++
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_cond(&tx_buf, tx_id, tx_saved.S_num)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-counter:af-1",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_counter{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			tx_saved.S_counter--
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_counter(&tx_buf, tx_id, tx_saved.S_counter)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-counter:af-2",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_counter{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			tx_saved.S_counter++
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_counter(&tx_buf, tx_id, tx_saved.S_counter)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-double:af-1",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_double{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			tx_saved.S_val *= 2
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_double(&tx_buf, tx_id, tx_saved.S_val)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-greeting:greet",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_greeting{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			var name string
			json.Unmarshal([]byte(tx_r.PostFormValue("name")), &name)
			tx_saved.S_greeting = "Hello, " + name
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_greeting(&tx_buf, tx_id, tx_saved.S_greeting, "tx-greeting:greet", tx_id)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-todo:add",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_todo{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			var item string
			json.Unmarshal([]byte(tx_r.PostFormValue("item")), &item)
			tx_saved.S_list = append(tx_saved.S_list, item)
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_todo(&tx_buf, tx_id, tx_saved.S_list, "tx-todo:add", tx_id, "tx-todo:remove", tx_id)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-todo:remove",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_todo{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			var i int
			json.Unmarshal([]byte(tx_r.PostFormValue("i")), &i)
			tx_saved.S_list = append(tx_saved.S_list[0:i], tx_saved.S_list[i+1:]...)
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_todo(&tx_buf, tx_id, tx_saved.S_list, "tx-todo:add", tx_id, "tx-todo:remove", tx_id)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
	{
		Pattern: "POST /tx/tx-triangle:af-1",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_r.ParseForm()
			tx_id := tx_r.PostFormValue("tx-swap")
			tx_curr_saved := map[string]string{}
			for k, v := range tx_r.PostForm {
				if k != "tx-swap" {
					tx_curr_saved[k] = v[0]
				}
			}
			tx_next_saved := map[string]any{}
			tx_saved := &tx_H_triangle{}
			json.Unmarshal([]byte(tx_curr_saved[tx_id]), &tx_saved)
			tx_saved.S_counter++
			tx_next_saved[tx_id] = tx_saved
			var tx_buf bytes.Buffer
			render_tx_H_triangle(&tx_buf, tx_id, tx_saved.S_counter)
			tx_w.Write(tx_buf.Bytes())
			tx_w.Write([]byte("<script id=\"tx-saved\" type=\"application/json\">"))
			tx_savedBytes, _ := json.Marshal(tx_next_saved)
			tx_w.Write(tx_savedBytes)
			tx_w.Write([]byte("</script>"))
		},
	},
}

func Routes() []TxRoute { return txRoutes }
