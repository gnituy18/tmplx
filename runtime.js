document.addEventListener('DOMContentLoaded', function() {
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

    const res = await fetch("TX_HANDLER_PREFIX" + fun, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString() })
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
