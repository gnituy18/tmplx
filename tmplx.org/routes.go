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
	"tmplx.org/albums"
)

type tx_HY_addn struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_HY_addn(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_addn {
	tx_comp := &tx_HY_addn{}
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

func (tx_comp *tx_HY_addn) addNum(num int) {
	tx_comp.V_counter += num
}

func (tx_comp *tx_HY_addn) tx_eh1(i int) {
	tx_comp.addNum(i)
}

func (tx_comp *tx_HY_addn) tx_compute(tx_id string) {
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

func (tx_comp *tx_HY_addn) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>")
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
	tx_w.WriteString(" ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_callback struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_total int `json:"total"`
}

func tx_new_tx_HY_callback(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_callback {
	tx_comp := &tx_HY_callback{}
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

func (tx_comp *tx_HY_callback) add() {
	tx_comp.V_total++
}

func (tx_comp *tx_HY_callback) tx_compute(tx_id string) {
	{
		tx_id := tx_id + ":tx-incbtn-1"
		tx_val_label := "press me"
		tx_child := tx_new_tx_HY_incbtn(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_comp.tx_target, &tx_val_label, tx_comp.add)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_HY_callback) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>total: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_total)))
	tx_w.WriteString("</p> ")
	{
		tx_id := tx_id + ":tx-incbtn-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_incbtn)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_cond struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_num int `json:"num"`
}

func tx_new_tx_HY_cond(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_cond {
	tx_comp := &tx_HY_cond{}
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

func (tx_comp *tx_HY_cond) tx_eh1() {
	tx_comp.V_num++
}

func (tx_comp *tx_HY_cond) tx_compute(tx_id string) {
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

func (tx_comp *tx_HY_cond) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <button data-tx-trigger=\"")
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
	tx_w.WriteString("</div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_condrows struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_nums []int `json:"nums"`
}

func tx_new_tx_HY_condrows(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_condrows {
	tx_comp := &tx_HY_condrows{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_nums = []int{1, 2, 3, 4, 5, 6}
	}
	return tx_comp
}

func (tx_comp *tx_HY_condrows) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <ul> ")

	for _, n := range tx_comp.V_nums {
		_ = n
		tx_w.WriteString("<li> ")
		if n > 3 {
			tx_w.WriteString("<b>")
			tx_w.WriteString(html.EscapeString(fmt.Sprint(n)))
			tx_w.WriteString(" (big)</b> ")
		} else {
			tx_w.WriteString("<i>")
			tx_w.WriteString(html.EscapeString(fmt.Sprint(n)))
			tx_w.WriteString(" (small)</i> ")

		}
		tx_w.WriteString("</li>")

	}
	tx_w.WriteString(" </ul> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_counter struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_HY_counter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_counter {
	tx_comp := &tx_HY_counter{}
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

func (tx_comp *tx_HY_counter) tx_eh1() {
	tx_comp.V_counter--
}

func (tx_comp *tx_HY_counter) tx_eh2() {
	tx_comp.V_counter++
}

func (tx_comp *tx_HY_counter) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		case "eh2":
			tx_comp.tx_eh2()
		}
	}
}

func (tx_comp *tx_HY_counter) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">-</button> <span> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w.WriteString(" </span> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh2-on=\"click\">+</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_current_HY_time struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_t string `json:"t"`
}

func tx_new_tx_HY_current_HY_time(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_current_HY_time {
	tx_comp := &tx_HY_current_HY_time{}
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

func (tx_comp *tx_HY_current_HY_time) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_t)))
	tx_w.WriteString("</p> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_derived struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_count   int `json:"count"`
	V_doubled int `json:"-"`
}

func tx_new_tx_HY_derived(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_derived {
	tx_comp := &tx_HY_derived{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_doubled = tx_comp.V_count * 2
	} else {
		tx_comp.V_count = 1
		tx_comp.V_doubled = tx_comp.V_count * 2
	}
	return tx_comp
}

func (tx_comp *tx_HY_derived) tx_eh1() {
	tx_comp.V_count++
	tx_comp.V_doubled = tx_comp.V_count * 2
}

func (tx_comp *tx_HY_derived) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_HY_derived) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>count = ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count)))
	tx_w.WriteString(", doubled = ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_doubled)))
	tx_w.WriteString("</p> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">increment</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_double struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_val int `json:"val"`
}

func tx_new_tx_HY_double(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_double {
	tx_comp := &tx_HY_double{}
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

func (tx_comp *tx_HY_double) tx_eh1() {
	tx_comp.V_val *= 2
}

func (tx_comp *tx_HY_double) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_HY_double) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_val)))
	tx_w.WriteString("</p> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">double it!</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_double_HY_state struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_b int `json:"b"`
	V_a int `json:"a"`
}

func tx_new_tx_HY_double_HY_state(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_double_HY_state {
	tx_comp := &tx_HY_double_HY_state{}
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

func (tx_comp *tx_HY_double_HY_state) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div> ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_b * tx_comp.V_a)))
	tx_w.WriteString(" </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_example_HY_wrapper struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_HY_example_HY_wrapper(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_example_HY_wrapper {
	tx_comp := &tx_HY_example_HY_wrapper{}
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

func (tx_comp *tx_HY_example_HY_wrapper) tx_render(tx_w *bytes.Buffer, tx_id string, tx_render_fill_ func()) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString("<div class=\"example-box\"> <div> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	}
	tx_w.WriteString(" </div> </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_greeting struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_greeting string `json:"greeting"`
}

func tx_new_tx_HY_greeting(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_greeting {
	tx_comp := &tx_HY_greeting{}
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

func (tx_comp *tx_HY_greeting) greet(name string) {
	tx_comp.V_greeting = "Hello, " + name
}

func (tx_comp *tx_HY_greeting) tx_eh1(name string) {
	tx_comp.greet(name)
}

func (tx_comp *tx_HY_greeting) tx_compute(tx_id string) {
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

func (tx_comp *tx_HY_greeting) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <form data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-action=\"tx-greeting/eh1\"> <input name=\"name\" type=\"text\" required=\"\"/> <button type=\"submit\">Greet</button> </form> ")
	if tx_comp.V_greeting != "" {
		tx_w.WriteString("<p>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_greeting)))
		tx_w.WriteString("</p> ")

	}
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_incbtn struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_label   *string `json:"-"`
	V_onpress func()  `json:"-"`
}

func tx_new_tx_HY_incbtn(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, label *string, onpress func()) *tx_HY_incbtn {
	tx_comp := &tx_HY_incbtn{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_label = label
	tx_comp.V_onpress = onpress
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_HY_incbtn) tx_eh1() {
	tx_comp.V_onpress()
}

func (tx_comp *tx_HY_incbtn) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_HY_incbtn) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_label)))
	tx_w.WriteString("</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_inputlive struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_inputValue string `json:"inputValue"`
}

func tx_new_tx_HY_inputlive(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_inputlive {
	tx_comp := &tx_HY_inputlive{}
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

func (tx_comp *tx_HY_inputlive) tx_eh1(tx_ev_target_value string) {
	tx_comp.V_inputValue = tx_ev_target_value
}

func (tx_comp *tx_HY_inputlive) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var tx_ev_target_value string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("tx_ev_target_value")), &tx_ev_target_value)
			tx_comp.tx_eh1(tx_ev_target_value)
		}
	}
}

func (tx_comp *tx_HY_inputlive) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <input type=\"text\" placeholder=\"type here\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"input\"/> <p>You typed: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_inputValue)))
	tx_w.WriteString(" (")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_inputValue))))
	tx_w.WriteString(" chars)</p> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_nav struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_HY_nav(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_nav {
	tx_comp := &tx_HY_nav{}
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

func (tx_comp *tx_HY_nav) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString("<header class=\"bar\"> <a class=\"bar-brand\" href=\"/\"> <img src=\"/logo.svg\" alt=\"\" width=\"28\" height=\"28\"/> tmplx </a> <nav class=\"bar-links\"> <a href=\"/docs\">Docs</a> <a href=\"/playground\">Playground</a> <a href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </nav> </header> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_props struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_clicks  int `json:"clicks"`
	V_doubled int `json:"-"`
}

func tx_new_tx_HY_props(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_props {
	tx_comp := &tx_HY_props{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_doubled = tx_comp.V_clicks * 2
	} else {
		tx_comp.V_doubled = tx_comp.V_clicks * 2
	}
	return tx_comp
}

func (tx_comp *tx_HY_props) tx_eh1() {
	tx_comp.V_clicks++
	tx_comp.V_doubled = tx_comp.V_clicks * 2
}

func (tx_comp *tx_HY_props) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	{
		tx_id := tx_id + ":tx-stat-1"
		tx_val_label := "Clicks"
		tx_child := tx_new_tx_HY_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_comp.tx_target, &tx_val_label, &tx_comp.V_clicks)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := tx_id + ":tx-stat-2"
		tx_val_label := "Doubled"
		tx_child := tx_new_tx_HY_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_comp.tx_target, &tx_val_label, &tx_comp.V_doubled)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_HY_props) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p>")
	{
		tx_id := tx_id + ":tx-stat-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_stat)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString("</p> <p>")
	{
		tx_id := tx_id + ":tx-stat-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_stat)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString("</p> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">click</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_search struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_q       string         `json:"q"`
	V_results []albums.Album `json:"-"`
}

func tx_new_tx_HY_search(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_search {
	tx_comp := &tx_HY_search{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_results = albums.Search(tx_comp.V_q)
	} else {
		tx_comp.V_results = albums.Search(tx_comp.V_q)
	}
	return tx_comp
}

func (tx_comp *tx_HY_search) tx_eh1(tx_ev_target_value string) {
	tx_comp.V_q = tx_ev_target_value
	tx_comp.V_results = albums.Search(tx_comp.V_q)
}

func (tx_comp *tx_HY_search) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var tx_ev_target_value string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("tx_ev_target_value")), &tx_ev_target_value)
			tx_comp.tx_eh1(tx_ev_target_value)
		}
	}

	for _, a := range tx_comp.V_results {
		_ = a

	}
	if len(tx_comp.V_results) == 0 {

	}
}

func (tx_comp *tx_HY_search) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <input type=\"search\" placeholder=\"Search albums…\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"input\"/> <table> <tbody>")

	for _, a := range tx_comp.V_results {
		_ = a
		tx_w.WriteString("<tr> <td>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(a.Title)))
		tx_w.WriteString("</td> <td>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(a.Artist)))
		tx_w.WriteString("</td> <td>")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(a.Year)))
		tx_w.WriteString("</td> </tr>")

	}
	tx_w.WriteString(" </tbody></table> ")
	if len(tx_comp.V_results) == 0 {
		tx_w.WriteString("<p>No albums match “")
		tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_q)))
		tx_w.WriteString("”.</p> ")

	}
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_slotcard struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_title *string `json:"-"`
}

func tx_new_tx_HY_slotcard(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, title *string) *tx_HY_slotcard {
	tx_comp := &tx_HY_slotcard{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_title = title
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_HY_slotcard) tx_render(tx_w *bytes.Buffer, tx_id string, tx_render_fill_ func()) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div style=\"border: 1px solid SlateGray; padding: 1rem; border-radius: 0.25rem\"> <strong>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_title)))
	tx_w.WriteString("</strong> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	}
	tx_w.WriteString(" </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_slotdemo struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_HY_slotdemo(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_slotdemo {
	tx_comp := &tx_HY_slotdemo{}
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

func (tx_comp *tx_HY_slotdemo) tx_compute(tx_id string) {
	{
		tx_id := tx_id + ":tx-slotcard-1"
		tx_val_title := "Card title"
		tx_child := tx_new_tx_HY_slotcard(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_comp.tx_target, &tx_val_title)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_HY_slotdemo) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	{
		tx_id := tx_id + ":tx-slotcard-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_slotcard)
		tx_child.tx_render(tx_w, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_slotcard_1_(tx_w)
		})
	}
	tx_w.WriteString(" ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

func (tx_comp *tx_HY_slotdemo) tx_render_fill_tx_HY_slotcard_1_(tx_w *bytes.Buffer) {
	tx_w.WriteString(" <p>Any markup here fills the &lt;slot&gt;.</p> ")
}

type tx_HY_snippet struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_name *string `json:"-"`
}

func tx_new_tx_HY_snippet(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, name *string) *tx_HY_snippet {
	tx_comp := &tx_HY_snippet{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_name = name
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_HY_snippet) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"> <span class=\"snippet-name\">components/")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_name)))
	tx_w.WriteString(".html</span> <div class=\"snippet-actions\"> <a href=\"https://github.com/gnituy18/tmplx/blob/master/tmplx.org/components/")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_name)))
	tx_w.WriteString(".html\" rel=\"noopener\" target=\"_blank\">source</a> <button class=\"copy-btn\" type=\"button\">copy</button> </div> </div> <div class=\"snippet-code\" data-snippet=\"")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_name)))
	tx_w.WriteString("\"></div> </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_stat struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_label *string `json:"-"`
	V_value *int    `json:"-"`
}

func tx_new_tx_HY_stat(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, label *string, value *int) *tx_HY_stat {
	tx_comp := &tx_HY_stat{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_label = label
	tx_comp.V_value = value
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_HY_stat) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <span>")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_label)))
	tx_w.WriteString(": ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_value)))
	tx_w.WriteString("</span> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_todo struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_list []string `json:"list"`
}

func tx_new_tx_HY_todo(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_todo {
	tx_comp := &tx_HY_todo{}
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

func (tx_comp *tx_HY_todo) add(item string) {
	tx_comp.V_list = append(tx_comp.V_list, item)
}

func (tx_comp *tx_HY_todo) remove(i int) {
	tx_comp.V_list = append(tx_comp.V_list[0:i], tx_comp.V_list[i+1:]...)
}

func (tx_comp *tx_HY_todo) tx_eh1(item string) {
	tx_comp.add(item)
}

func (tx_comp *tx_HY_todo) tx_eh2(i int) {
	tx_comp.remove(i)
}

func (tx_comp *tx_HY_todo) tx_compute(tx_id string) {
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

func (tx_comp *tx_HY_todo) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <form data-tx-trigger=\"")
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
	tx_w.WriteString(" </ol> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_triangle struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter int `json:"counter"`
}

func tx_new_tx_HY_triangle(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_triangle {
	tx_comp := &tx_HY_triangle{}
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

func (tx_comp *tx_HY_triangle) tx_eh1() {
	tx_comp.V_counter++
}

func (tx_comp *tx_HY_triangle) tx_compute(tx_id string) {
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

func (tx_comp *tx_HY_triangle) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div> <span> ")
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
	tx_w.WriteString(" ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_HY_types struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_count   int      `json:"count"`
	V_name    string   `json:"name"`
	V_enabled bool     `json:"enabled"`
	V_tags    []string `json:"tags"`
}

func tx_new_tx_HY_types(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_HY_types {
	tx_comp := &tx_HY_types{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_count = 3
		tx_comp.V_name = "tmplx"
		tx_comp.V_enabled = true
		tx_comp.V_tags = []string{"go", "html"}
	}
	return tx_comp
}

func (tx_comp *tx_HY_types) tx_eh1() {
	tx_comp.V_count++
}

func (tx_comp *tx_HY_types) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_HY_types) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <ul> <li>int: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count)))
	tx_w.WriteString("</li> <li>string: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_name)))
	tx_w.WriteString("</li> <li>bool: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_enabled)))
	tx_w.WriteString("</li> <li>slice: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_tags))))
	tx_w.WriteString(" items</li> </ul> <button data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">count++</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_SL_docs struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_SL_docs(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_SL_docs {
	tx_comp := &tx_SL_docs{}
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

func (tx_comp *tx_SL_docs) tx_compute() {
	{
		tx_id := "tx-nav-1"
		tx_child := tx_new_tx_HY_nav(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-todo-1"
			tx_child := tx_new_tx_HY_todo(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-types-1"
			tx_child := tx_new_tx_HY_types(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-derived-1"
			tx_child := tx_new_tx_HY_derived(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-4"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-counter-1"
			tx_child := tx_new_tx_HY_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-5"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-double-1"
			tx_child := tx_new_tx_HY_double(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-6"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-inputlive-1"
			tx_child := tx_new_tx_HY_inputlive(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-7"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-addn-1"
			tx_child := tx_new_tx_HY_addn(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-8"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-current-time-1"
			tx_child := tx_new_tx_HY_current_HY_time(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-9"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-cond-1"
			tx_child := tx_new_tx_HY_cond(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-10"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-triangle-1"
			tx_child := tx_new_tx_HY_triangle(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-11"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-condrows-1"
			tx_child := tx_new_tx_HY_condrows(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-12"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-greeting-1"
			tx_child := tx_new_tx_HY_greeting(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-13"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-props-1"
			tx_child := tx_new_tx_HY_props(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-14"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-callback-1"
			tx_child := tx_new_tx_HY_callback(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-example-wrapper-15"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-slotdemo-1"
			tx_child := tx_new_tx_HY_slotdemo(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
}

func (tx_comp *tx_SL_docs) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Docs | tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"/style.css\"/> <link rel=\"stylesheet\" href=\"/snippets.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/docs\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-nav-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_nav)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <div class=\"docs\"> <nav class=\"toc\"> <details> <summary>☰ Contents</summary> <ul> <li><a href=\"#introduction\">Introduction</a></li> <li><a href=\"#installing\">Installing</a></li> <li><a href=\"#quick-start\">Quick Start</a></li> <li><a href=\"#pages-and-routing\">Pages and Routing</a></li> <li> <a href=\"#tmplx-script\">tmplx Script</a> <ul> <li><a href=\"#imports\">Imports</a></li> <li><a href=\"#reserved-names\">Reserved Names</a></li> </ul> </li> <li> <a href=\"#expression-interpolation\">Expression Interpolation</a> </li> <li><a href=\"#state\">State</a></li> <li><a href=\"#derived\">Derived</a></li> <li> <a href=\"#event-handler\">Event Handlers</a> <ul> <li><a href=\"#event-properties\">Event Properties</a></li> </ul> </li> <li> <a href=\"#functions\">Functions</a> <ul> <li><a href=\"#captured-locals\">Captured Locals</a></li> </ul> </li> <li><a href=\"#init\">init()</a></li> <li><a href=\"#path-parameter\">Path Parameter</a></li> <li> <a href=\"#control-flow\">Control Flow</a> <ul> <li><a href=\"#conditionals\">Conditionals</a></li> <li><a href=\"#loops\">Loops</a></li> </ul> </li> <li><a href=\"#template\">&lt;template&gt;</a></li> <li><a href=\"#forms\">Forms</a></li> <li> <a href=\"#component\">Component</a> <ul> <li> <a href=\"#props\">Props</a> <ul> <li><a href=\"#callback-props\">Callback Props</a></li> </ul> </li> <li><a href=\"#slot\">&lt;slot&gt;</a></li> </ul> </li> <li><a href=\"#cli\">CLI</a></li> <li> Dev Tools <ul> <li><a href=\"#syntax-highlight\">Syntax Highlight</a></li> </ul> </li> </ul> </details> </nav> <main> <h2 id=\"introduction\">Introduction</h2> <p> tmplx is a framework for building full-stack web applications using only Go and HTML. Its goal is to make building web apps simple, intuitive, and fun again. It significantly reduces cognitive load by: </p> <ol> <li> <strong>keeping frontend and backend logic close together</strong> </li> <li> <strong>providing reactive UI updates driven by Go variables</strong> </li> <li><strong>requiring zero new syntax</strong></li> </ol> <p> Developing with tmplx feels like writing a more intuitive version of Go templates where the UI magically becomes reactive. </p> ")
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_1_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-01\"></div> </div> <p> You start by creating an HTML file. It can be a page or a reusable component, depending on where you place it. </p> <p> You use the <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> tag to embed Go code and make the page or component dynamic. tmplx uses a subset of Go syntax to provide reactive features like <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, and <a href=\"#event-handler\">event handler</a>. At the same time, because the script is valid Go, you can <strong>implement backend logic</strong>—such as database queries—directly in the template. </p> <p> tmplx compiles the HTML templates and embedded Go code into Go functions that render the HTML on the server and generate HTTP handlers for interactive events. On each interaction, the current state is sent to the server, which computes updates and returns both new HTML and the updated state. The result is server-rendered pages with lightweight client-side swapping (similar to <a href=\"https://htmx.org/\">htmx</a>). The interactivity plumbing is handled automatically by the tmplx compiler and runtime—you just implement the features. </p> <p> Most modern web applications separate the frontend and backend into different languages and teams. tmplx eliminates this split by letting you build the entire interactive application in a single language—Go. With this approach, the mental effort needed to track how data flows from the source to the UI is reduced to a minimum. The fewer transformations you perform on your data, the fewer bugs you introduce. </p> <h2 id=\"installing\">Installing</h2> <p>tmplx requires Go 1.25 or later.</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-02\"></div> </div> <p> This adds tmplx to your Go bin directory (usually $GOPATH/bin or $HOME/go/bin). Make sure that directory is in your PATH. </p> <p>After installation, verify it works:</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-03\"></div> </div> <h2 id=\"quick-start\">Quick Start</h2> <p>Get a tmplx app running in minutes.</p> <ol> <li> <p><strong>Create a project</strong></p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-04\"></div> </div> </li> <li> <p><strong>Add your first page (pages/index.html)</strong></p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-05\"></div> </div> </li> <li> <p><strong>Generate the Go code</strong></p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-06\"></div> </div> </li> <li> <p><strong>Create main.go to serve the app</strong></p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">go</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-07\"></div> </div> </li> <li> <p><strong>Run the server</strong></p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-08\"></div> </div> </li> </ol> <p> That&#39;s it! Open <a href=\"http://localhost:8080\">http://localhost:8080</a> and you now have a working interactive counter. </p> <h2 id=\"pages-and-routing\">Pages and Routing</h2> <p> A <strong>page</strong> is a standalone HTML file that has its own URL in your web app. </p> <p> All pages are placed in the <strong>pages</strong> directory. Default pages location is <code>./pages</code>. Change it with the <code>-pages-dir</code> flag: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">shell</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-09\"></div> </div> <p> tmplx uses <strong>filesystem-based routing</strong>. The route for a page is the relative path of the HTML file inside the <strong>pages</strong> directory, without the <code>.html</code> extension. For example: </p> <ul> <li><code>pages/index.html</code> → <code>/</code></li> <li><code>pages/about.html</code> → <code>/about</code></li> <li> <code>pages/admin/dashboard.html</code> → <code>/admin/dashboard</code> </li> </ul> <p> When the file is named <code>index.html</code>, it serves its directory&#39;s URL, which ends with a trailing slash. The page matches that URL <strong>exactly</strong> — it does not catch other paths under the directory — and a request without the trailing slash is redirected to it. </p> <ul> <li><code>pages/docs/index.html</code> → <code>/docs/</code></li> <li><code>pages/index/index.html</code> → <code>/index/</code></li> </ul> <p> A name and a directory are <strong>different URLs</strong>: <code>login.html</code> serves <code>/login</code> while <code>login/index.html</code> serves <code>/login/</code>, and both can exist in the same project. Two files that derive the same route cause compilation failure. </p> <p> The exact match is what <code>{$}</code> does: in Go&#39;s <code>net/http.ServeMux</code>, a pattern ending in <code>/</code> matches the <em>whole subtree</em>, so a bare <code>/docs/</code> pattern would serve the index page for every unmatched URL under it. tmplx therefore registers every <code>index.html</code> with <code>{$}</code> appended: <code>pages/docs/index.html</code> becomes the pattern <code>GET /docs/{$}</code>, which matches only <code>/docs/</code> itself. That string is the page&#39;s <strong>identity</strong>, so you will meet it wherever the page is referred to — in the <code>Routes()</code> patterns when you attach middleware, and in the event POST URLs in the network tab: </p> <ul> <li> <code>pages/docs/index.html</code> → pattern <code>GET /docs/{$}</code> </li> <li> <code>pages/index.html</code> → pattern <code>GET /{$}</code> </li> <li> <code>pages/docs.html</code> → pattern <code>GET /docs</code> (no <code>{$}</code>; a non-index route never ends in <code>/</code>) </li> </ul> <p> To add URL parameters (path wildcards), use curly braces  in directory or file names inside the pages directory. The name inside  must be a valid Go identifier. </p> <ul> <li> <code>pages/user/{user_id}.html</code> → <code>/user/{user_id}</code> </li> <li> <code>pages/blog/{year}/{slug}.html</code> → <code>/blog/{year}/{slug}</code> </li> </ul> <p> These patterns are compatible with Go&#39;s <code>net/http.ServeMux</code> (Go 1.22+). The parameter values are available in page initialisation through <code><a href=\"#path-parameter\">tx:path</a></code> comments. </p> <p> A <code>{name...}</code> wildcard matches the rest of the URL, slashes included — a <strong>catch-all</strong>. It serves every path under its directory that no more specific page matches: an exact page wins first, then the directory&#39;s <code>index.html</code>, then the catch-all. </p> <ul> <li> <code>pages/docs/{rest...}.html</code> → <code>/docs/{rest...}</code> </li> </ul> <p> tmplx compiles all pages into a single Go file you can import into your Go project. The pages directory can be outside your project, but keeping it inside is recommended. </p> <h2 id=\"tmplx-script\">tmplx Script</h2> <p> <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> is a special tag that you can add to your page or component to declare <a href=\"#state\">state</a>, <a href=\"#derived\">derived</a>, <a href=\"#event-handler\">event handlers</a>, <a href=\"#functions\">functions</a>, and the special <a href=\"#init\">init()</a> function to control your UI or add backend logic. </p> <p> Each page or component file can have exactly <strong>one</strong> tmplx script. Multiple scripts cause a compilation error. </p> <p> In pages, place it anywhere inside <code>&lt;head&gt;</code> or <code>&lt;body&gt;</code>. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-10\"></div> </div> <p>In components, place it at the root level.</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-11\"></div> </div> <h3 id=\"imports\">Imports</h3> <p> The script is plain Go, so you pull in packages with normal Go <code>import</code> declarations—single or grouped—at the top of the block. </p> <p> Imports resolve against your project&#39;s <code>go.mod</code>. tmplx walks up from the working directory to the nearest <code>go.mod</code> (see <a href=\"#cli\">CLI</a>) and type-checks the script against that module, so you can import: </p> <ul> <li>the <strong>standard library</strong>;</li> <li> any <strong>third-party module</strong> already in your <code>go.mod</code> (run <code>go get</code> first); </li> <li>your <strong>own packages</strong> within the module.</li> </ul> <p> An import that does not resolve fails compilation with <code>cannot import &lt;path&gt;</code>. Imported struct or named types can be used as <a href=\"#state\">state</a>. The <a href=\"/playground\">playground</a> resolves the standard library only. </p> <h3 id=\"reserved-names\">Reserved Names</h3> <p> The compiler reserves two naming patterns for its own use: </p> <ul> <li> Identifiers (variables, function names, parameter names) declared in the tmplx script cannot start with <code>tx_</code>. </li> <li> HTML attributes starting with <code>tx-</code> are reserved for tmplx directives (<code>tx-if</code>, <code>tx-for</code>, <code>tx-on*</code>, <code>tx-action</code>, ...). Do not introduce your own <code>tx-</code> attributes. </li> </ul> <h2 id=\"expression-interpolation\">Expression Interpolation</h2> <p> Use curly braces <code>{}</code> to insert <a href=\"https://go.dev/ref/spec#Expressions\">Go expressions</a> into HTML. Expressions are allowed only in: </p> <ul> <li><strong>text nodes</strong></li> <li><strong>attribute values</strong></li> </ul> <p>Placing expressions anywhere else causes a parsing error.</p> <p>\n        tmplx converts expression results to strings using\n        <code><a href=\"https://pkg.go.dev/fmt#Sprint\">fmt.Sprint</a></code>. The output is <strong>HTML-escaped</strong> in both\n        <strong>text nodes</strong> and <strong>attribute values</strong> to\n        prevent cross-site scripting (XSS)—an interpolated value cannot\n        inject markup or break out of its attribute.\n      </p> <p> Expressions run on the server every time the page loads or a component re-renders after an event. Avoid side effects in expressions, such as database queries or heavy computations, because they execute on every render. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-12\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-13\"></div> </div> <p>\n        Add the <code>tx-ignore</code> attribute to an element to disable\n        expression interpolation in that element&#39;s attributes and its direct\n        text children. Descendant elements are still processed normally.\n      </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-14\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-15\"></div> </div> <h2 id=\"state\">State</h2> <p> <strong>State</strong> is the mutable data that describes a component&#39;s current condition. </p> <p> Declaring state works like declaring variables in Go&#39;s package scope. If you provide no initial value, the state starts with the zero value for its type. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-16\"></div> </div> <p>To set an initial value, use the <code>=</code> operator.</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-17\"></div> </div> <p>Although the syntax follows valid Go code, these rules apply:</p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong> The type must be JSON-compatible. </strong> </li> </ol> <p> The 1st rule is enforced by the compiler. General JSON-compatibility is not checked at compile time (for now)—an interface or channel type compiles and then fails at runtime. The one exception is a <strong>function-typed</strong> state, which is rejected at compile time; use a <a href=\"#callback-props\">callback prop</a> for a function input. </p> <h3>Some invalid state declarations:</h3> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-18\"></div> </div> <h3>Some valid state declarations:</h3> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-19\"></div> </div> <p>State can hold any JSON-compatible Go type. A few of them, live:</p> ")
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_2_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">types.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"types\"></div> </div> <p> The tmplx script cannot contain Go <code>type</code> declarations. To use your own struct or named types as state, declare them in a regular Go package and import them—then reference the imported type in the <code>var</code> declaration. </p> <h2 id=\"derived\">Derived</h2> A <strong>derived</strong> is a <strong>read-only</strong> value that is automatically calculated from states. It updates whenever those states change. <p> Declaring a derived works the same way as declaring package-level variables in Go. When the right-hand side of the declaration <strong>references existing state or other derived values</strong>, it is treated as a derived value. </p> <p> Derived values follow most of the same rules as regular state variables, but with some differences: </p> <ol> <li><strong>Only one identifier per declaration.</strong></li> <li> <strong>Derived values cannot be modified directly in event handlers, though they may be read.</strong> </li> </ol> ")
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_3_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">derived.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"derived\"></div> </div> <p> A derived can read any number of states—here a slice is joined into a class string: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-21\"></div> </div> <h2 id=\"event-handler\">Event Handlers</h2> <p> An <strong>event handler</strong> binds a DOM event to Go code that runs on the server. Add an attribute that starts with <code>tx-on</code> followed by the event name (<code>tx-onclick</code>, <code>tx-oninput</code>, <code>tx-onchange</code>, …); its value is a body of Go statements. </p> <p> The body is plain Go—mutate state directly, call a <a href=\"#functions\">function</a>, or both. The simplest handler just assigns to a state variable; no function is required. </p> ")
	{
		tx_id := "tx-example-wrapper-4"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_4_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">counter.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"counter\"></div> </div> <p> Any valid Go statement works in the body—here a compound assignment doubles a value on each click: </p> ")
	{
		tx_id := "tx-example-wrapper-5"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_5_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">double.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"double\"></div> </div> <h3 id=\"event-properties\">Event Properties</h3> <p> For events fired on a form control, you can read values directly from the DOM event object inside the handler body. The runtime injects them into the request automatically. </p> <ul> <li> <code>event.target.value</code> resolves to the target’s current string value. It is available on the value-bearing events: <code>input</code>, <code>change</code>, <code>keydown</code>, <code>keyup</code>, <code>keypress</code>, <code>blur</code>, <code>focus</code>, <code>focusin</code>, <code>focusout</code>, and <code>search</code> (i.e. <code>tx-oninput</code>, <code>tx-onkeyup</code>, <code>tx-onblur</code>, …). </li> <li> On <code>keydown</code> and <code>keypress</code> the value is the one <strong>before</strong> the keystroke is applied—use <code>keyup</code> or <code>input</code> for the value after. </li> <li> Reading <code>event.target.value</code> on any other event is a compile error. </li> </ul> ")
	{
		tx_id := "tx-example-wrapper-6"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_6_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">inputlive.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"inputlive\"></div> </div> <h2 id=\"functions\">Functions</h2> <p> A <strong>function</strong> is a standalone, reusable Go function declared in the tmplx script as <code>func name(...) { ... }</code>. It is a <strong>separate idea</strong> from an event handler: a handler is a <code>tx-on*</code> binding, while a function is plain logic you can call from many places—an event handler, a <a href=\"#forms\">form</a>, <a href=\"#init\">init()</a>, another function, or an expression. </p> <p> Functions are not one-to-one with events. One function can back many bindings, and a binding may call no function at all (inline statements, as above) or several. Only <code>tx-on*</code> and <code>tx-action</code> bindings compile to HTTP endpoints—a function on its own does not. </p> <p> Functions take parameters and may return values. Below, a single <code>addNum</code> backs ten buttons; each click passes a different argument: </p> ")
	{
		tx_id := "tx-example-wrapper-7"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_7_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">addn.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"addn\"></div> </div> <p> A state-mutating function runs on the server through the handler or form that calls it. A function used inside an <a href=\"#expression-interpolation\">expression</a> or <a href=\"#derived\">derived</a> value runs on <strong>every render</strong>, so keep functions in that position pure—no state mutation or write-side effects. </p> <h3 id=\"captured-locals\">Captured Locals</h3> <p> Any local variable bound by an enclosing <code>tx-for</code> init clause, <code>tx-for</code> range form, or <code>tx-if</code>/ <code>tx-else-if</code> init form is <strong>automatically captured</strong> by handlers in the subtree. Just reference the local by name; the framework figures out which values cross the wire and decodes them with their inferred Go type. </p> <ul> <li> The captured local can appear anywhere in the handler body—a function argument, an array index, an expression operand. Whatever is valid Go. </li> <li> Types are recovered from the binding’s context (the surrounding state declarations and imports), so you don’t need to annotate. </li> <li> State, derived, and prop variables are <strong>not</strong> captured this way—the handler already reads them from the component’s saved state. </li> </ul> <p> In the <code>addn</code> example above, <code>i</code> is captured from the <code>tx-for</code> init clause and decoded as <code>int</code> on the server. Captures from <code>tx-if</code>/<code>tx-else-if</code> init forms work the same way: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-24\"></div> </div> <p> Both <code>f</code> (from <code>tx-for</code>) and <code>n</code> (from the <code>tx-if</code> init) are captured automatically. </p> <h2 id=\"init\">init()</h2> <p> <code>init()</code> is a special function that runs automatically the first time a page or component is rendered. For pages, it runs on every GET request. For components, it runs when the component has no saved state yet (for example, the first time it appears on the page, or the first time a new <code>tx-for</code> iteration produces it). After that, subsequent renders reuse the saved state and skip <code>init()</code>. </p> ")
	{
		tx_id := "tx-example-wrapper-8"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_8_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">current-time.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"current-time\"></div> </div> <p> Another common use case is to initialize one state from another state without turning the second variable into a derived state. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-28\"></div> </div> <h2 id=\"path-parameter\">Path Parameters</h2> <p> When a page route contains a wildcard (see <a href=\"#pages-and-routing\">Pages and Routing</a>), you can pull the captured value into a state variable by annotating the declaration with a <code>//tx:path</code> comment. </p> <p>Rules:</p> <ul> <li> The comment must sit directly above the <code>var</code> line (Go doc-comment position). </li> <li> The value after <code>tx:path</code> is the wildcard name from the route pattern. </li> <li> The variable must be declared as <code>string</code>. No initial value is allowed—the captured string is the initial value. </li> <li> Only <a href=\"#pages-and-routing\">pages</a> support <code>tx:path</code>; components cannot declare path-bound state. </li> </ul> <p> The captured value is assigned <strong>before</strong> <a href=\"#init\"><code>init()</code></a> runs, so <code>init()</code> can use it to populate other state (for example, by loading a record from the database). </p> <p> <strong>Single parameter.</strong> For a route <code>pages/blog/post/{post_id}.html</code>: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-29\"></div> </div> <p> <strong>Multiple parameters.</strong> Each wildcard gets its own declaration. For a route <code>pages/blog/{year}/{slug}.html</code>: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-30\"></div> </div> <p> After initialization, the variable behaves like any other state: it&#39;s serialized, sent to the server on events, and can be reassigned from handlers (though reassigning it does not change the URL). </p> <p> Try it live: <a href=\"/hello/world\">/hello/world</a> · <a href=\"/hello/tmplx\">/hello/tmplx</a>. This page binds the <code>{name}</code> URL segment with <code>//tx:path name</code>: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">pages/hello/{name}.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"pathparam\"></div> </div> <h2 id=\"control-flow\">Control Flow</h2> <p> tmplx avoids new custom syntax for conditionals and loops. It embeds control flow directly into HTML attributes, similar to Vue.js and <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h3 id=\"conditionals\">Conditionals</h3> <p> To conditionally render elements, use the <code>tx-if</code>, <code>tx-else-if</code>, and <code>tx-else</code> attributes on the desired tags. The values for <code>tx-if</code> and <code>tx-else-if</code> can be any valid Go expression that would fit in an <code>if</code> or <code>else if</code> statement. The <code>tx-else</code> attribute needs no value. </p> ")
	{
		tx_id := "tx-example-wrapper-9"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_9_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">cond.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"cond\"></div> </div> <p> You can declare <strong>local variables</strong> and handle errors exactly as you would in regular Go code. Local variables declared in conditionals are available to the element and its descendants, just like in Go. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-32\"></div> </div> <p> A conditional group consists of <strong>consecutive sibling nodes</strong> that share the same parent. Disconnected nodes are not treated as part of the same group. A standalone <code>tx-else-if</code> or <code>tx-else</code> without a preceding <code>tx-if</code> will cause a compilation error. </p> <h3 id=\"loops\">Loops</h3> <p> To repeat elements, use the <code>tx-for</code> attribute. Its value can be any valid Go <code>for</code> statement, including <strong>classic for</strong> or <strong>range for</strong>. </p> <p> Local variables declared in the loop are available to the element and all of its descendants, just like in Go. </p> <p> Always add a <code>tx-key</code> attribute with a unique value for each item. This gives the compiler a unique identifier for the node during updates. </p> ")
	{
		tx_id := "tx-example-wrapper-10"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_10_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">triangle.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"triangle\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-34\"></div> </div> <p> To branch per item, do <strong>not</strong> put <code>tx-if</code> and <code>tx-for</code> on the same element—the condition is compiled outside the loop and cannot see the loop variable, which is a compile error. Put the conditional on an element <strong>inside</strong> the loop so it can read the bound variable: </p> ")
	{
		tx_id := "tx-example-wrapper-11"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_11_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">condrows.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"condrows\"></div> </div> <h2 id=\"template\">&lt;template&gt;</h2> <p> The <code>&lt;template&gt;</code> tag is a non-rendering container that lets you apply control flow attributes (<code>tx-if</code>, <code>tx-else-if</code>, <code>tx-else</code>, or <code>tx-for</code>) to a group of elements at once. </p> <p> The <code>&lt;template&gt;</code> itself is removed from the output; only its children are rendered (or not, depending on the control flow). </p> <p> You can nest <code>&lt;template&gt;</code> tags and combine them with other control flow attributes on child elements. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-35\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-36\"></div> </div> <h2 id=\"forms\">Forms</h2> <p> Attach a handler to a <code>&lt;form&gt;</code> with <code>tx-action</code>. When the form is submitted, tmplx cancels the default submission, collects every named form element, and calls the handler on the server. </p> <p> The value of <code>tx-action</code> must be the name of a function declared in the tmplx script. Each form element&#39;s <code>name</code> attribute must match a parameter name on that function; unnamed elements are ignored. </p> ")
	{
		tx_id := "tx-example-wrapper-12"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_12_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">greeting.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"greeting\"></div> </div> <p> Values are JSON-decoded into each parameter&#39;s Go type, so the parameter type is what determines how the string is parsed. The runtime serializes form elements by input type: </p> <ul> <li> <code>text</code>, <code>email</code>, <code>password</code>, <code>textarea</code>, <code>select</code>, etc.—sent as a JSON string. Decode into <code>string</code>. </li> <li> <code>number</code>, <code>range</code>—sent as the raw numeric value, or <code>null</code> when empty. Decode into a numeric type or pointer. </li> <li> <code>checkbox</code>—sent as <code>true</code> or <code>false</code>. Decode into <code>bool</code>. </li> <li> <code>radio</code>—only the checked radio in a group is sent (using its shared <code>name</code>). Decode into <code>string</code>. </li> </ul> <p> Because submission goes through a full server round-trip, use native HTML validation (<code>required</code>, <code>minlength</code>, <code>pattern</code>, ...) to catch client-side errors before the request is sent. For richer live-updating inputs, combine tmplx with a client-side library like <a href=\"https://alpinejs.dev/\">Alpine.js</a>. </p> <h2 id=\"component\">Component</h2> <p> Components are reusable UI building blocks that encapsulate HTML, state, and behavior. </p> <p> Create a component by placing an <code>.html</code> file in the <code>components</code> directory (default: <code>./components</code>). tmplx automatically registers it as a custom element with the tag name <code>tx-</code> followed by the relative path (without the <code>.html</code> extension), with directory separators replaced by <code>-</code>. </p> <p> Filenames and directory names may contain only <code>a-z</code>, <code>0-9</code>, <code>-</code>, and <code>_</code>. Uppercase letters are rejected. </p> <p>Examples:</p> <ul> <li> <code>components/button.html</code> → <code>&lt;tx-button&gt;</code> </li> <li> <code>components/user/card.html</code> → <code>&lt;tx-user-card&gt;</code> </li> <li> <code>components/todo/list.html</code> → <code>&lt;tx-todo-list&gt;</code> </li> </ul> <p> Components can contain their own <code>&lt;script type=&#34;text/tmplx&#34;&gt;</code> for local state and logic, and can be used in pages or nested inside other components. </p> <h3 id=\"props\">Props</h3> <p> Props are inputs the parent passes to a child component. Inside the child, a prop is declared like a state variable, but with a <code>//tx:prop</code> doc comment. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-38\"></div> </div> <p>Rules:</p> <ul> <li> The <code>//tx:prop</code> comment must sit directly above the <code>var</code> line. </li> <li> Prop names must be <strong>lowercase</strong>. HTML lowercases attribute names, so a camelCase prop name would never match the attribute the parent writes. </li> <li> An initial value (e.g. <code>= 0</code>) becomes the <strong>default</strong> used when the parent omits the attribute. </li> <li> Props are <strong>read-only</strong> inside the child. Event handlers can read them but cannot assign to them. Derived values referencing a prop recompute automatically when the prop changes. </li> <li> Pages cannot declare props—only components can. </li> </ul> <h4>Passing props</h4> <p> Prop attribute values on the parent are parsed as <strong>Go expressions</strong>, not as plain strings. Pass a literal by writing the literal directly; pass a parent variable by its name. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-39\"></div> </div> <p> The expression is re-evaluated whenever the parent re-renders, so the child stays in sync with the parent&#39;s state automatically. </p> ")
	{
		tx_id := "tx-example-wrapper-13"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_13_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">stat.html (the component)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"stat\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">props.html (using it)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"props\"></div> </div> <h4 id=\"callback-props\">Callback Props</h4> <p> A <strong>callback prop</strong> lets a child notify the parent when something happens. It is just a prop whose type is a <strong>function</strong>: declare it with <code>//tx:prop</code> and a function type. With no default the parent <strong>must</strong> supply an implementation (a required prop); give it a function-literal default to make the parent override optional. </p> <p>In the child, call it from a handler the same way you call a function:</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-40\"></div> </div> <p> In the parent, pass the <strong>bare name</strong> of a tmplx-script function as the attribute whose key matches the child&#39;s prop name: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-41\"></div> </div> <p> When the child calls <code>onselect(42)</code>, the parent&#39;s <code>pick</code> runs on the server with that argument and the parent re-renders. A callback call can be mixed freely with other statements in the same handler—for example <code>tx-onclick=&#34;count++; onselect(42)&#34;</code>. </p> <p> To make the override optional, give the prop a function-literal default. The parent may then omit the attribute and the child falls back to its own implementation: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-42\"></div> </div> <p> A runnable version: <code>&lt;tx-incbtn&gt;</code> takes a <code>label</code> and an <code>onpress</code> callback; the parent passes its own <code>add</code> function and counts the presses. </p> ")
	{
		tx_id := "tx-example-wrapper-14"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_14_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">incbtn.html (the component)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"incbtn\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">callback.html (using it)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"callback\"></div> </div> <h3 id=\"slot\">&lt;slot&gt;</h3> <p> A <code>&lt;slot&gt;</code> marks a place in a component&#39;s template where the parent can inject content. Slots are how components stay composable: the child decides the shape, the parent fills in the details. </p> <h4>Declaring slots in a component</h4> <p> Each slot is either the <strong>default slot</strong> (no <code>name</code>) or a <strong>named slot</strong>. A component may have at most one default slot, and named slots must be unique. Slots cannot be nested inside other slots. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-43\"></div> </div> <p> Content placed inside <code>&lt;slot&gt;...&lt;/slot&gt;</code> is <strong>fallback content</strong>—it renders only when the parent does not fill that slot. </p> <h4>Filling slots from the parent</h4> <p> Put fill content directly inside the component tag. Use the <code>slot</code> attribute on a child element to target a named slot; everything else becomes the default fill. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-44\"></div> </div> <p> Only the <strong>direct children</strong> of the component tag are considered when matching slots—a <code>slot</code> attribute on a deeply nested element has no effect. </p> ")
	{
		tx_id := "tx-example-wrapper-15"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_15_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">slotcard.html (the component)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"slotcard\"></div> </div> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">slotdemo.html (using it)</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"slotdemo\"></div> </div> <h4>Scope: fills use the parent&#39;s state</h4> <p> This is the most important rule. The content you pass into a slot is still <strong>parent code</strong>: expressions, event handlers, and directives inside a fill see the parent&#39;s state, derived, and prop variables—not the child&#39;s. </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">tmplx</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-45\"></div> </div> <p> Here <code>user</code> and <code>logout</code> are defined on the page that uses <code>&lt;tx-card&gt;</code>, not inside the card component. When the button is clicked the page&#39;s handler runs and the fill re-renders against the page&#39;s updated state. </p> <h4>Live example</h4> <p> The docs site uses a simple <code>&lt;tx-example-wrapper&gt;</code> component with a single default slot to frame every live demo on this page. The component is just: </p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">example-wrapper.html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"example-wrapper\"></div> </div> <p>And callers wrap any demo with it:</p> <div class=\"snippet\"> <div class=\"snippet-bar\"><span class=\"snippet-name\">html</span> <div class=\"snippet-actions\"><button class=\"copy-btn\" type=\"button\">copy</button></div> </div> <div class=\"snippet-code\" data-snippet=\"docs-47\"></div> </div> <h2 id=\"cli\">CLI</h2> <p> Running <code>tmplx</code> inside any directory of your Go module walks up to the nearest <code>go.mod</code> and uses that as the project root. All path flags default relative to that root. </p> <table> <thead> <tr> <th>Flag</th> <th>Default</th> <th>Description</th> </tr> </thead> <tbody> <tr> <td><code>-pages-dir</code></td> <td><code>./pages</code></td> <td>Directory containing pages.</td> </tr> <tr> <td><code>-components-dir</code></td> <td><code>./components</code></td> <td>Directory containing reusable components.</td> </tr> <tr> <td><code>-output-file</code></td> <td><code>./routes.go</code></td> <td>Path to the generated Go file.</td> </tr> <tr> <td><code>-package-name</code></td> <td><code>main</code></td> <td>Package name for the generated Go code.</td> </tr> <tr> <td><code>-handler-prefix</code></td> <td><code>/tx/</code></td> <td>URL path prefix for generated event handler routes.</td> </tr> </tbody> </table> <h2 id=\"syntax-highlight\">Syntax Highlight</h2> <a href=\"https://github.com/gnituy18/tmplx.nvim\">Neovim Plugin</a> </main> </div> <script src=\"/snippets.js\"></script> </body></html>")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_1_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-todo-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_todo)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_10_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-triangle-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_triangle)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_11_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-condrows-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_condrows)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_12_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-greeting-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_greeting)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_13_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-props-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_props)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_14_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-callback-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_callback)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_15_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-slotdemo-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_slotdemo)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_2_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-types-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_types)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_3_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-derived-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_derived)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_4_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_counter)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_5_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-double-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_double)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_6_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-inputlive-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_inputlive)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_7_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-addn-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_addn)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_8_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-current-time-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_current_HY_time)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL_docs) tx_render_fill_tx_HY_example_HY_wrapper_9_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-cond-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_cond)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

type tx_SL_examples_SL_state struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_count int    `json:"count"`
	V_label string `json:"label"`
	V_flag  bool   `json:"flag"`
}

func tx_new_tx_SL_examples_SL_state(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_SL_examples_SL_state {
	tx_comp := &tx_SL_examples_SL_state{}
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

func (tx_comp *tx_SL_examples_SL_state) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
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

type tx_SL_hello_SL__LB_name_RB_ struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_name string `json:"-"`
}

func tx_new_tx_SL_hello_SL__LB_name_RB_(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, name string) *tx_SL_hello_SL__LB_name_RB_ {
	tx_comp := &tx_SL_hello_SL__LB_name_RB_{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_name = name
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_SL_hello_SL__LB_name_RB_) tx_compute() {
	{
		tx_id := "tx-nav-1"
		tx_child := tx_new_tx_HY_nav(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_SL_hello_SL__LB_name_RB_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Hello | tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"icon\" type=\"image/svg+xml\" href=\"/logo.svg\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"/style.css\"/>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/hello/{name}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-nav-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_nav)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <main> <h1>Hello, ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_name)))
	tx_w2.WriteString("!</h1> <p> The name above came from the URL path segment, bound with <code>//tx:path</code>. <a href=\"/docs#path-parameter\">Path Parameters docs</a> </p> </main> </body></html>")
}

type tx_SL__EX_ struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_SL__EX_(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_SL__EX_ {
	tx_comp := &tx_SL__EX_{}
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

func (tx_comp *tx_SL__EX_) tx_compute() {
	{
		tx_id := "tx-nav-1"
		tx_child := tx_new_tx_HY_nav(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-search-1"
			tx_child := tx_new_tx_HY_search(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-snippet-1"
		tx_val_name := "search"
		tx_child := tx_new_tx_HY_snippet(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_name)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-counter-1"
			tx_child := tx_new_tx_HY_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-snippet-2"
		tx_val_name := "counter"
		tx_child := tx_new_tx_HY_snippet(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_name)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_new_tx_HY_example_HY_wrapper(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-todo-1"
			tx_child := tx_new_tx_HY_todo(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
	{
		tx_id := "tx-snippet-3"
		tx_val_name := "todo"
		tx_child := tx_new_tx_HY_snippet(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_name)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_SL__EX_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>tmplx — Write Go in HTML</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <meta name=\"description\" content=\"tmplx compiles HTML files with embedded Go into a single net/http server. State is a var, handlers run on the server, and there is no JavaScript to write.\"/> <link rel=\"icon\" type=\"image/svg+xml\" href=\"/logo.svg\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"/style.css\"/> <link rel=\"stylesheet\" href=\"/snippets.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/{$}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-nav-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_nav)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <main> <header class=\"hero\"> <div class=\"hero-title\"> <img src=\"/logo.svg\" alt=\"tmplx logo\" width=\"96\" height=\"96\"/> <h1>tmplx</h1> </div> <h2 class=\"hero-tagline\">Write Go in HTML</h2> <p class=\"hero-pitch\"> A page is one HTML file with a Go script block. State is a <code>var</code>, handlers are Go statements in attributes, and every handler runs on the server. tmplx compiles it all into a single <code>routes.go</code> on <code>net/http</code> — there is no JavaScript to write. </p> <div class=\"hero-actions\"> <a class=\"btn btn-primary\" href=\"/docs\">Get started</a> <a class=\"btn\" href=\"/playground\">Playground</a> <a class=\"btn\" href=\"https://github.com/gnituy18/tmplx\">GitHub</a> </div> </header> <section class=\"section\"> <h2>A query in your HTML</h2> <p class=\"muted\"> Type below. Every keystroke posts the state up, runs <code>albums.Search(q)</code> on the server, and morphs the rows that come back. <code>results</code> is derived from <code>q</code>, so it recomputes whenever <code>q</code> changes. </p> ")
	{
		tx_id := "tx-example-wrapper-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_1_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-snippet-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_snippet)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <p class=\"muted\"> <code>albums.Search</code> is an ordinary Go function in <a href=\"https://github.com/gnituy18/tmplx/blob/master/tmplx.org/albums/albums.go\" rel=\"noopener\" target=\"_blank\">an ordinary Go package</a> — an in-memory table here, your real database in production. </p> </section> <section class=\"section\"> <h2>The tmplx way</h2> <ul class=\"way\"> <li> <strong>State is a var.</strong> Declare it in the script block, render it with <code>{ }</code>. When a handler changes it, every place that reads it updates. </li> <li> <strong>Handlers are server code.</strong> <code>tx-onclick=&#34;counter++&#34;</code> is Go, not JavaScript. Call anything in your module — no fetch, no API routes, no JSON layer. </li> <li> <strong>The server stays stateless.</strong> State travels with the page: each event sends it up, HTML comes back down, and the DOM morphs in place. </li> <li> <strong>Components are HTML files.</strong> <code>components/card.html</code> becomes <code>&lt;tx-card&gt;</code>, with props, slots, and callbacks. </li> </ul> </section> <section class=\"section\"> <h2>Demos</h2> <h3>Counter</h3> <p class=\"muted\">State, rendered and mutated, in five lines.</p> ")
	{
		tx_id := "tx-example-wrapper-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_2_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-snippet-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_snippet)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <h3>To Do</h3> <p class=\"muted\"> A form posts straight to a Go function; click an item to remove it. </p> ")
	{
		tx_id := "tx-example-wrapper-3"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_example_HY_wrapper)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_HY_example_HY_wrapper_3_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-snippet-3"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_snippet)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </section> <section class=\"section\"> <h2>From file to server</h2> <ol class=\"steps\"> <li> Install the compiler: <code>go install github.com/gnituy18/tmplx@latest</code> </li> <li> Write pages in <code>pages/</code> and components in <code>components/</code>. </li> <li> Run <code>tmplx</code> — it type-checks every file and generates <code>routes.go</code>. Then <code>go run .</code> </li> </ol> <div class=\"hero-actions\"> <a class=\"btn btn-primary\" href=\"/docs\">Read the docs</a> </div> </section> </main> <script src=\"/snippets.js\"></script> </body></html>")
}

func (tx_comp *tx_SL__EX_) tx_render_fill_tx_HY_example_HY_wrapper_1_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-search-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_search)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL__EX_) tx_render_fill_tx_HY_example_HY_wrapper_2_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_counter)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

func (tx_comp *tx_SL__EX_) tx_render_fill_tx_HY_example_HY_wrapper_3_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-todo-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_todo)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

type tx_SL_playground struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_SL_playground(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_SL_playground {
	tx_comp := &tx_SL_playground{}
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

func (tx_comp *tx_SL_playground) tx_compute() {
	{
		tx_id := "tx-nav-1"
		tx_child := tx_new_tx_HY_nav(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_SL_playground) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!-- prettier-ignore --><!DOCTYPE html><html lang=\"en\"><head> <title>Playground | tmplx</title> <meta charset=\"UTF-8\"/> <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"/> <link rel=\"icon\" type=\"image/svg+xml\" href=\"/logo.svg\"/> <link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/modern-normalize@3.0.1/modern-normalize.min.css\"/> <link rel=\"stylesheet\" href=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/codemirror.min.css\"/> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/codemirror.min.js\"></script> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/mode/xml/xml.min.js\"></script> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/mode/javascript/javascript.min.js\"></script> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/mode/css/css.min.js\"></script> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/mode/htmlmixed/htmlmixed.min.js\"></script> <script src=\"https://cdnjs.cloudflare.com/ajax/libs/codemirror/5.65.16/mode/go/go.min.js\"></script> <link rel=\"stylesheet\" href=\"/style.css\"/> <link rel=\"stylesheet\" href=\"/snippets.css\"/> <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/playground\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body class=\"pg\"> ")
	{
		tx_id := "tx-nav-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_HY_nav)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <main class=\"pg-main\"> <div class=\"pg-bar\"> <h1>Playground</h1> <button id=\"compile\" class=\"btn-compile\" type=\"button\">Compile</button> <span class=\"note\">Your page compiles on the server as <code>pages/index.html</code> — type-checks only, never runs your code. Standard-library imports only.</span> </div> <div class=\"pg-wrap\"> <div class=\"pg-col\"> <h3>pages/index.html — tmplx source</h3> <div class=\"pg-box\"><textarea id=\"src\" spellcheck=\"false\">&lt;!DOCTYPE html&gt;\n&lt;html&gt;\n&lt;head&gt;\n  &lt;title&gt;Counter&lt;/title&gt;\n  &lt;script type=&#34;text/tmplx&#34;&gt;\n    var counter int\n  &lt;/script&gt;\n&lt;/head&gt;\n&lt;body&gt;\n  &lt;button tx-onclick=&#34;counter--&#34;&gt;-&lt;/button&gt;\n  &lt;span&gt;{ counter }&lt;/span&gt;\n  &lt;button tx-onclick=&#34;counter++&#34;&gt;+&lt;/button&gt;\n&lt;/body&gt;\n&lt;/html&gt;</textarea></div> </div> <div class=\"pg-col\"> <h3>routes.go — generated Go</h3> <div id=\"out\" class=\"pg-box\"></div> <pre id=\"errors\" class=\"pg-errors\" style=\"display: none\"></pre> </div> </div> </main> <script>\n    (function () {\n      var inCM = CodeMirror.fromTextArea(document.getElementById(\"src\"), {\n        mode: { name: \"htmlmixed\", scriptTypes: [{ matches: /text\\/tmplx/, mode: \"go\" }] },\n        theme: \"tmplx\",\n        lineNumbers: true,\n        tabSize: 2,\n        indentUnit: 2,\n        lineWrapping: false,\n      });\n\n      var outCM = CodeMirror(document.getElementById(\"out\"), {\n        mode: \"go\",\n        theme: \"tmplx\",\n        lineNumbers: true,\n        readOnly: true,\n        value: \"\",\n      });\n\n      var outEl = document.getElementById(\"out\");\n      var errEl = document.getElementById(\"errors\");\n      var btn = document.getElementById(\"compile\");\n\n      function showCode(code) {\n        errEl.style.display = \"none\";\n        outEl.style.display = \"\";\n        outCM.setValue(code);\n        outCM.refresh();\n      }\n      function showErrors(lines) {\n        outEl.style.display = \"none\";\n        errEl.style.display = \"block\";\n        errEl.textContent = lines.join(\"\\n\");\n      }\n\n      async function compile() {\n        var old = btn.textContent;\n        btn.textContent = \"Compiling...\";\n        btn.disabled = true;\n        try {\n          var res = await fetch(\"/playground/compile\", { method: \"POST\", body: inCM.getValue() });\n          var data = await res.json();\n          if (data.diagnostics && data.diagnostics.length) showErrors(data.diagnostics);\n          else showCode(data.code || \"\");\n        } catch (e) {\n          showErrors([\"request failed: \" + e]);\n        } finally {\n          btn.textContent = old;\n          btn.disabled = false;\n        }\n      }\n\n      btn.addEventListener(\"click\", compile);\n      compile();\n    })();\n  </script> </body></html>")
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
		tx_comp := tx_new_tx_SL_docs(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/examples/state":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_SL_examples_SL_state(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/hello/{name}":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_SL_hello_SL__LB_name_RB_(tx_prev, tx_next, tx_trigger, tx_trigger_handler, "")
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/{$}":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_SL__EX_(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/playground":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_SL_playground(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
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
		tx_comp := tx_new_tx_HY_addn(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-callback":
		tx_comp := tx_new_tx_HY_callback(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-cond":
		tx_comp := tx_new_tx_HY_cond(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-condrows":
		tx_comp := tx_new_tx_HY_condrows(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-counter":
		tx_comp := tx_new_tx_HY_counter(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-current-time":
		tx_comp := tx_new_tx_HY_current_HY_time(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-derived":
		tx_comp := tx_new_tx_HY_derived(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-double":
		tx_comp := tx_new_tx_HY_double(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-double-state":
		tx_comp := tx_new_tx_HY_double_HY_state(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-example-wrapper":
		tx_comp := tx_new_tx_HY_example_HY_wrapper(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target, nil)
	case "tx-greeting":
		tx_comp := tx_new_tx_HY_greeting(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-inputlive":
		tx_comp := tx_new_tx_HY_inputlive(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-nav":
		tx_comp := tx_new_tx_HY_nav(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target)
	case "tx-props":
		tx_comp := tx_new_tx_HY_props(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-search":
		tx_comp := tx_new_tx_HY_search(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-slotdemo":
		tx_comp := tx_new_tx_HY_slotdemo(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-todo":
		tx_comp := tx_new_tx_HY_todo(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-triangle":
		tx_comp := tx_new_tx_HY_triangle(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-types":
		tx_comp := tx_new_tx_HY_types(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
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
			tx_comp := tx_new_tx_SL_docs(nil, tx_next, "", "")
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
		Pattern: "GET /examples/state",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_SL_examples_SL_state(nil, tx_next, "", "")
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
		Pattern: "GET /hello/{name}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_SL_hello_SL__LB_name_RB_(nil, tx_next, "", "", tx_r.PathValue("name"))
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
		Pattern: "GET /{$}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_SL__EX_(nil, tx_next, "", "")
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
		Pattern: "GET /playground",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_SL_playground(nil, tx_next, "", "")
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
		Pattern: "POST /tx/tx-derived/eh1",
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
		Pattern: "POST /tx/tx-incbtn/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-inputlive/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-props/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-search/eh1",
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
	{
		Pattern: "POST /tx/tx-types/eh1",
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

  const underTarget = (k, t) => k === t || (k.startsWith(t) && ':@;'.includes(k[t.length]))

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
    morphChildren(a, a.firstChild, b.firstChild, null, null)
  }

  const morphChildren = (parent, a, b, aEnd, bEnd) => {
    while (a !== aEnd && b !== bEnd) {
      const a2 = a.nextSibling, b2 = b.nextSibling
      morph(a, b)
      a = a2
      b = b2
    }
    while (a !== aEnd) {
      const n = a.nextSibling
      a.remove()
      a = n
    }
    while (b !== bEnd) {
      parent.insertBefore(b.cloneNode(true), aEnd)
      b = b.nextSibling
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
      if (target === 'page' || underTarget(key, target)) {
        params.append(key, JSON.stringify(state[key]))
      }
    }

    const res = await fetch("/tx/" + fun, { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString() })
    const html = await res.text()

    if (target === 'page') {
      const doc = new DOMParser().parseFromString(html, 'text/html')
      const txState = doc.getElementById('tx-saved')
      if (txState) state = JSON.parse(txState.textContent)
      morph(document.documentElement, doc.documentElement)
      return
    }

    const comp = document.createElement('body')
    comp.innerHTML = html
    const txState = comp.querySelector("#tx-saved")
    if (!txState) return
    const next = JSON.parse(txState.textContent)
    for (const k in state) {
      if (underTarget(k, target)) delete state[k]
    }
    Object.assign(state, next)

    const respStart = findComment(comp, 'tx:' + target)
    const respEnd = findComment(comp, 'tx:' + target + '_e')
    const docStart = findComment(document.documentElement, 'tx:' + target)
    const docEnd = findComment(document.documentElement, 'tx:' + target + '_e')
    if (!respStart || !respEnd || !docStart || !docEnd) return

    morphChildren(docStart.parentNode, docStart.nextSibling, respStart.nextSibling, docEnd, respEnd)
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
          if (typeof e.target.value === 'string') {
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
            if (el.name && (el.type !== 'radio' || el.checked)) {
              let v
              if (el.type === 'checkbox') v = el.checked ? 'true' : 'false'
              else if (el.type === 'number' || el.type === 'range') v = el.value === '' ? 'null' : el.value
              else v = JSON.stringify(el.value)
              params.append(el.name, v)
            }
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
