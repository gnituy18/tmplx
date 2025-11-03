# tmplx

tmplx is a compile-time framework using **Go** for building **state-driven** web apps. It allows you to build UIs with React-like reactivity purely in Go. Embed Go code in HTML to define states and event handlers. Go manages backend logic while HTML defines the UI, all in one file. This creates a seamless integration, eliminating the mental context switch between backend and frontend development.

- [Installing](#installing)
- [Quick Start](#quick-start)
- [Guide](#guide)
  - [.html](#html)
  - [Go expression interpolation](#go-expression-interpolation)
  - [<script type="text/tmplx">](#script-typetexttmplx)
  - [State](#state)
  - [Derived State](#derived-state)
  - [Event Handler](#event-handler)
    - [Arguments](#arguments)
    - [Inline Statements](#inline-statements)
  - [init()](#init)
  - [Control Flow](#control-flow)
    - [tx-if / tx-else-if / tx-else](#tx-if-tx-else-if-tx-else)
    - [tx-for](#tx-for)
  - [`<template>`](#template)
  - [`tx-value`](#tx-value)
  - [Components](#components)
  - [Morphing (WIP)](#morphing-wip)
  - [tmplx cmd](#tmplx-cmd)
    - [pages directory](#pages-directory)
    - [components directory](#components-directory)
    - [output](#output)

> [!WARNING]
> The project is in active development, with some of the features incomplete, and bugs or undefined behavior may occur.

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

### 2. Edit `pages/index.html`
```html
<script type="text/tmplx">
  var name string = "tmplx" // name is a state
  var greeting string = fmt.Sprintf("Hello, %s!", name) // greeting is a derived state

  var counter int = 0 // counter is a state
  var counterTimes10 int = counter * 10 // counterTimes10 is automatically changed if counter modified.

  // declare a event handler in Go!
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

  <!-- update counter by calling event handler -->
  <button tx-onclick="addOne()">Add 1</button>
</body>
</html>

```

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
Visit http://localhost:8080/ to see your web app in action.

## Guide

### `.html`
Your web app begins by creating an HTML file in the [pages](#pages-directory) directory. We chose HTML because—duh!—HTML is standardized and ensures backward compatibility across all browsers! This means you can build your app today, and even 10 years from now, tmplx will still be able to parse and handle it without issues.

Additionally, there's no need to invent a new file type to deliver all the features tmplx provides, as HTML already includes [extensibility](https://html.spec.whatwg.org/#extensibility) in its design.

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

### Go Expression Interpolation

Embed Go expressions in HTML using `{}` for dynamic content. You can only place Go expressions in **text nodes** or **attribute values**; other placements cause parsing errors.

For text nodes, the output is HTML-escaped; for attribute values, it is not escaped.

Expressions are wrapped in `fmt.Sprint()` in the [output Go file](#output).

You can add `tx-ignore` to disable Go expression interpolation for that specific node's attribute values and its text children, but not the element children.
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

A derived state is a variable whose value is computed from other states or derived states.

It is declared as a standard Go variable. If the right-hand side (RHS) of the declaration references existing states or derived states, it is considered a derived state.

Derived states follow similar rules to regular states, with some differences:

#### Rules:
1. **Use `var` keyword; no `:=`.**
1. **Must define a type.**
1. **Initialization is optional. If initializing, the number of variables on the left must match the expressions on the right.**
1. **You cannot update derived states directly in [event handlers](#event-handler), but referencing them is allowed.**

Derived states do not require the type to be JSON marshalable/unmarshalable because they exist only on the backend.

```html
<script type="text/tmplx">
  var num1 int = 100
  var num2 int = num1 * 2
</script>

...
<p> { num1 } * 2 = { num2 } </p>
```
```html
<script type="text/tmplx">
  var classes []string = []string{"c1", "c2", "c3"}
  var class string = strings.Join(classes, " ") // derived from state 'classes'
</script>

...
<p class="{class}"> ... </p>
```

### Event Handler
You could have guessed by now: An event handler is simply a regular Go function. Event handlers mutate states or perform actions in response to user events.

Rules:
1. **Triggered via attributes with the `tx-on` prefix.**
2. **No return values. You don't need them.**

You can bind multiple events to one element: `<div tx-onmouseleave="show = false" tx-onmouseenter="show = true">`
```html
<script type="text/tmplx">
  var counter int = 0

  func add1() {
    counter += 1
  }
</script>
...
<p>{ counter }</p>
<button tx-onclick="add1()">Add 1</button>
```

#### Arguments
Event handlers can accept arguments, following these rules:

1. **Argument names cannot match state or derived state names.**
2. **Argument types must be JSON marshalable and unmarshalable.**

```html
<script type="text/tmplx">
  var counter int = 0

  func addNum(num int) {
    counter += num
  }
</script>
...
<p>{ counter }</p>
<button tx-for="i := 0; i < 10 i++" tx-onclick="addNum(i)">
  Add { i }
</button>
```

#### Inline Statements
For simple actions, embed Go statements directly in `tx-on*` attributes to mutate states, avoiding the need for separate handler functions.

Use ';' to separate multiple statements.
```html
<script type="text/tmplx">
  var num int = 1
</script>
...
<p> { num } </p>
<button tx-onclick="num++;num++">Add 2</button>
```

### `init()`

You can declare a function named `init()`, similar to Go's `init()` function. It runs once when the page loads.

It won't be compiled into an HTTP request, so you cannot call it using `tx-on*` attributes.
```html
<script type="text/tmplx">
  var user User

  func init() {
    user = user.Get("user_id")
  }
</script>
```

Another use case is when you want to initialize a state from another state but don't want it to become a derived state.
```html
<script type="text/tmplx">
  var a int = 100
  var b int

  func init() {
    b = a * 2 // b is still a state
  }
</script>
...
```

### Control Flow

#### `tx-if`, `tx-else-if`, `tx-else`

You can use any valid expression in the value of `tx-if`, `tx-else-if` that fits Go's if statement condition. The `tx-else` attribute does not require any value. New variables created in the expression will also be accessible to the children of the node. It works just like Go's conditional statements.

```html
<script type="text/tmplx">
  var num int = 1
</script>
...
<button tx-onclick="num++">Add 1</button>
<p tx-if="counter % 2 == 1"> odd </p>
<p tx-else> even </p>
```

```html
<p tx-if="user, err := user.GetUser(); err != nil">
  <span tx-if="err == ErrNotFound"> User not found</span>
</p>
<p tx-else-if='user.Name == ""'> user.Name not set </p>
<p tx-else > Hi!, { user.Name } </p>
```

#### `tx-for`
You can put every thing that fit Go's `for` statement.
```html
...
<div tx-for="_, user := range users">
  { user.Id }: { user.Name }
</div>
```
```html
...
<div tx-for="i := 0; i < 10; i++">
  <div tx-for="j := 0; j < 10; j++">
    { i } * { j } = { i * j }
  </div>
</div>
```
### `<template>`
### `tx-value`
### components
### DOM morphing (WIP)
### `tmplx` cmd
#### `pages` directory
#### `components` directory
#### output
