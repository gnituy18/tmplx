# tmplx

tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by:

- keeping frontend and backend logic close together
- providing reactive UI updates driven by Go variables
- requiring zero new syntax

Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive.

```html
<script type="text/tmplx">
  var list []string

  func add(item string) {
    list = append(list, item)
  }

  func remove(i int) {
    list = append(list[0:i], list[i+1:]...)
  }
</script>

<form tx-action="add">
  <label><input name="item" type="text" required></label>
  <button type="submit">Add</button>
</form>
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

Roadmap https://tmplx.org/roadmap
