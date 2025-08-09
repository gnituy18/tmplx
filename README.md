# tmplx

tmplx is a compile-time framework using Go for building state-driven web apps. It allows you to build UIs with React-like reactivity, purely in Go, without JavaScript. Embed Go code in HTML to define states and event handlers. Go manages backend logic while HTML defines the UI. This creates seamless integration, eliminating the mental context switch between backend and frontend development. This embraces hypermedia by letting the server drive UI updates via HTML responses. It compiles into Go handlers that update state, rerender UI sections, and return HTML snippets.

> [!WARNING]
> The project is in active development, with most of the features incomplete, and bugs or undefined behavior may occur. 

## Installing
```sh
go install github.com/gnituy18/tmplx@latest
```

## Quick Start
### 1. Set up a new project directory
```sh
mkdir proj && cd proj
mkdir pages
touch pages/index.html
touch main.go
go mod init proj
```
> [!NOTE]  
> The `pages` directory defines the app's routes based on file structure.
> 
> 1. `pages/index.html` → URL: `/`
> 1. `pages/this/is/a/path.html` → URL: `/this/is/a/path`

### 2. Edit `pages/index.html`
```html
<script type="text/tmplx">
  var name string = "tmplx" // name is declared as a state
  var greeting string = fmt.Sprintf("Hello ,%s!", name) // greeting is declared as a derived

  var counter int = 0
  var counterTimes10 int = counter * 10

  func addOne() {
    counter++
  }
</script>

<html>
<head>
  <title> { name } </title>
</head>
<body>
  <h1> { greeting } </h1>

  <p>counter: { counter }</p>
  <p>counter * 10 = { counterTimes10 }</p>

  <button tx-onclick="addOne()">Add 1</button>
</body>
</html>

```
> [!NOTE]  
> 1. `<script>` tags with `type="text/tmplx"` are parsed and transformed into output code. Everything inside is valid Go code; no new concepts to learn for building the app.
> 2. Variable declarations are **states**; changing them via `tx-on*` events triggers UI updates. States must be JSON stringifiable.
> 3. If a variable declaration's right-hand side references other states, it's a **derived** **state** that updates automatically when the referenced states change. Derived states cannot be modified; attempts cause compile errors.
> 4. Since it's compiled into Go code, you can perform server-side operations like: `var user = db.GetUser(userId)`

### 3. Edit `main.go`
```go
package main

import (
	"log"
	"net/http"

	"proj/tmplx"
)

func main() {
	for _, th := range tmplx.Handlers() {
		http.HandleFunc("GET "+th.Url, th.HandlerFunc)
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 4. Compile and Run the App
```sh
# From the proj directory
tmplx && go run .
```
Visit http://localhost:8080/ to see your web app in action

## Features
### Go Expression
```
...
<p> { "Hello, " + "Tmplx!" } </p>
<p> { 100 / 2 - 3 } is 47 </p>
```

### State
```html
<script type="text/tmplx">
  var str string = "Hello,"
  var m map[string]int = map[string]int{ "key": 100 }
</script>

...
<p> { str } World! </p>
<p> { m["key"] } </p>
```

### Derived
```html
<script type="text/tmplx">
  var num1 int = 100
  var num2 int = num1 * 2
</script>

...
<p> { num2 } is 200 </p>
```

### Actions (`tx-on*`)
```html
<script type="text/tmplx">
  var str string = "A"

  func appendA() {
    str = append(str, 'A')
  }
</script>

...
<p>{ str }</p>
<button tx-onclick="appendA()">Append A</button>
```
### inline statements
```html
<script type="text/tmplx">
  var num int = 1
</script>

...
<button tx-onclick="num++">Append A</button>
```
### `tx-if`
```html
<script type="text/tmplx">
  var num int = 1
</script>

...
<button tx-onclick="num++">Append A</button>
<p tx-if="counter % 2 == 1"> odd </p>
<p tx-else > even </p>
```
### `tx-for`
```html
...
<p tx-for="i := 0; i < 10; i++"> { i } </p>
```
### Components (Unimplemented)
### Styles and Classes (Unimplemented)
### ...


