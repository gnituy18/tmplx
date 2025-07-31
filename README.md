# tmplx

tmplx is a compiler that transforms hybrid HTML/Go templates into a fully dynamic hypermedia web application.
Check out [Example Project](https://github.com/gnituy18/tmplx/tree/main/example_project) to see what it can do.

tmplx tries bring state-driven UI development to the hypermedia world.
Paired with Go, my language of choice, it offers a seamless experience for building web apps.

HTMX is an powerful tool that enhances the current state of HTML. However, it lacks a robust framework for our minds to manage the complexity of larger web apps.

> [!WARNING]
> The project is in active development, with most of the features incomplete, and bugs or undefined behavior may occur. 

## Installing
Right now, you have to compile the compiler yourself.

```sh
git clone https://github.com/gnituy18/tmplx.git
cd tmplx
go build
```
After compiling tmplx, move the executable to a directory in your PATH (e.g., /usr/local/bin).
```sh
mv tmplx /usr/local/bin
```

## Quick Start
Create a new project
```sh
mkdir my_project
cd my_project

mkdir pages
touch pages/index.tmplx # or index.html
```
Edit `pages/index.tmplx`
```html
<script type="text/tmplx">
  var title string = "My project"
  var h1Text string = "Hello, Tmplx!"
  var counter int = 0

  func addOne() {
    counter++
  }
</script>

<html>
<head>
  <title> { title } </title>
</head>

<body>
  <h1> { h1Text } </h1>
  <p>counter: { counter}</p>
  <button tx-onclick="addOne()">Add 1</button>
</body>
</html>
```
compile
```sh
# ./my_project
tmplx
```
you will then see a new package created: `tmplx/handler.go`.
To run the app, create a `main.go`
```go
package main

import (
	"log"
	"net/http"

	"my_project/tmplx"
)

func main() {
	for _, th := range tmplx.Handlers() {
		http.HandleFunc("GET "+th.Url, th.HandlerFunc)
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
You can run your server code and go to http://localhost:8080/.
Now, you have a hypermedia app.
```
go run .
```

## Todos
