# tmplx

tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by:

- keeping frontend and backend logic close together
- providing reactive UI updates driven by Go variables
- requiring zero new syntax

Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive.

```html
<script type="text/tmplx">
  var list []string
  var item string = ""
  
  func add() {
    list = append(list, item)
    item = ""
  }
  
  func remove(i int) {
    list = append(list[0:i], list[i+1:]...)
  }
</script>

<label><input type="text" tx-value="item"></label>
<button tx-onclick="add()">Add</button>
<ol>
  <li 
    tx-for="i, l := range list"
    tx-key="l"
    tx-onclick="remove(i)">
    { l }
  </li>
</ol>
```

> [!WARNING]
> The project is in active development, with some of the features incomplete, and bugs or undefined behavior may occur.

## Links

Website & Demos https://tmplx.org

Docs https://tmplx.org/docs

