package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type tx_H_addn struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_H_addn(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_addn {
	tx_comp := &tx_H_addn{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_counter = 0
	}
	return tx_comp
}

func (tx_comp *tx_H_addn) addNum(num int) {
	tx_comp.V_counter += num
}

func (tx_comp *tx_H_addn) tx_eh1(i int) {
	tx_comp.addNum(i)
}

func (tx_comp *tx_H_addn) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var i int
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("i")), &i)
			tx_comp.tx_eh1(i)
		}
	}

	for i := 0; i < 10; i++ {
		_ = i

	}
}

func (tx_comp *tx_H_addn) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w.WriteString("</p> ")

	for i := 0; i < 10; i++ {
		_ = i
		tx_w.WriteString("<button data-tx-trigger=\"")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("\" data-tx-target=\"")
		fmt.Fprint(tx_w, tx_comp.tx_target)
		tx_w.WriteString("\" data-tx-eh1-on=\"click\" data-tx-eh1-arg-i=\"")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(func() string { tx_b, _ := json.Marshal(i); return string(tx_b) }())))
		tx_w.WriteString("\"> +")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(i)))
		tx_w.WriteString(" </button>")

	}
	tx_w.WriteString(" <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_cond struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_num int `json:"num"`
}

func tx_new_tx_H_cond(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_cond {
	tx_comp := &tx_H_cond{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_cond) tx_eh1() {
	tx_comp.V_num++
}

func (tx_comp *tx_H_cond) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	if tx_comp.V_num%3 == 0 {
	} else if tx_comp.V_num%3 == 1 {
	} else {

	}
}

func (tx_comp *tx_H_cond) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">change</button> <div> ")
	if tx_comp.V_num%3 == 0 {
		tx_w.WriteString("<p style=\"background: red; color: white\">red</p> ")
	} else if tx_comp.V_num%3 == 1 {
		tx_w.WriteString("<p style=\"background: blue; color: white\">blue</p> ")
	} else {
		tx_w.WriteString("<p style=\"background: green; color: white\">green</p> ")

	}
	tx_w.WriteString("</div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_counter struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_H_counter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_counter {
	tx_comp := &tx_H_counter{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_counter) tx_eh1() {
	tx_comp.V_counter--
}

func (tx_comp *tx_H_counter) tx_eh2() {
	tx_comp.V_counter++
}

func (tx_comp *tx_H_counter) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		case "eh2":
			tx_comp.tx_eh2()
		}
	}
}

func (tx_comp *tx_H_counter) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">-</button> <span> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w.WriteString(" </span> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh2-on=\"click\">+</button> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_current_H_time struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_t string `json:"t"`
}

func tx_new_tx_H_current_H_time(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_current_H_time {
	tx_comp := &tx_H_current_H_time{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_t = fmt.Sprint(time.Now().Format(time.RFC3339))
	}
	return tx_comp
}

func (tx_comp *tx_H_current_H_time) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_t)))
	tx_w.WriteString("</p> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_double struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_val int `json:"val"`
}

func tx_new_tx_H_double(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_double {
	tx_comp := &tx_H_double{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_val = 1
	}
	return tx_comp
}

func (tx_comp *tx_H_double) tx_eh1() {
	tx_comp.V_val *= 2
}

func (tx_comp *tx_H_double) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_double) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_val)))
	tx_w.WriteString("</p> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">double it!</button> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_double_H_state struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_b int `json:"b"`
	V_a int `json:"a"`
}

func tx_new_tx_H_double_H_state(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_double_H_state {
	tx_comp := &tx_H_double_H_state{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_a = 1
		tx_comp.V_b = tx_comp.V_a * 2
	}
	return tx_comp
}

func (tx_comp *tx_H_double_H_state) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <div> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_b * tx_comp.V_a)))
	tx_w.WriteString(" </div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_example_H_wrapper struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_H_example_H_wrapper(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_example_H_wrapper {
	tx_comp := &tx_H_example_H_wrapper{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_example_H_wrapper) tx_render(tx_w *bytes.Buffer, tx_id string, tx_render_fill_ func()) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--><div style=\"\n    margin-top: 0.5rem;\n    margin-bottom: 0.5rem;\n    padding: 2rem;\n    display: flex;\n    justify-content: center;\n    align-items: center;\n    border: solid SlateGray;\n    border-radius: 0.25rem;\n  \"> <div> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	}
	tx_w.WriteString(" </div> </div> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_greeting struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_greeting string `json:"greeting"`
}

func tx_new_tx_H_greeting(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_greeting {
	tx_comp := &tx_H_greeting{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_greeting) greet(name string) {
	tx_comp.V_greeting = "Hello, " + name
}

func (tx_comp *tx_H_greeting) tx_eh1(name string) {
	tx_comp.greet(name)
}

func (tx_comp *tx_H_greeting) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var name string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("name")), &name)
			tx_comp.tx_eh1(name)
		}
	}
	if tx_comp.V_greeting != "" {

	}
}

func (tx_comp *tx_H_greeting) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <form data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-action=\"tx-greeting/eh1\"> <input name=\"name\" type=\"text\" required=\"\"/> <button type=\"submit\">Greet</button> </form> ")
	if tx_comp.V_greeting != "" {
		tx_w.WriteString("<p>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_greeting)))
		tx_w.WriteString("</p> ")

	}
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_todo struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_list []string `json:"list"`
}

func tx_new_tx_H_todo(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_todo {
	tx_comp := &tx_H_todo{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_todo) add(item string) {
	tx_comp.V_list = append(tx_comp.V_list, item)
}

func (tx_comp *tx_H_todo) remove(i int) {
	tx_comp.V_list = append(tx_comp.V_list[0:i], tx_comp.V_list[i+1:]...)
}

func (tx_comp *tx_H_todo) tx_eh1(item string) {
	tx_comp.add(item)
}

func (tx_comp *tx_H_todo) tx_eh2(i int) {
	tx_comp.remove(i)
}

func (tx_comp *tx_H_todo) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var item string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("item")), &item)
			tx_comp.tx_eh1(item)
		case "eh2":
			var i int
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("i")), &i)
			tx_comp.tx_eh2(i)
		}
	}

	for i, l := range tx_comp.V_list {
		_ = i
		_ = l

	}
}

func (tx_comp *tx_H_todo) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <form data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-action=\"tx-todo/eh1\"> <label><input name=\"item\" type=\"text\" required=\"\"/></label> <button type=\"submit\">Add</button> </form> <ol> ")

	for i, l := range tx_comp.V_list {
		_ = i
		_ = l
		tx_w.WriteString("<li data-tx-trigger=\"")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("\" data-tx-target=\"")
		fmt.Fprint(tx_w, tx_comp.tx_target)
		tx_w.WriteString("\" data-tx-eh2-on=\"click\" data-tx-eh2-arg-i=\"")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(func() string { tx_b, _ := json.Marshal(i); return string(tx_b) }())))
		tx_w.WriteString("\"> ")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(l)))
		tx_w.WriteString(" </li>")

	}
	tx_w.WriteString(" </ol> <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_H_triangle struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_H_triangle(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_triangle {
	tx_comp := &tx_H_triangle{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_counter = 5
	}
	return tx_comp
}

func (tx_comp *tx_H_triangle) tx_eh1() {
	tx_comp.V_counter++
}

func (tx_comp *tx_H_triangle) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}

	for h := 0; h < tx_comp.V_counter; h++ {
		_ = h

		for s := 0; s < tx_comp.V_counter-h-1; s++ {
			_ = s

		}

		for i := 0; i < h*2+1; i++ {
			_ = i

		}

	}
}

func (tx_comp *tx_H_triangle) tx_render(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString("<!--tx:")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("--> <div> <span> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w.WriteString(" </span> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">+</button> </div> ")

	for h := 0; h < tx_comp.V_counter; h++ {
		_ = h
		tx_w.WriteString("<div> ")

		for s := 0; s < tx_comp.V_counter-h-1; s++ {
			_ = s
			tx_w.WriteString("<span>_</span>")

		}
		tx_w.WriteString(" ")

		for i := 0; i < h*2+1; i++ {
			_ = i
			tx_w.WriteString("<span>*</span>")

		}
		tx_w.WriteString(" </div>")

	}
	tx_w.WriteString(" <!--tx:")
	fmt.Fprint(tx_w, tx_id+"_e")
	tx_w.WriteString("-->")
}

type tx_S_docs struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_docs(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_docs {
	tx_comp := &tx_S_docs{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_S_docs) tx_compute() {
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-todo-1"
			tx_child := tx_new_tx_H_todo(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-addn-1"
			tx_child := tx_new_tx_H_addn(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-double-1"
			tx_child := tx_new_tx_H_double(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-4"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-current-time-1"
			tx_child := tx_new_tx_H_current_H_time(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-5"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-cond-1"
			tx_child := tx_new_tx_H_cond(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-6"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-triangle-1"
			tx_child := tx_new_tx_H_triangle(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-7"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-greeting-1"
			tx_child := tx_new_tx_H_greeting(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
}

func (tx_comp *tx_S_docs) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Docs | tmplx</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/docs\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <nav> <h2>tmplx Docs</h2> <ul> <li><a href=\"#introduction\">Introduction</a></li> <li><a href=\"#installing\">Installing</a></li> <li><a href=\"#quick-start\">Quick Start</a></li> <li><a href=\"#pages-and-routing\">Pages and Routing</a></li> <li> <a href=\"#tmplx-script\">tmplx Script</a> <ul> <li><a href=\"#reserved-names\">Reserved Names</a></li> </ul> </li> <li> <a href=\"#expression-interpolation\">Expression Interpolation</a> </li> <li><a href=\"#state\">State</a></li> <li><a href=\"#derived\">Derived</a></li> <li><a href=\"#event-handler\">Event Handler</a></li> <li><a href=\"#init\">init()</a></li> <li><a href=\"#path-parameter\">Path Parameter</a></li> <li> <a href=\"#control-flow\">Control Flow</a> <ul> <li><a href=\"#conditionals\">Conditionals</a></li> <li><a href=\"#loops\">Loops</a></li> </ul> </li> <li><a href=\"#template\">&lt;template&gt;</a></li> <li><a href=\"#forms\">Forms</a></li> <li> <a href=\"#component\">Component</a> <ul> <li> <a href=\"#props\">Props</a> <ul> <li><a href=\"#callback-props\">Callback Props</a></li> </ul> </li> <li><a href=\"#slot\">&lt;slot&gt;</a></li> </ul> </li> <li><a href=\"#cli\">CLI</a></li> <li> Dev Tools <ul> <li><a href=\"#syntax-highlight\">Syntax Highlight</a></li> </ul> </li> </ul> </nav> <main> <h2 id=\"introduction\">Introduction</h2> <p> tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by: </p> <ol> <li> <strong>keeping frontend and backend logic close together</strong> </li> <li> <strong>providing reactive UI updates driven by Go variables</strong> </li> <li><strong>requiring zero new syntax</strong></li> </ol> <p> Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive. </p> ")
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_1_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n\n  func add(item string) {\n    list = append(list, item)\n  }\n\n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;add&#34;&gt;\n  &lt;label&gt;&lt;input name=&#34;item&#34; type=&#34;text&#34; required&gt;&lt;/label&gt;\n  &lt;button type=&#34;submit&#34;&gt;Add&lt;/button&gt;\n&lt;/form&gt;\n&lt;ol&gt;\n  &lt;li\n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code></pre> <p> You start by creating an HTML file. It can be a page or a reusable component, depending on where you place it. </p> <p> You use the <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> tag to embed Go code and make the page or component dynamic. tmplx uses a subset of Go syntax to provide reactive features like <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, and <a href=\"#event-handler\">event handler</a>. At the same time, because the script is valid Go, you can <strong>implement backend logic</strong>—such as database queries—directly in the template. </p> <p> tmplx compiles the HTML templates and embedded Go code into Go functions that render the HTML on the server and generate HTTP handlers for interactive events. On each interaction, the current state is sent to the server, which computes updates and returns both new HTML and the updated state. The result is server-rendered pages with lightweight client-side swapping (similar to <a href=\"https://htmx.org/\">htmx</a>). The interactivity plumbing is handled automatically by the tmplx compiler and runtime—you just implement the features. </p> <p> Most modern web applications separate the frontend and backend into different languages and teams. tmplx eliminates this split by letting you build the entire interactive application in a single language—Go. With this approach, the mental effort needed to track how data flows from the source to the UI is reduced to a minimum. The fewer transformations you perform on your data, the fewer bugs you introduce. </p> <h2 id=\"installing\">Installing</h2> <p>tmplx requires Go 1.25 or later.</p> <pre><code>$ go install github.com/gnituy18/tmplx@latest</code></pre> <p> This adds tmplx to your Go bin directory (usually $GOPATH/bin or $HOME/go/bin). Make sure that directory is in your PATH. </p> <p>After installation, verify it works:</p> <pre><code>$ tmplx --help</code></pre> <h2 id=\"quick-start\">Quick Start</h2> <p>Get a tmplx app running in minutes.</p> <ol> <li> <p><strong>Create a project</strong></p> <pre><code>$ mkdir hello-tmplx\n$ cd hello-tmplx\n$ go mod init hello-tmplx\n$ mkdir pages</code></pre> </li> <li> <p><strong>Add your first page (pages/index.html)</strong></p> <pre><code>&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n&lt;head&gt;\n  &lt;meta charset=&#34;UTF-8&#34;&gt;\n  &lt;title&gt;Hello tmplx&lt;/title&gt;\n&lt;/head&gt;\n&lt;body&gt;\n  &lt;script type=&#34;text/tmplx&#34;&gt;\n    var count int\n  &lt;/script&gt;\n\n  &lt;h1&gt;Counter&lt;/h1&gt;\n\n  &lt;button tx-onclick=&#34;count--&#34;&gt;-&lt;/button&gt;\n  &lt;span&gt;{ count }&lt;/span&gt;\n  &lt;button tx-onclick=&#34;count++&#34;&gt;+&lt;/button&gt;\n&lt;/body&gt;\n&lt;/html&gt;</code></pre> </li> <li> <p><strong>Generate the Go code</strong></p> <pre><code>$ tmplx</code></pre> </li> <li> <p><strong>Create main.go to serve the app</strong></p> <pre><code>package main\n\nimport (\n\t&#34;log&#34;\n\t&#34;net/http&#34;\n)\n\nfunc main() {\n\tfor _, route := range Routes() {\n\t\thttp.Handle(route.Pattern, route.Handler)\n\t}\n\n\tlog.Fatal(http.ListenAndServe(&#34;:8080&#34;, nil))\n}</code></pre> </li> <li> <p><strong>Run the server</strong></p> <pre><code>$ go run .\n&gt; Listening on :8080</code></pre> </li> </ol> <p> That&#39;s it! Open <a href=\"http://localhost:8080\">http://localhost:8080</a> and you now have a working interactive counter. </p> <h2 id=\"pages-and-routing\">Pages and Routing</h2> <p> A <strong>page</strong> is a standalone HTML file that has its own URL in your web app. </p> <p> All pages are placed in the <strong>pages</strong> directory. Default pages location is <code>./pages</code>. Change it with the <code>-pages-dir</code> flag: </p> <pre><code>$ tmplx -pages-dir=&#34;/some/other/location&#34;</code></pre> <p> tmplx uses <strong>filesystem-based routing</strong>. The route for a page is the relative path of the HTML file inside the <strong>pages</strong> directory, without the <code>.html</code> extension. For example: </p> <ul> <li><code>pages/index.html</code> → <code>/</code></li> <li><code>pages/about.html</code> → <code>/about</code></li> <li> <code>pages/admin/dashboard.html</code> → <code>/admin/dashboard</code> </li> </ul> <p> When the file is named <code>index.html</code>, the <code>index</code> part is omitted from the route (it serves the directory path). To get a route like <code>/index</code>, place <code>index.html</code> in a subdirectory named <code>index</code>. </p> <ul> <li><code>pages/index/index.html</code> → <code>/index</code></li> </ul> <p> Multiple file paths can map to the same route. Choose the style you prefer. Duplicate routes cause compilation failure. </p> <ul> <li><code>pages/login/index.html</code> → <code>/login</code></li> <li><code>pages/login.html</code> → <code>/login</code></li> </ul> <p> To add URL parameters (path wildcards), use curly braces  in directory or file names inside the pages directory. The name inside  must be a valid Go identifier. </p> <ul> <li> <code>pages/user/{user_id}.html</code> → <code>/user/{user_id}</code> </li> <li> <code>pages/blog/{year}/{slug}.html</code> → <code>/blog/{year}/{slug}</code> </li> </ul> <p> These patterns are compatible with Go&#39;s <code>net/http.ServeMux</code> (Go 1.22+). The parameter values are available in page initialisation through <code><a href=\"#path-parameter\">tx:path</a></code> comments. </p> <p> tmplx compiles all pages into a single Go file you can import into your Go project. The pages directory can be outside your project, but keeping it inside is recommended. </p> <h2 id=\"tmplx-script\">tmplx Script</h2> <p> <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> is a special tag that you can add to your page or component to declare <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, <a href=\"#event-handler\">event handler</a>, and the special <a href=\"#init\">init()</a> function to control your UI or add backend logic. </p> <p> Each page or component file can have exactly <strong>one</strong> tmplx script. Multiple scripts cause a compilation error. </p> <p> In pages, place it anywhere inside <code>&lt;head&gt;</code> or <code>&lt;body&gt;</code>. </p> <pre><code>&lt;!DOCTYPE html&gt;\n&lt;html lang=&#34;en&#34;&gt;\n  &lt;head&gt;\n    ...\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // Go code here\n    &lt;/script&gt;\n    ...\n  &lt;/head&gt;\n  &lt;body&gt;\n    ...\n  &lt;/body&gt;\n&lt;/html&gt;</code>\n      </pre> <p>In components, place it at the root level.</p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  // Go code here\n&lt;/script&gt;\n...\n...</code></pre> <h3 id=\"reserved-names\">Reserved Names</h3> <p> The compiler reserves two naming patterns for its own use: </p> <ul> <li> Identifiers (variables, function names, parameter names) declared in the tmplx script cannot start with <code>tx_</code>. </li> <li> HTML attributes starting with <code>tx-</code> are reserved for tmplx directives (<code>tx-if</code>, <code>tx-for</code>, <code>tx-on*</code>, <code>tx-action</code>, ...). Do not introduce your own <code>tx-</code> attributes. </li> </ul> <h2 id=\"expression-interpolation\">Expression Interpolation</h2> <p> Use curly braces <code>{}</code> to insert <a href=\"https://go.dev/ref/spec#Expressions\">Go expressions</a> into HTML. Expressions are allowed only in: </p> <ul> <li><strong>text nodes</strong></li> <li><strong>attribute values</strong></li> </ul> <p>Placing expressions anywhere else causes a parsing error.</p> <p>\n        tmplx converts expression results to strings using\n        <code><a href=\"https://pkg.go.dev/fmt#Sprint\">fmt.Sprint</a></code>. The output is <strong>HTML-escaped</strong> in both\n        <strong>text nodes</strong> and <strong>attribute values</strong> to\n        prevent cross-site scripting (XSS)—an interpolated value cannot\n        inject markup or break out of its attribute.\n      </p> <p> Expressions run on the server every time the page loads or a component re-renders after an event. Avoid side effects in expressions, such as database queries or heavy computations, because they execute on every render. </p> <pre><code class=\"language-html\">&lt;p class=&#39;{ strings.Join([]string{&#34;c1&#34;, &#34;c2&#34;}, &#34; &#34;) }&#39;&gt;\n Hello, { user.GetNameById(0) }!\n&lt;/p&gt;</code>\n      </pre> <pre><code class=\"language-html\">&lt;p class=&#34;c1 c2&#34;&gt;\n Hello, tmplx!\n&lt;/p&gt;</code></pre> <p>\n        Add the <code>tx-ignore</code> attribute to an element to disable\n        expression interpolation in that element&#39;s attributes and its direct\n        text children. Descendant elements are still processed normally.\n      </p> <pre><code class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;{ &#34;not&#34; + &#34; ignored&#34; }&lt;/span&gt;\n&lt;/p&gt;</code>\n      </pre> <pre><code class=\"language-html\">&lt;p tx-ignore&gt;\n  { &#34;ignored&#34; }\n  &lt;span&gt;not ignored&lt;/span&gt;\n&lt;/p&gt;</code></pre> <h2 id=\"state\">State</h2> <p> <strong>State</strong> is the mutable data that describes a component&#39;s current condition. </p> <p> Declaring state works like declaring variables in Go&#39;s package scope. If you provide no initial value, the state starts with the zero value for its type. </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string\n&lt;/script&gt;</code></pre> <p>To set an initial value, use the <code>=</code> operator.</p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\nvar name string = &#34;tmplx&#34;\n&lt;/script&gt;</code></pre> <p>Although the syntax follows valid Go code, these rules apply:</p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong> The type must be JSON-compatible. It may be either declared explicitly or inferred from the initializer. </strong> </li> </ol> <p> The 1st rule is enforced by the compiler. JSON-compatibility is not checked at compile time (for now) and will cause a runtime error if violated. </p> <h3>Some invalid state declarations:</h3> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n// ❌ Cannot use the := short declaration (Go does not allow it at package scope)\nnum := 1\n\n// ❌ Type must be JSON-marshalable/unmarshalable\nvar f func(int) = func(i int) { ... }\nvar w io.Writer\n\n// ❌ Only one identifier per declaration\nvar a, b int = 10, 20\nvar a, b int = f()\n&lt;/script&gt;</code></pre> <h3>Some valid state declarations:</h3> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n// ✅ Zero value\nvar id int64\n\n// ✅ With initial value (explicit type)\nvar address string = &#34;...&#34;\n\n// ✅ Type inferred from the initializer\nvar counter = 0\nvar fruits = []string{&#34;apple&#34;, &#34;banana&#34;}\nvar stock = map[string]int{&#34;apple&#34;: 10}\n\n// ✅ Initialized with a function call (assuming the package is imported)\nvar username = user.GetNameById(&#34;id&#34;)\n\n// ✅ Complex JSON-compatible types\nvar m map[string]int = map[string]int{&#34;key&#34;: 100}\n&lt;/script&gt;</code></pre> <p> The tmplx script cannot contain Go <code>type</code> declarations. To use your own struct or named types as state, declare them in a regular Go package and import them—then reference the imported type in the <code>var</code> declaration. </p> <h2 id=\"derived\">Derived</h2> A <strong>derived</strong> is a <strong>read-only</strong> value that is automatically calculated from states. It updates whenever those states change. <p> Declaring a derived works the same way as declaring package-level variables in Go. When the right-hand side of the declaration <strong>references existing state or other derived values</strong>, it is treated as a derived value. </p> <p> Derived values follow most of the same rules as regular state variables, but with some differences: </p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong> The type may be inferred from the right-hand side or declared explicitly</strong>, just like state. </li> <li> <strong>Derived values cannot be modified directly in event handlers, though they may be read.</strong> </li> </ol> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num1 int = 100 // state\n  var num2 int = num1 * 2 // derived\n&lt;/script&gt;\n\n...\n&lt;p&gt;{num1} * 2 = {num2}&lt;/p&gt;</code></pre> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var classStrs []string = []string{&#34;c1&#34;, &#34;c2&#34;, &#34;c3&#34;} // state\n  var class string = strings.Join(classStrs, &#34; &#34;) // derived\n&lt;/script&gt;\n\n...\n&lt;p class=&#34;{class}&#34;&gt; ... &lt;/p&gt;</code></pre> <h2 id=\"event-handler\">Event Handler</h2> <p> Event handlers let you respond to frontend events with backend logic or update state to trigger UI changes. </p> <p> To declare an event handler, define a Go function in the global scope of the <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> block. Bind it to a DOM event by adding an attribute that starts with <code>tx-on</code> followed by the event name (e.g., <code>tx-onclick</code>). </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func add1() {\n    counter += 1\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-onclick=&#34;add1()&#34;&gt;Add 1&lt;/button&gt;</code></pre> <p> In this example, the <code>add1</code> handler runs every time the button is clicked. The <code>counter</code> state increases by 1, and the paragraph updates automatically. </p> <p> It’s not magic. tmplx compiles each event handler into an HTTP endpoint. The runtime JavaScript attaches a lightweight listener that sends the required state to the endpoint, receives the updated HTML fragment, merges the new state, and swaps the affected part of the DOM. It feels like direct backend access from the client, but it’s just a simple API call with targeted DOM swapping. </p> <h3>Captured Locals</h3> <p> Any local variable bound by an enclosing <code>tx-for</code> init clause, <code>tx-for</code> range form, or <code>tx-if</code>/ <code>tx-else-if</code> init form is <strong>automatically captured</strong> by event handlers in the subtree. Just reference the local by name; the framework figures out which values cross the wire and decodes them with their inferred Go type. </p> <ul> <li> The captured local can appear anywhere in the handler body—as a function argument, an array index, an expression operand. Whatever is valid Go. </li> <li> Types are recovered from the binding’s context (the surrounding state declarations and imports), so you don’t need to annotate. </li> <li> State, derived, and prop variables are <strong>not</strong> captured this way—the handler already reads them from the component’s saved state. </li> </ul> ")
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_2_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 0\n\n  func addNum(num int) {\n    counter += num\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ counter }&lt;/p&gt;\n&lt;button tx-for=&#34;i := 0; i &lt; 10; i++&#34; tx-key=&#34;i&#34; tx-onclick=&#34;addNum(i)&#34;&gt;\n  +{ i }\n&lt;/button&gt;</code></pre> <p> <code>i</code> is captured from <code>tx-for</code> and decoded as <code>int</code> on the server. Captures from <code>tx-if</code>/ <code>tx-else-if</code> init forms work the same way: </p> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var stock = map[string]int{&#34;apple&#34;: 10, &#34;banana&#34;: 5}\n  var fruits = []string{&#34;apple&#34;, &#34;banana&#34;, &#34;cherry&#34;}\n&lt;/script&gt;\n\n&lt;div tx-for=&#34;_, f := range fruits&#34; tx-key=&#34;f&#34;&gt;\n  &lt;p tx-if=&#34;n, ok := stock[f]; ok&#34;&gt;\n    { f } in stock: { n }\n    &lt;button tx-onclick=&#34;stock[f] = n - 1&#34;&gt;buy&lt;/button&gt;\n  &lt;/p&gt;\n&lt;/div&gt;</code></pre> <p> Both <code>f</code> (from <code>tx-for</code>) and <code>n</code> (from <code>tx-if</code> init) are captured automatically. </p> <h3>Inline Statements</h3> <p> For simple actions, embed Go statements directly in <code>tx-on*</code> attributes to update state. This avoids defining separate handler functions. The body can be any valid Go statement sequence—direct state mutation, function calls, or compound expressions—not just a single function call. </p> ")
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_3_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var val int = 1\n&lt;/script&gt;\n\n&lt;p&gt;{ val }&lt;/p&gt;\n&lt;button tx-onclick=&#34;val *= 2&#34;&gt;double it!&lt;/button&gt;</code>\n      </pre> <h3>Event Properties</h3> <p> For input-typed events, you can read values directly from the DOM event object inside the handler body. The runtime injects them into the request automatically. </p> <ul> <li> <code>tx-oninput</code>, <code>tx-onchange</code>: <code>event.target.value</code> resolves to the input’s current string value. </li> </ul> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var query = &#34;&#34;\n&lt;/script&gt;\n\n&lt;input type=&#34;text&#34; tx-oninput=&#34;query = event.target.value&#34; /&gt;\n&lt;p&gt;You typed: { query }&lt;/p&gt;</code>\n      </pre> <h2 id=\"init\">init()</h2> <p> <code>init()</code> is a special function that runs automatically the first time a page or component is rendered. For pages, it runs on every GET request. For components, it runs when the component has no saved state yet (for example, the first time it appears on the page, or the first time a new <code>tx-for</code> iteration produces it). After that, subsequent renders reuse the saved state and skip <code>init()</code>. </p> ")
	{
		tx_id := "tx-example-wrapper-4"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_4_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var t string\n\n  func init() {\n    t = fmt.Sprint(time.Now().Format(time.RFC3339))\n  }\n&lt;/script&gt;\n\n&lt;p&gt;{ t }&lt;/p&gt;</code></pre> <p> Another common use case is to initialize one state from another state without turning the second variable into a derived state. </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var a int = 1\n  var b int\n\n  func init() {\n    b = a * 2 // b remains a regular state\n  }\n&lt;/script&gt;</code></pre> <h2 id=\"path-parameter\">Path Parameters</h2> <p> When a page route contains a wildcard (see <a href=\"#pages-and-routing\">Pages and Routing</a>), you can pull the captured value into a state variable by annotating the declaration with a <code>//tx:path</code> comment. </p> <p>Rules:</p> <ul> <li> The comment must sit directly above the <code>var</code> line (Go doc-comment position). </li> <li> The value after <code>tx:path</code> is the wildcard name from the route pattern. </li> <li> The variable must be declared as <code>string</code>. No initial value is allowed—the captured string is the initial value. </li> <li> Only <a href=\"#pages-and-routing\">pages</a> support <code>tx:path</code>; components cannot declare path-bound state. </li> </ul> <p> The captured value is assigned <strong>before</strong> <a href=\"#init\"><code>init()</code></a> runs, so <code>init()</code> can use it to populate other state (for example, by loading a record from the database). </p> <p> <strong>Single parameter.</strong> For a route <code>pages/blog/post/{post_id}.html</code>: </p> <pre><code>&lt;!DOCTYPE html&gt;\n&lt;html&gt;\n  &lt;head&gt;\n    &lt;script type=&#34;text/tmplx&#34;&gt;\n      // tx:path post_id\n      var postId string\n\n      var post Post\n\n      func init() {\n        post = db.GetPost(postId)\n      }\n    &lt;/script&gt;\n  &lt;/head&gt;\n  &lt;body&gt;\n    &lt;h1&gt;{ post.Title }&lt;/h1&gt;\n  &lt;/body&gt;\n&lt;/html&gt;</code></pre> <p> <strong>Multiple parameters.</strong> Each wildcard gets its own declaration. For a route <code>pages/blog/{year}/{slug}.html</code>: </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  // tx:path year\n  var year string\n\n  // tx:path slug\n  var slug string\n&lt;/script&gt;\n\n&lt;p&gt;Viewing { slug } from { year }&lt;/p&gt;</code></pre> <p> After initialization, the variable behaves like any other state: it&#39;s serialized, sent to the server on events, and can be reassigned from handlers (though reassigning it does not change the URL). </p> <h2 id=\"control-flow\">Control Flow</h2> <p> tmplx avoids new custom syntax for conditionals and loops because that would increase compiler complexity. Instead, it embeds control flow directly into HTML attributes, similar to Vue.js and <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h3 id=\"conditionals\">Conditionals</h3> <p> To conditionally render elements, use the <code>tx-if</code>, <code>tx-else-if</code>, and <code>tx-else</code> attributes on the desired tags. The values for <code>tx-if</code> and <code>tx-else-if</code> can be any valid Go expression that would fit in an <code>if</code> or <code>else if</code> statement. The <code>tx-else</code> attribute needs no value. </p> ")
	{
		tx_id := "tx-example-wrapper-5"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_5_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var num int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;num++&#34;&gt;change&lt;/button&gt;\n&lt;div&gt;\n  &lt;p tx-if=&#34;num % 3 == 0&#34; style=&#34;background: red; color: white&#34;&gt;red&lt;/p&gt;\n  &lt;p tx-else-if=&#34;num % 3 == 1&#34; style=&#34;background: blue; color: white&#34;&gt;blue&lt;/p&gt;\n  &lt;p tx-else style=&#34;background: green; color: white&#34;&gt;green&lt;/p&gt;\n&lt;/div&gt;</code>\n      </pre> <p> You can declare <strong>local variables</strong> and handle errors exactly as you would in regular Go code. Local variables declared in conditionals are available to the element and its descendants, just like in Go. </p> <pre><code>&lt;p tx-if=&#34;user, err := user.GetUser(); err != nil&#34;&gt;\n  &lt;span tx-if=&#34;err == ErrNotFound&#34;&gt;User not found&lt;/span&gt;\n&lt;/p&gt;\n&lt;p tx-else-if=&#39;user.Name == &#34;&#34;&#39;&gt;user.Name not set&lt;/p&gt;\n&lt;p tx-else&gt;Hi, { user.Name }&lt;/p&gt;</code></pre> <p> A conditional group consists of <strong>consecutive sibling nodes</strong> that share the same parent. Disconnected nodes are not treated as part of the same group. A standalone <code>tx-else-if</code> or <code>tx-else</code> without a preceding <code>tx-if</code> will cause a compilation error. </p> <h3 id=\"loops\">Loops</h3> <p> To repeat elements, use the <code>tx-for</code> attribute. Its value can be any valid Go <code>for</code> statement, including <strong>classic for</strong> or <strong>range for</strong>. </p> <p> Local variables declared in the loop are available to the element and all of its descendants, just like in Go. </p> <p> Always add a <code>tx-key</code> attribute with a unique value for each item. This gives the compiler a unique identifier for the node during updates. </p> ")
	{
		tx_id := "tx-example-wrapper-6"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_6_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code>\n      </pre> <pre><code>&lt;div tx-for=&#34;_, user := range users&#34; tx-key=&#34;user.Id&#34;&gt;\n  { user.Id }: { user.Name }\n&lt;/div&gt;</code></pre> <h2 id=\"template\">&lt;template&gt;</h2> <p> The <code>&lt;template&gt;</code> tag is a non-rendering container that lets you apply control flow attributes (<code>tx-if</code>, <code>tx-else-if</code>, <code>tx-else</code>, or <code>tx-for</code>) to a group of elements at once. </p> <p> The <code>&lt;template&gt;</code> itself is removed from the output; only its children are rendered (or not, depending on the control flow). </p> <p> You can nest <code>&lt;template&gt;</code> tags and combine them with other control flow attributes on child elements. </p> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var loggedIn bool = true\n&lt;/script&gt;\n\n&lt;template tx-if=&#34;loggedIn&#34;&gt;\n  &lt;p&gt;Welcome back!&lt;/p&gt;\n  &lt;button tx-onclick=&#34;logout()&#34;&gt;Logout&lt;/button&gt;\n&lt;/template&gt;\n\n&lt;template tx-else&gt;\n  &lt;p&gt;Please sign in.&lt;/p&gt;\n  &lt;button tx-onclick=&#34;login()&#34;&gt;Login&lt;/button&gt;\n&lt;/template&gt;</code>\n      </pre> <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var posts []Post = []Post{\n    {Title: &#34;First Post&#34;, Body: &#34;Hello world&#34;},\n    {Title: &#34;Second Post&#34;, Body: &#34;tmplx is great&#34;},\n  }\n&lt;/script&gt;\n\n&lt;template tx-for=&#34;i, p := range posts&#34; tx-key=&#34;i&#34;&gt;\n  &lt;article&gt;\n    &lt;h3&gt;{ p.Title }&lt;/h3&gt;\n    &lt;p&gt;{ p.Body }&lt;/p&gt;\n    &lt;hr&gt;\n  &lt;/article&gt;\n&lt;/template&gt;</code>\n      </pre> <h2 id=\"forms\">Forms</h2> <p> Attach a handler to a <code>&lt;form&gt;</code> with <code>tx-action</code>. When the form is submitted, tmplx cancels the default submission, collects every named form element, and calls the handler on the server. </p> <p> The value of <code>tx-action</code> must be the name of a function declared in the tmplx script. Each form element&#39;s <code>name</code> attribute must match a parameter name on that function; unnamed elements are ignored. </p> ")
	{
		tx_id := "tx-example-wrapper-7"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_7_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var greeting string\n\n  func greet(name string) {\n    greeting = &#34;Hello, &#34; + name\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;greet&#34;&gt;\n  &lt;input name=&#34;name&#34; type=&#34;text&#34; required /&gt;\n  &lt;button type=&#34;submit&#34;&gt;Greet&lt;/button&gt;\n&lt;/form&gt;\n\n&lt;p tx-if=&#39;greeting != &#34;&#34;&#39;&gt;{ greeting }&lt;/p&gt;</code></pre> <p> Values are JSON-decoded into each parameter&#39;s Go type, so the parameter type is what determines how the string is parsed. The runtime serializes form elements by input type: </p> <ul> <li> <code>text</code>, <code>email</code>, <code>password</code>, <code>textarea</code>, <code>select</code>, etc.—sent as a JSON string. Decode into <code>string</code>. </li> <li> <code>number</code>, <code>range</code>—sent as the raw numeric value, or <code>null</code> when empty. Decode into a numeric type or pointer. </li> <li> <code>checkbox</code>—sent as <code>true</code> or <code>false</code>. Decode into <code>bool</code>. </li> <li> <code>radio</code>—only the checked radio in a group is sent (using its shared <code>name</code>). Decode into <code>string</code>. </li> </ul> <p> Because submission goes through a full server round-trip, use native HTML validation (<code>required</code>, <code>minlength</code>, <code>pattern</code>, ...) to catch client-side errors before the request is sent. For richer live-updating inputs, combine tmplx with a client-side library like <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h2 id=\"component\">Component</h2> <p> Components are reusable UI building blocks that encapsulate HTML, state, and behavior. </p> <p> Create a component by placing an <code>.html</code> file in the <code>components</code> directory (default: <code>./components</code>). tmplx automatically registers it as a custom element with the tag name <code>tx-</code> followed by the relative path (without the <code>.html</code> extension), with directory separators replaced by <code>-</code>. </p> <p> Filenames and directory names may contain only <code>a-z</code>, <code>0-9</code>, <code>-</code>, and <code>_</code>. Uppercase letters are rejected. </p> <p>Examples:</p> <ul> <li> <code>components/button.html</code> → <code>&lt;tx-button&gt;</code> </li> <li> <code>components/user/card.html</code> → <code>&lt;tx-user-card&gt;</code> </li> <li> <code>components/todo/list.html</code> → <code>&lt;tx-todo-list&gt;</code> </li> </ul> <p> Components can contain their own <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> for local state and logic, and can be used in pages or nested inside other components. </p> <h3 id=\"props\">Props</h3> <p> Props are inputs the parent passes to a child component. Inside the child, a prop is declared like a state variable, but with a <code>//tx:prop</code> doc comment. </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  //tx:prop\n  var title string\n\n  //tx:prop\n  var count int = 0\n&lt;/script&gt;\n\n&lt;h3&gt;{ title }&lt;/h3&gt;\n&lt;span&gt;{ count } items&lt;/span&gt;</code></pre> <p>Rules:</p> <ul> <li> The <code>//tx:prop</code> comment must sit directly above the <code>var</code> line. </li> <li> Prop names must be <strong>lowercase</strong>. HTML lowercases attribute names, so a camelCase prop name would never match the attribute the parent writes. </li> <li> An initial value (e.g. <code>= 0</code>) becomes the <strong>default</strong> used when the parent omits the attribute. </li> <li> Props are <strong>read-only</strong> inside the child. Event handlers can read them but cannot assign to them. Derived values referencing a prop recompute automatically when the prop changes. </li> <li> Pages cannot declare props—only components can. </li> </ul> <h4>Passing props</h4> <p> Prop attribute values on the parent are parsed as <strong>Go expressions</strong>, not as plain strings. Pass a literal by writing the literal directly; pass a parent variable by its name. </p> <pre><code>&lt;!-- Go string literal (quotes are part of the expression) --&gt;\n&lt;tx-card title=&#39;&#34;Hello&#34;&#39; count=&#34;5&#34;&gt;&lt;/tx-card&gt;\n\n&lt;!-- A parent state/derived/prop variable by name --&gt;\n&lt;tx-card title=&#34;heading&#34; count=&#34;itemCount&#34;&gt;&lt;/tx-card&gt;\n\n&lt;!-- Any Go expression that matches the prop type --&gt;\n&lt;tx-card title=&#39;strings.ToUpper(heading)&#39; count=&#34;len(items)&#34;&gt;&lt;/tx-card&gt;</code></pre> <p> The expression is re-evaluated whenever the parent re-renders, so the child stays in sync with the parent&#39;s state automatically. </p> <h4 id=\"callback-props\">Callback Props</h4> <p> A <strong>callback prop</strong> lets a child notify the parent when something happens. It is just a prop whose type is a <strong>function</strong>: declare it with <code>//tx:prop</code> and a function type. With no default the parent <strong>must</strong> supply an implementation (a required prop); give it a function-literal default to make the parent override optional. </p> <p>In the child, call it like any other event handler:</p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  //tx:prop\n  var onselect func(id int)\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;onselect(42)&#34;&gt;Pick&lt;/button&gt;</code></pre> <p> In the parent, pass the <strong>bare name</strong> of a tmplx-script function as the attribute whose key matches the child&#39;s prop name: </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var selected int\n\n  func pick(id int) {\n    selected = id\n  }\n&lt;/script&gt;\n\n&lt;tx-picker onselect=&#34;pick&#34;&gt;&lt;/tx-picker&gt;</code></pre> <p> When the child calls <code>onselect(42)</code>, the parent&#39;s <code>pick</code> runs on the server with that argument and the parent re-renders. A callback call can be mixed freely with other statements in the same handler—for example <code>tx-onclick=&#34;count++; onselect(42)&#34;</code>. </p> <p> To make the override optional, give the prop a function-literal default. The parent may then omit the attribute and the child falls back to its own implementation: </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  //tx:prop\n  var onselect func(id int) = func(id int) {}\n&lt;/script&gt;</code></pre> <h3 id=\"slot\">&lt;slot&gt;</h3> <p> A <code>&lt;slot&gt;</code> marks a place in a component&#39;s template where the parent can inject content. Slots are how components stay composable: the child decides the shape, the parent fills in the details. </p> <h4>Declaring slots in a component</h4> <p> Each slot is either the <strong>default slot</strong> (no <code>name</code>) or a <strong>named slot</strong>. A component may have at most one default slot, and named slots must be unique. Slots cannot be nested inside other slots. </p> <pre><code>&lt;div class=&#34;card&#34;&gt;\n  &lt;slot name=&#34;header&#34;&gt;Default Header&lt;/slot&gt;\n  &lt;div class=&#34;body&#34;&gt;\n    &lt;slot&gt;Default Body&lt;/slot&gt;\n  &lt;/div&gt;\n  &lt;slot name=&#34;footer&#34;&gt;&lt;/slot&gt;\n&lt;/div&gt;</code></pre> <p> Content placed inside <code>&lt;slot&gt;...&lt;/slot&gt;</code> is <strong>fallback content</strong>—it renders only when the parent does not fill that slot. </p> <h4>Filling slots from the parent</h4> <p> Put fill content directly inside the component tag. Use the <code>slot</code> attribute on a child element to target a named slot; everything else becomes the default fill. </p> <pre><code>&lt;tx-card&gt;\n  &lt;h2 slot=&#34;header&#34;&gt;Custom Title&lt;/h2&gt;\n  &lt;p&gt;Custom content goes in the default slot.&lt;/p&gt;\n  &lt;div slot=&#34;footer&#34;&gt;Actions&lt;/div&gt;\n&lt;/tx-card&gt;</code></pre> <p> Only the <strong>direct children</strong> of the component tag are considered when matching slots—a <code>slot</code> attribute on a deeply nested element has no effect. </p> <h4>Scope: fills use the parent&#39;s state</h4> <p> This is the most important rule. The content you pass into a slot is still <strong>parent code</strong>: expressions, event handlers, and directives inside a fill see the parent&#39;s state, derived, and prop variables—not the child&#39;s. </p> <pre><code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var user string = &#34;tmplx&#34;\n\n  func logout() {\n    user = &#34;&#34;\n  }\n&lt;/script&gt;\n\n&lt;tx-card&gt;\n  &lt;h2 slot=&#34;header&#34;&gt;Hello, { user }&lt;/h2&gt;\n  &lt;button tx-onclick=&#34;logout()&#34;&gt;Sign out&lt;/button&gt;\n&lt;/tx-card&gt;</code></pre> <p> Here <code>user</code> and <code>logout</code> are defined on the page that uses <code>&lt;tx-card&gt;</code>, not inside the card component. When the button is clicked the page&#39;s handler runs and the fill re-renders against the page&#39;s updated state. </p> <h4>Live example</h4> <p> The docs site uses a simple <code>&lt;tx-example-wrapper&gt;</code> component with a single default slot to frame every live demo on this page. The component is just: </p> <pre><code>&lt;div class=&#34;example-frame&#34;&gt;\n  &lt;slot&gt;&lt;/slot&gt;\n&lt;/div&gt;</code></pre> <p>And callers wrap any demo with it:</p> <pre><code>&lt;tx-example-wrapper&gt;\n  &lt;tx-counter&gt;&lt;/tx-counter&gt;\n&lt;/tx-example-wrapper&gt;</code></pre> <h2 id=\"cli\">CLI</h2> <p> Running <code>tmplx</code> inside any directory of your Go module walks up to the nearest <code>go.mod</code> and uses that as the project root. All path flags default relative to that root. </p> <table> <thead> <tr> <th>Flag</th> <th>Default</th> <th>Description</th> </tr> </thead> <tbody> <tr> <td><code>-pages-dir</code></td> <td><code>./pages</code></td> <td>Directory containing pages.</td> </tr> <tr> <td><code>-components-dir</code></td> <td><code>./components</code></td> <td>Directory containing reusable components.</td> </tr> <tr> <td><code>-output-file</code></td> <td><code>./routes.go</code></td> <td>Path to the generated Go file.</td> </tr> <tr> <td><code>-package-name</code></td> <td><code>main</code></td> <td>Package name for the generated Go code.</td> </tr> <tr> <td><code>-handler-prefix</code></td> <td><code>/tx/</code></td> <td>URL path prefix for generated event handler routes.</td> </tr> </tbody> </table> <h2 id=\"syntax-highlight\">Syntax Highlight</h2> <a href=\"https://github.com/gnituy18/tmplx.nvim\">Neovim Plugin</a> </main> </body></html>")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_1_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-todo-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_todo)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_2_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-addn-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_addn)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_3_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-double-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_double)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_4_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-current-time-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_current_H_time)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_5_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-cond-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_cond)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_6_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-triangle-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_triangle)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S_docs) tx_render_fill_tx_H_example_H_wrapper_7_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-greeting-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_greeting)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

type tx_S_examples_S__EX_ struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_examples_S__EX_(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_examples_S__EX_ {
	tx_comp := &tx_S_examples_S__EX_{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_S_examples_S__EX_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<html><head> <title>tmplx fixture</title> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/examples/{$}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <h1>tmplx fixture</h1> <ul> <li><a href=\"/state\">state</a> — state variables, initial values, interpolation</li> </ul> </body></html>")
}

type tx_S_examples_S_state struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_count int    `json:"count"`
	V_label string `json:"label"`
	V_flag  bool   `json:"flag"`
}

func tx_new_tx_S_examples_S_state(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_examples_S_state {
	tx_comp := &tx_S_examples_S_state{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_count = 42
		tx_comp.V_label = "hello"
		tx_comp.V_flag = true
	}
	return tx_comp
}

func (tx_comp *tx_S_examples_S_state) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<html><head>  <title>state</title> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/examples/state\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <h1>state</h1> <p>int state with initial value: <b id=\"count\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count)))
	tx_w2.WriteString("</b> (expect: 42)</p> <p>string state with initial value: <b id=\"label\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_label)))
	tx_w2.WriteString("</b> (expect: hello)</p> <p>bool state with initial value: <b id=\"flag\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_flag)))
	tx_w2.WriteString("</b> (expect: true)</p> </body></html>")
}

type tx_S__EX_ struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S__EX_(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S__EX_ {
	tx_comp := &tx_S__EX_{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_S__EX_) tx_compute() {
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-counter-1"
			tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-todo-1"
			tx_child := tx_new_tx_H_todo(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_new_tx_H_example_H_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-triangle-1"
			tx_child := tx_new_tx_H_triangle(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
}

func (tx_comp *tx_S__EX_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/tokyo-night-dark.min.css\"/> <script src=\"https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js\"></script> <script>\n      hljs.highlightAll();\n    </script> <link rel=\"stylesheet\" href=\"/style.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/{$}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <main> <h1 style=\"text-align: center\">&lt;tmplx&gt;</h1> <section class=\"logo-gallery\"> <h2 style=\"text-align: center; margin: 0\">Logo styles</h2> <div class=\"logo-variant\"> <div class=\"label\">1. Current — Verdana + shadow</div> <span class=\"logo-original\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">2. Mono + colored brackets</div> <span class=\"logo-mono\"><span class=\"bracket\">&lt;</span>tmplx<span class=\"bracket\">&gt;</span></span> </div> <div class=\"logo-variant\"> <div class=\"label\">3. DOS dialog box</div> <span class=\"logo-dos\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">4. CRT terminal</div> <span class=\"logo-crt\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">5. Win95 beveled</div> <span class=\"logo-win95\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">6. GeoCities rainbow</div> <span class=\"logo-geocities\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">7. Old newspaper</div> <span class=\"logo-newspaper\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">8. Pixel / chromatic-aberration</div> <span class=\"logo-pixel\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">9. Macintosh classic</div> <span class=\"logo-mac\">&lt;tmplx&gt;</span> </div> <div class=\"logo-variant\"> <div class=\"label\">10. Scrolling marquee</div> <div class=\"logo-marquee\"><span>&lt;tmplx&gt;</span></div> </div> </section> <h2 style=\"text-align: center; margin-top: 1.5rem\"> Write Go in HTML intuitively </h2> <ul style=\"margin-top: 4rem\"> <li>Full Go backend logic and HTML in the same file</li> <li>Reactive UIs driven by plain Go variables</li> <li>Reusable components written as regular HTML files</li> </ul> <div style=\"\n          display: flex;\n          gap: 2rem;\n          justify-content: center;\n          text-align: center;\n          margin-top: 4rem;\n        \"> <a class=\"btn\" href=\"/docs\">Docs</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> <p style=\"text-align: center; margin-top: 1.5rem\"> or see the <a href=\"/roadmap\">roadmap</a> </p> <h2 style=\"text-align: center\">Demos</h2> <h3>Counter</h3> ")
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_1_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int\n&lt;/script&gt;\n\n&lt;button tx-onclick=&#34;counter--&#34;&gt;-&lt;/button&gt;\n&lt;span&gt; { counter } &lt;/span&gt;\n&lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;</code>\n      </pre> <h3>To Do</h3> ")
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_2_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code class=\"language-html\">&lt;script type=&#34;text/tmplx&#34;&gt;\n  var list []string\n\n  func add(item string) {\n    list = append(list, item)\n  }\n\n  func remove(i int) {\n    list = append(list[0:i], list[i+1:]...)\n  }\n&lt;/script&gt;\n\n&lt;form tx-action=&#34;add&#34;&gt;\n  &lt;label&gt;&lt;input name=&#34;item&#34; type=&#34;text&#34; required&gt;&lt;/label&gt;\n  &lt;button type=&#34;submit&#34;&gt;Add&lt;/button&gt;\n&lt;/form&gt;\n&lt;ol&gt;\n  &lt;li\n    tx-for=&#34;i, l := range list&#34;\n    tx-key=&#34;l&#34;\n    tx-onclick=&#34;remove(i)&#34;&gt;\n    { l }\n  &lt;/li&gt;\n&lt;/ol&gt;</code>\n      </pre> <h3>Triangle</h3> ")
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_example_H_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_example_H_wrapper_3_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <pre>        <code>&lt;script type=&#34;text/tmplx&#34;&gt;\n  var counter int = 5\n&lt;/script&gt;\n\n&lt;div&gt;\n  &lt;span&gt; { counter } &lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/div&gt;\n&lt;div tx-for=&#34;h := 0; h &lt; counter; h++&#34; tx-key=&#34;h&#34;&gt;\n  &lt;span tx-for=&#34;s := 0; s &lt; counter-h-1; s++&#34; tx-key=&#34;s&#34;&gt;_&lt;/span&gt;\n  &lt;span tx-for=&#34;i := 0; i &lt; h*2+1; i++&#34; tx-key=&#34;i&#34;&gt;*&lt;/span&gt;\n&lt;/div&gt;</code>\n      </pre> </main> </body></html>")
}

func (tx_comp *tx_S__EX_) tx_render_fill_tx_H_example_H_wrapper_1_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S__EX_) tx_render_fill_tx_H_example_H_wrapper_2_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-todo-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_todo)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_S__EX_) tx_render_fill_tx_H_example_H_wrapper_3_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-triangle-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_triangle)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

type tx_S_roadmap struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_roadmap(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_roadmap {
	tx_comp := &tx_S_roadmap{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_S_roadmap) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Roadmap | tmplx</title> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"/style.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/roadmap\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <main> <h1>Roadmap</h1> <p> tmplx is pre-1.0 and moving fast. Expect breaking changes between minor versions until 1.0. For the full record of released changes, see the <a href=\"https://github.com/gnituy18/tmplx/blob/master/CHANGELOG.md\">changelog</a>. </p> <ul> <li><code>[Compiler]</code> for work inside the compiler</li> <li><code>[DX]</code> fro tools around the compiler.</li> <li><code>[Learning]</code> for docs, examples, playground, and other learning material.</li> </ul> <h2>In progress (toward 0.1.0)</h2> <ul> <li><input type=\"checkbox\" checked=\"\"/> [Compiler] A stable product that can be used as a benchmark for progress</li> <li><input type=\"checkbox\"/> [DX] Test suite scaffolding</li> <li><input type=\"checkbox\"/> [Learning] Docs</li> <li><input type=\"checkbox\"/> [Learning] Examples</li> <li><input type=\"checkbox\"/> A Logo</li> </ul> <h2>Planned for 0.2</h2> <ul> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Verifiable Go imports in tmplx script</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Detect unused fills</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Detect unreachable conditional branches</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Type-check template expressions against the Go types they reference</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Validate <code>//tx:path</code> matches a path segment on the page route</li> <li><input type=\"checkbox\" disabled=\"\"/> [DX] Language server</li> <li><input type=\"checkbox\" disabled=\"\"/> [DX] Tree-sitter grammar</li> <li><input type=\"checkbox\" disabled=\"\"/> [Learning] Tutorial</li> </ul> <h2>Planned for 0.3+</h2> <ul> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] DOM morphing</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] Scoped <code>&lt;style&gt;</code> in components</li> <li><input type=\"checkbox\" disabled=\"\"/> [Compiler] <code>tx-class</code> and <code>tx-style</code></li> <li><input type=\"checkbox\" disabled=\"\"/> [Learning] In-browser playground</li> </ul> <h2>Considering</h2> <ul> <li>Compressing the embedded <code>tx-saved</code> state</li> </ul> </main> </body></html>")
}

func tx_dispatch(tx_w http.ResponseWriter, tx_r *http.Request) {
	tx_r.ParseForm()
	tx_prev := tx_r.PostForm
	tx_target := tx_r.PostFormValue("target")
	tx_trigger := tx_r.PostFormValue("trigger")
	tx_trigger_handler := tx_r.URL.Path
	if i := strings.LastIndexByte(tx_trigger_handler, '/'); i >= 0 {
		tx_trigger_handler = tx_trigger_handler[i+1:]
	}
	tx_next := map[string]any{}
	switch tx_target {
	case "/docs":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_docs(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/examples/{$}":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_examples_S__EX_(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/examples/state":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_examples_S_state(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/{$}":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S__EX_(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/roadmap":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_roadmap(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	}
	seg := tx_target
	if i := max(strings.LastIndexByte(seg, ':'), strings.LastIndexByte(seg, '@')); i >= 0 {
		seg = seg[i+1:]
	}
	name := seg
	if i := strings.LastIndexByte(name, '-'); i >= 0 {
		name = name[:i]
	}
	var buf bytes.Buffer
	switch name {
	case "tx-addn":
		tx_comp := tx_new_tx_H_addn(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-cond":
		tx_comp := tx_new_tx_H_cond(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-counter":
		tx_comp := tx_new_tx_H_counter(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-current-time":
		tx_comp := tx_new_tx_H_current_H_time(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-double":
		tx_comp := tx_new_tx_H_double(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-double-state":
		tx_comp := tx_new_tx_H_double_H_state(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-example-wrapper":
		tx_comp := tx_new_tx_H_example_H_wrapper(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target, nil)
	case "tx-greeting":
		tx_comp := tx_new_tx_H_greeting(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-todo":
		tx_comp := tx_new_tx_H_todo(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-triangle":
		tx_comp := tx_new_tx_H_triangle(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	default:
		return
	}
	tx_json, _ := json.Marshal(tx_next)
	buf.WriteString("<script type=\"application/json\" id=\"tx-saved\">")
	buf.Write(tx_json)
	buf.WriteString("</script>")
	tx_w.Write(buf.Bytes())
}

type TxRoute struct {
	Pattern string
	Handler http.HandlerFunc
}

var tx_routes []TxRoute = []TxRoute{
	{
		Pattern: "GET /docs",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_docs(nil, tx_next, "", "")
			tx_next["page"] = tx_comp
			tx_comp.tx_compute()
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /examples/{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_examples_S__EX_(nil, tx_next, "", "")
			tx_next["page"] = tx_comp
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /examples/state",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_examples_S_state(nil, tx_next, "", "")
			tx_next["page"] = tx_comp
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S__EX_(nil, tx_next, "", "")
			tx_next["page"] = tx_comp
			tx_comp.tx_compute()
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /roadmap",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_roadmap(nil, tx_next, "", "")
			tx_next["page"] = tx_comp
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "POST /tx/tx-addn/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-cond/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-counter/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-counter/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-double/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-greeting/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-todo/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-todo/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-triangle/eh1",
		Handler: tx_dispatch,
	},
}

func Routes() []TxRoute { return tx_routes }

var tx_runtime_script = `document.addEventListener('DOMContentLoaded', function() {
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

    const res = await fetch("/tx/" + fun, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString() })
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
`
