# tmplx

tmplx is a compiler that transforms a hybrid HTML/Go template language into a fully dynamic hypermedia web application.

## Example

```html
<script type="text/tmplx">
  var title string = "Tmplx!"
  var h1 string = "Hello, Tmplx!"

  var counter int = 0
  var counterTimes10 int = counter*10

  func addOne() {
    counter++
  }

  func subOne() {
    counter--
  }
</script>

<html>
<head>
  <title> { title } </title>
</head>

<body>
  <h1> { h1 } </h1>
  
  <h2> Counter <h2>
  <p>counter: { counter}</p>

  <button tx-onclick="addOne()">Add 1</button>
  <button tx-onclick="subOne()">Subtract 1</button>

  <h2> Derived <h2>
  <p>counter * 10 = { counterTimes10 }</p>
</body>
</html>
```
