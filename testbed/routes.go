package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
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
type state__S__EX_ struct {
}

func render__S__EX_(tx_w *bytes.Buffer, tx_key string, tx_states map[string]string, tx_newStates map[string]any) {
	tx_w.WriteString("<html><head> <title>tmplx fixture</title> <script id=\"tx-runtime\">")
	tx_w.WriteString(runtimeScript)
	tx_w.WriteString("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <h1>tmplx fixture</h1> <ul> <li><a href=\"/state\">state</a> — state variables, initial values, interpolation</li> </ul> </body></html>")
}

type state__S_state struct {
	S_count int    `json:"count"`
	S_label string `json:"label"`
	S_flag  bool   `json:"flag"`
}

func render__S_state(tx_w *bytes.Buffer, tx_key string, tx_states map[string]string, tx_newStates map[string]any, count int, label string, flag bool) {
	tx_w.WriteString("<html><head>  <title>state</title> <script id=\"tx-runtime\">")
	tx_w.WriteString(runtimeScript)
	tx_w.WriteString("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <h1>state</h1> <p>int state with initial value: <b id=\"count\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(count)))
	tx_w.WriteString("</b> (expect: 42)</p> <p>string state with initial value: <b id=\"label\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(label)))
	tx_w.WriteString("</b> (expect: hello)</p> <p>bool state with initial value: <b id=\"flag\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(flag)))
	tx_w.WriteString("</b> (expect: true)</p> </body></html>")
}

var txRoutes []TxRoute = []TxRoute{
	{
		Pattern: "GET /{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_state := &state__S__EX_{}
			tx_newStates := map[string]any{}
			tx_newStates["tx_"] = tx_state
			var tx_buf bytes.Buffer
			render__S__EX_(&tx_buf, "tx_", map[string]string{}, tx_newStates)
			tx_stateBytes, _ := json.Marshal(tx_newStates)
			tx_w.Write(bytes.Replace(tx_buf.Bytes(), []byte("TX_STATE_JSON"), tx_stateBytes, 1))
		},
	},
	{
		Pattern: "GET /state",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			var count int = 42
			var label string = "hello"
			var flag bool = true
			tx_state := &state__S_state{
				S_count: count,
				S_label: label,
				S_flag:  flag,
			}
			tx_newStates := map[string]any{}
			tx_newStates["tx_"] = tx_state
			var tx_buf bytes.Buffer
			render__S_state(&tx_buf, "tx_", map[string]string{}, tx_newStates, count, label, flag)
			tx_stateBytes, _ := json.Marshal(tx_newStates)
			tx_w.Write(bytes.Replace(tx_buf.Bytes(), []byte("TX_STATE_JSON"), tx_stateBytes, 1))
		},
	},
}

func Routes() []TxRoute { return txRoutes }
