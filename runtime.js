document.addEventListener('DOMContentLoaded', function() {
  let txSaved = this.getElementById("tx-saved")
  let state = JSON.parse(txSaved.innerHTML)
  let page = txSaved.getAttribute("data-tx-page")
  let tasks = [];
  let isProcessing = false;

  const findComment = (root, text) => {
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_COMMENT)
    while (walker.nextNode()) {
      if (walker.currentNode.nodeValue === text) return walker.currentNode
    }
  }

  const morph = (a, b) => {
    if (a.nodeName !== b.nodeName) {
      a.replaceWith(b.cloneNode(true))
      return
    }
    if (a.nodeType !== Node.ELEMENT_NODE) {
      if (a.nodeValue !== b.nodeValue) a.nodeValue = b.nodeValue
      return
    }
    for (const at of [...a.attributes]) {
      if (!b.hasAttribute(at.name)) a.removeAttribute(at.name)
    }
    for (const at of b.attributes) {
      if (a.getAttribute(at.name) !== at.value) a.setAttribute(at.name, at.value)
    }
    let an = a.firstChild, bn = b.firstChild
    while (an && bn) {
      const an2 = an.nextSibling, bn2 = bn.nextSibling
      morph(an, bn)
      an = an2
      bn = bn2
    }
    while (an) {
      const n = an.nextSibling
      an.remove()
      an = n
    }
    while (bn) {
      a.appendChild(bn.cloneNode(true))
      bn = bn.nextSibling
    }
  }

  const send = async (cn, fun, target, params) => {
    if (!target) return

    const trigger = cn.getAttribute("data-tx-trigger")
    if (trigger !== null) {
      params.append("trigger", trigger)
    }
    params.append("target", target === 'page' ? page : target)

    for (let key in state) {
      params.append(key, JSON.stringify(state[key]))
    }

    const res = await fetch("TX_HANDLER_PREFIX" + fun, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString() })
    const html = await res.text()

    if (target === 'page') {
      const doc = new DOMParser().parseFromString(html, 'text/html')
      const txState = doc.getElementById('tx-saved')
      if (txState) state = { ...state, ...JSON.parse(txState.textContent) }
      morph(document.documentElement, doc.documentElement)
      return
    }

    const comp = document.createElement('body')
    comp.innerHTML = html
    const txState = comp.querySelector("#tx-saved")
    if (!txState) return
    state = { ...state, ...JSON.parse(txState.textContent) }

    const respStart = findComment(comp, 'tx:' + target)
    const respEnd = findComment(comp, 'tx:' + target + '_e')
    const docStart = findComment(document.documentElement, 'tx:' + target)
    const docEnd = findComment(document.documentElement, 'tx:' + target + '_e')
    if (!respStart || !respEnd || !docStart || !docEnd) return

    const respRange = document.createRange()
    respRange.setStartBefore(respStart)
    respRange.setEndAfter(respEnd)

    const range = document.createRange()
    range.setStartBefore(docStart)
    range.setEndAfter(docEnd)
    range.deleteContents()
    range.insertNode(respRange.extractContents())
  }

  const init = (cn) => {
    const trigger = cn.getAttribute('data-tx-trigger')
    const target = cn.getAttribute('data-tx-target')
    let base
    if (trigger === 'page') {
      base = encodeURIComponent(page)
    } else if (trigger !== null) {
      const seg = trigger.slice(Math.max(trigger.lastIndexOf(':'), trigger.lastIndexOf('@')) + 1)
      base = seg.slice(0, seg.lastIndexOf('-'))
    }
    for (let attr of cn.attributes) {
      if (attr.name.startsWith('data-tx-') && attr.name.endsWith('-on')) {
        const id = attr.name.slice(8, -3)
        const eventName = attr.value
        const fun = base + '/' + id
        const argPfx = 'data-tx-' + id + '-arg-'
        cn.addEventListener(eventName, (e) => {
          const p = new URLSearchParams()
          if (eventName === 'input' || eventName === 'change') {
            p.append('tx_ev_target_value', JSON.stringify(e.target.value))
          }
          for (let a of cn.attributes) {
            if (a.name.startsWith(argPfx)) {
              p.append(a.name.slice(argPfx.length), a.value)
            }
          }
          tasks.push(() => send(cn, fun, target, p))
          processQueue()
        })
      } else if (attr.name === 'data-tx-action') {
        const fun = attr.value
        const target = cn.getAttribute('data-tx-target')
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
          tasks.push(() => send(cn, fun, target, params))
          processQueue()
        })
      }
    }
  }

  async function processQueue() {
    if (isProcessing) return;
    isProcessing = true;
    try {
      while (tasks.length > 0) {
        const task = tasks.shift();
        await task();
      }
    } finally {
      isProcessing = false;
    }
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
          if (attr.name.startsWith('data-tx-')) {
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
