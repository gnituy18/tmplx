# tmplx

tmplx is a compiler that transforms a hybrid HTML/Go template language into a fully dynamic hypermedia web application.

## Example

```html
<script type="text/tmplx">
  appName := "Todo"

  input := ""

  list := []string{"init"}

  add := func() {
    list = append(list, input)
  }

  delete := func(i int) {
    list = slices.Delete(list, i, i+1)
  }
</script>

<head>
</head>

<body>
  <h1> { appName } </h1>

  <input type="text" tx-value="input">
  <button tx-onclick="add">Add</button>

  <ul>
    <li tx-for="i, item := range list">
      <button tx-onclick="delete(i)">delete</button>
    </li>
  </ul>

</body>
```

## State
1. Must specifiy a type
1. Must have it's own init

## Derived
