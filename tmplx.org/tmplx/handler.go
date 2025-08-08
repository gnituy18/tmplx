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
	w.Write([]byte(`<!DOCTYPE html><html lang="en"><head>
  <title>a tmplx </title>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css"/>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/prismjs@1.30.0/themes/prism.min.css" integrity="sha256-ko4j5rn874LF8dHwW29/xabhh8YBleWfvxb8nQce4Fc=" crossorigin="anonymous"/>
  <script src="https://cdn.jsdelivr.net/npm/prismjs@1.30.0/prism.min.js"></script>
  <style>
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
  </style>
<script id="tx-runtime">`))
	w.Write([]byte(runtimeScript))
	w.Write([]byte(`</script><script type="application/json" id="tx-state">`))
	w.Write([]byte(fmt.Sprint(state)))
	w.Write([]byte(`</script></head>


<body>
  <main>
    <h1 style="text-align:center"> &lt;tmplx&gt; </h1>
    <p> <mark>tmplx is a compile-time framework for building state-driven web apps. </mark>
    </p>
    <p>
      Embed Go code in HTML to define states and
      event handlers, which compiles into Go handlers that update state, rerender specific UI sections, and return HTML
      snippets. This embraces hypermedia by having the server drive UI updates via direct HTML responses.</p>
    <pre>      <code tx-ignore="" class="language-html">
&lt;script type=&#34;text/tmplx&#34;&gt;
  // name is declared as a state
  var name string = &#34;tmplx&#34;
  // greeting is declared as a derived
  var greeting string = fmt.Sprintf(&#34;Hello ,%s!&#34;, name)

  var counter int = 0

  // addOne event handler
  func addOne() {
    counter++
  }
&lt;/script&gt;

&lt;html&gt;

&lt;head&gt;
  &lt;title&gt; { name } &lt;/title&gt;
&lt;/head&gt;

&lt;body&gt;
  &lt;h1&gt; { greeting } &lt;/h1&gt;

  &lt;p&gt;counter: { counter }&lt;/p&gt;
  &lt;p&gt;counter * 10 = { counterTimes10 }&lt;/p&gt;

  &lt;button tx-onclick=&#34;addOne()&#34;&gt;Add 1&lt;/button&gt;
  &lt;button tx-onclick=&#34;counter--&#34;&gt;Subtract 1&lt;/button&gt;

  &lt;p tx-if=&#34;i % 2 == 0&#34;&gt; counter is even &lt;/p&gt;
  &lt;p tx-else&gt; counter is odd &lt;/p&gt;

  &lt;p tx-for=&#34;i := 0; i &lt; 10; i++&#34;&gt; { i } &lt;/p&gt;

  &lt;a href=&#34;/second-page&#34;&gt;second page&lt;/a&gt;

&lt;/body&gt;

&lt;/html&gt;
      </code>
    </pre>
  </main>



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
