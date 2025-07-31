# Example Project

## Demo

[demo.webm](https://github.com/user-attachments/assets/b58e69de-e071-464b-a8c4-05a37063225d)

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
  
  <p>counter: { counter}</p>

  <button tx-onclick="addOne()">Add 1</button>
  <button tx-onclick="subOne()">Subtract 1</button>

  <p>counter * 10 = { counterTimes10 }</p>
</body>
</html>
```
