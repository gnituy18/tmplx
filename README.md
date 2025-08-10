# tmplx

tmplx is a compile-time framework using Go for building state-driven web apps. It allows you to build UIs with React-like reactivity purely in Go. Embed Go code in HTML to define states and event handlers. Go manages backend logic while HTML defines the UI, all in one file. This creates a seamless integration, eliminating the mental context switch between backend and frontend development.

> [!WARNING]
> The project is in active development, with most of the features incomplete, and bugs or undefined behavior may occur. 

## Installing
```sh
go install github.com/gnituy18/tmplx@latest
```

## Quick Start
### 1. Set up a new project directory
```sh
mkdir proj
cd proj

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
  var name string = "tmplx"
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

## Guide

### `.html`
Your web app begins by creating an HTML file in the [pages](#pages-directory) directory. We chose HTML because—duh!—HTML is standardized and ensures backward compatibility across all browsers! This means you can build your app today, and even 10 years from now, tmplx will still be able to parse and handle it without issues.

Additionally, there's no need to invent a new file type to deliver all the features tmplx provides, as HTML already includes [built-in customizability](https://html.spec.whatwg.org/#extensibility) in its design.

You can access and modify every part of the HTML file, such as:
- adding attributes to the `<html>` tag
- customizing the `<body>` style
- adding a comment

This is often not obvious in modern frameworks. You can do whatever you want as long as it's valid HTML.
```html
<!-- /pages/index.html -->
<!DOCTYPE html>

<html lang="en">
  <head>
    <title> ... </title>
    ...
  </head>
  <body style="...">
  ...
  </body>
</html>
```

### Go expression interpolation
Embed Go expressions in HTML using `{ }` for dynamic content. You can only place Go expressions as `text nodes` or `attribute values`; other placements cause parsing errors.

For text nodes, output is HTML-escaped; for attribute values, it is not escaped.

Expressions are wrapped in `fmt.Sprint()` in the output file.

```html
<!-- /pages/index.html -->
<p class='{ strings.Join([]string{"c1", "c2"}, " ") }'>
 Hello, { user.GetNameById("id") }!
</p>

<!-- output -->
<p class="c1 c2">
 Hello, tmplx!
</p>
```

### `<script type="text/tmplx">`
tmplx extends HTML by embedding Go code within `<script>` tags. Set `type="text/tmplx"` to differentiate it from JavaScript or other languages.

The `<script>` contains valid Go code. tmplx uses a subset of Go syntax for declarative UI, including [state](#state), [derived state](#derived-state), and [event handler](#event-handler).
```diff
<!-- /pages/index.html -->
<!DOCTYPE html>

+ <script type="text/tmplx">
+ ...
+ </script>

<html lang="en">
  <head>
    <title> ... </title>
    ...
  </head>
  <body style="...">
  ...
  </body>
</html>
```

### State
State is the core of declarative UI development. It means that whenever a state changes, other UI parts react automatically.

State declaration is simply Go's variable declaration with a few rules. Since tmplx is a compiler, no special keyword is needed. Nothing new to learn.

#### Rules:
1. **Use `var` keyword; no `:=`.**
1. **Must define a type; it must be JSON marshalable and unmarshalable.**
1. **Initialization is optional. If initializing, the number of variables on the left must match expressions on the right.**

##### ❌ invalid state declarations
```html
...
<script type="text/tmplx">
  // no type
  var str = ""

  // no :=
  num := 1

  // f, w are not JSON marshalable and unmarshalable.
  var f func(int) = func(i int) {...} 
  var w io.Writer

  // the number of variables on the left must match expressions on the right.
  var a, b int = f() 
</script>
...
```
##### ✅ valid state declarations 
```html
<script type="text/tmplx">
  var name string = user.GetNameById("id")
  var m map[string]int = map[string]int{ "key": 100 }
</script>

...
<p> Hi, { name }! </p>
<p> { m["key"] } </p>
```

### Derived State
```html
<script type="text/tmplx">
  var num1 int = 100
  var num2 int = num1 * 2
</script>

...
<p> { num2 } is 200 </p>
```

### Event Handler
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

### `tmplx` cmd
The tmplx compiler scans `pages` and `components` directories for `.html` files, generates a `handler.go` file to include in your project.
```
```
you can custimize the pages's path or components' or the output path using -pages, -components -output flags

#### `pages` directory
#### `components` directory
#### `handler.go`


Your web app lives under a go project. all you html file should live under the pages folder. it support subfolder.
tmplx cli compiles multiple html file into a go package. the package exports a list of the handler and url. you can then import the package serve the handler to activate the app.

### Components (Unimplemented)
### Styles and Classes (Unimplemented)
### ...


