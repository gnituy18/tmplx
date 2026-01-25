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
)

var runtimeScript = `document.addEventListener('DOMContentLoaded', function() {
  const handlerPrefix = "/tx"
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

            const res = await fetch(handlerPrefix + "/" + fun + "?" + searchParams.toString())
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
type state_tx_2D_button struct {
	S_name string `json:"name"`
}

func render_tx_2D_button(w io.Writer, key string, states map[string]string, newStates map[string]any, name string, handleClick, handleClick_swap string) {
	w.Write([]byte("<template id=\""))
	fmt.Fprint(w, key)
	w.Write([]byte("\"></template>  <button tx-onclick=\""))
	fmt.Fprint(w, handleClick)
	w.Write([]byte("\" tx-swap=\""))
	fmt.Fprint(w, handleClick_swap)
	w.Write([]byte("\"> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(name))))
	w.Write([]byte(" </button>   <template id=\""))
	fmt.Fprint(w, key+"_e")
	w.Write([]byte("\"></template>"))
}

type state_ struct {
	S_name           string `json:"name"`
	S_greeting       string `json:"greeting"`
	S_path           string `json:"path"`
	S_counter        int    `json:"counter"`
	S_counterTimes10 int    `json:"counterTimes10"`
	S_str            string `json:"str"`
}

func render_(w io.Writer, key string, states map[string]string, newStates map[string]any, name string, greeting string, path string, counter int, counterTimes10 int, str string, addOne, addOne_swap string, appendS, appendS_swap string, _1, _1_swap string) {
	w.Write([]byte("<html><head> <title> "))
	fmt.Fprint(w, name)
	w.Write([]byte(" </title> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <h1> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(greeting))))
	w.Write([]byte(" </h1> <p>counter: "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counter))))
	w.Write([]byte("</p> <p>counter * 10 = "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(counterTimes10))))
	w.Write([]byte("</p> <button tx-onclick=\""))
	fmt.Fprint(w, addOne)
	w.Write([]byte("\" tx-swap=\""))
	fmt.Fprint(w, addOne_swap)
	w.Write([]byte("\">Add 1</button> <button tx-onclick=\"__1\"\" tx-swap=\""))
	fmt.Fprint(w, _1_swap)
	w.Write([]byte("\">Subtract 1</button> "))
	if counter%2 == 0 {
		w.Write([]byte("<p> counter is even </p> "))
	} else {
		w.Write([]byte("<p> counter is odd </p> "))

	}
	w.Write([]byte("<button tx-onclick=\""))
	fmt.Fprint(w, appendS)
	w.Write([]byte("?s="))
	if param, err := json.Marshal("str"); err != nil {
		log.Panic(err)
	} else {
		w.Write([]byte(url.QueryEscape(string(param))))
	}
	w.Write([]byte("\" tx-swap=\""))
	fmt.Fprint(w, appendS_swap)
	w.Write([]byte("\">append str</button> <p>"))
	w.Write([]byte(html.EscapeString(fmt.Sprint(str))))
	w.Write([]byte("</p> "))

	for i := 0; i < 10; i++ {
		w.Write([]byte("<p tx-key=\"i\"> "))
		w.Write([]byte(html.EscapeString(fmt.Sprint(i))))
		w.Write([]byte(" </p>"))

	}
	w.Write([]byte(" <a href=\"/second-page\">second page</a> </body></html>"))
}

type state_second_2D_page struct {
}

func render_second_2D_page(w io.Writer, key string, states map[string]string, newStates map[string]any) {
	w.Write([]byte("<html><head> <title> "))
	fmt.Fprint(w, 1+2)
	w.Write([]byte(" </title> <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <h1> "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(fmt.Sprintf("a + b = %d", 1+2)))))
	w.Write([]byte(" </h1> </body></html>"))
}

type state__7B_name_7D_ struct {
	S_name string `json:"name"`
}

func render__7B_name_7D_(w io.Writer, key string, states map[string]string, newStates map[string]any, name string) {
	w.Write([]byte("<!DOCTYPE html><html><head> <title> "))
	fmt.Fprint(w, name)
	w.Write([]byte(" </title>  <script id=\"tx-runtime\">"))
	w.Write([]byte(runtimeScript))
	w.Write([]byte("</script><script type=\"application/json\" id=\"tx-state\">TX_STATE_JSON</script></head> <body> <h1> Hello, "))
	w.Write([]byte(html.EscapeString(fmt.Sprint(name))))
	w.Write([]byte(" </h1> </body></html>"))
}

var txRoutes []TxRoute = []TxRoute{
	{
		Pattern: "GET /{$}",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			var name string = "tmplx"
			var greeting string = fmt.Sprintf("Hello ,%s!", name)
			var path string = tx_r.PathValue("index")
			var counter int
			var counterTimes10 int = counter * 10
			var str string = ""
			state := &state_{
				S_name:    name,
				S_path:    path,
				S_counter: counter,
				S_str:     str,
			}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render_(&buf, "tx_", map[string]string{}, newStates, name, greeting, path, counter, counterTimes10, str, "_addOne", "tx_", "_appendS", "tx_", "__1", "tx_")
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /tx/_addOne",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			query := tx_r.URL.Query()
			states := map[string]string{}
			for k, v := range query {
				if strings.HasPrefix(k, "tx_") {
					states[k] = v[0]
				}
			}
			newStates := map[string]any{}
			state := &state_{}
			json.Unmarshal([]byte(states["tx_"]), &state)
			name := state.S_name
			greeting := fmt.Sprintf("Hello ,%s!", name)
			path := state.S_path
			counter := state.S_counter
			counterTimes10 := counter * 10
			str := state.S_str
			counter++
			greeting = fmt.Sprintf("Hello ,%s!", name)
			counterTimes10 = counter * 10
			var buf bytes.Buffer
			render_(&buf, "tx_", states, newStates, name, greeting, path, counter, counterTimes10, str, "_addOne", "tx_", "_appendS", "tx_", "__1", "tx_")
			newStates["tx_"] = &state_{
				S_name:    name,
				S_path:    path,
				S_counter: counter,
				S_str:     str,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /tx/_appendS",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			query := tx_r.URL.Query()
			states := map[string]string{}
			for k, v := range query {
				if strings.HasPrefix(k, "tx_") {
					states[k] = v[0]
				}
			}
			newStates := map[string]any{}
			state := &state_{}
			json.Unmarshal([]byte(states["tx_"]), &state)
			name := state.S_name
			greeting := fmt.Sprintf("Hello ,%s!", name)
			path := state.S_path
			counter := state.S_counter
			counterTimes10 := counter * 10
			str := state.S_str
			var s string
			json.Unmarshal([]byte(query.Get("s")), &s)
			str += s
			greeting = fmt.Sprintf("Hello ,%s!", name)
			counterTimes10 = counter * 10
			var buf bytes.Buffer
			render_(&buf, "tx_", states, newStates, name, greeting, path, counter, counterTimes10, str, "_addOne", "tx_", "_appendS", "tx_", "__1", "tx_")
			newStates["tx_"] = &state_{
				S_name:    name,
				S_path:    path,
				S_counter: counter,
				S_str:     str,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: " /tx/__1",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			query := tx_r.URL.Query()
			states := map[string]string{}
			for k, v := range query {
				if strings.HasPrefix(k, "tx_") {
					states[k] = v[0]
				}
			}
			newStates := map[string]any{}
			state := &state_{}
			json.Unmarshal([]byte(states["tx_"]), &state)
			name := state.S_name
			greeting := fmt.Sprintf("Hello ,%s!", name)
			path := state.S_path
			counter := state.S_counter
			counterTimes10 := counter * 10
			str := state.S_str
			counter--
			greeting = fmt.Sprintf("Hello ,%s!", name)
			counterTimes10 = counter * 10
			var buf bytes.Buffer
			render_(&buf, "tx_", states, newStates, name, greeting, path, counter, counterTimes10, str, "_addOne", "tx_", "_appendS", "tx_", "__1", "tx_")
			newStates["tx_"] = &state_{
				S_name:    name,
				S_path:    path,
				S_counter: counter,
				S_str:     str,
			}
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /second-page",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			state := &state_second_2D_page{}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render_second_2D_page(&buf, "tx_", map[string]string{}, newStates)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
	{
		Pattern: "GET /{name}",
		Handler: func(w http.ResponseWriter, tx_r *http.Request) {
			var name string = tx_r.PathValue("name")
			state := &state__7B_name_7D_{
				S_name: name,
			}
			newStates := map[string]any{}
			newStates["tx_"] = state
			var buf bytes.Buffer
			render__7B_name_7D_(&buf, "tx_", map[string]string{}, newStates, name)
			stateBytes, _ := json.Marshal(newStates)
			w.Write(bytes.Replace(buf.Bytes(), []byte("TX_STATE_JSON"), stateBytes, 1))
		},
	},
}

func Routes() []TxRoute { return txRoutes }
