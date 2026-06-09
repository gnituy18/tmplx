package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
)

type tx_H_badge struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_name  *string `json:"-"`
	V_ticks int     `json:"ticks"`
}

func tx_new_tx_H_badge(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, name *string) *tx_H_badge {
	tx_comp := &tx_H_badge{}
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

func (tx_comp *tx_H_badge) tx_eh1() {
	tx_comp.V_ticks++
}

func (tx_comp *tx_H_badge) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_badge) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <span> <span data-test=\"badge-text\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_name)))
	tx_w.WriteString(": ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_ticks)))
	tx_w.WriteString("</span> <button data-test=\"badge-tick\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">+1</button> </span> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_box struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_H_box(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_box {
	tx_comp := &tx_H_box{}
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

func (tx_comp *tx_H_box) tx_render(tx_w *bytes.Buffer, tx_id string, tx_render_fill_ func()) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div data-test=\"box\"> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	} else {
		tx_w.WriteString("fallback content")
	}
	tx_w.WriteString(" </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_button struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_label   *string          `json:"-"`
	V_amount  *int             `json:"-"`
	V_clicked func(amount int) `json:"-"`
}

func tx_new_tx_H_button(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, label *string, amount *int, clicked func(amount int)) *tx_H_button {
	tx_comp := &tx_H_button{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_label = label
	tx_comp.V_amount = amount
	tx_comp.V_clicked = clicked
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_button) tx_eh1() {
	tx_comp.V_clicked(*tx_comp.V_amount)
}

func (tx_comp *tx_H_button) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_button) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <button data-test=\"btn\" data-tx-trigger=\"")
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

type tx_H_calc struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_compute func(x int) int `json:"-"`
}

func tx_new_tx_H_calc(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, compute func(x int) int) *tx_H_calc {
	tx_comp := &tx_H_calc{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_compute = compute
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_calc) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div data-test=\"calc\">result: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_compute(5))))
	tx_w.WriteString("</div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_compound struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_local  int    `json:"local"`
	V_notify func() `json:"-"`
}

func tx_new_tx_H_compound(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, notify func()) *tx_H_compound {
	tx_comp := &tx_H_compound{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_notify = notify
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_compound) tx_eh1() {
	tx_comp.V_local++
	tx_comp.V_notify()
}

func (tx_comp *tx_H_compound) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_compound) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div data-test=\"compound-local\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_local)))
	tx_w.WriteString("</div> <button data-test=\"compound-btn\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">go</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_counter struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_clicks int `json:"clicks"`
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
	tx_comp.V_clicks++
}

func (tx_comp *tx_H_counter) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_counter) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <button data-test=\"cbtn\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">click me (")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_clicks)))
	tx_w.WriteString(")</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_defaulter struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_n     int     `json:"n"`
	V_label *string `json:"-"`
	V_bump  func()  `json:"-"`
}

func tx_new_tx_H_defaulter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, label *string, bump func()) *tx_H_defaulter {
	tx_comp := &tx_H_defaulter{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	if label != nil {
		tx_comp.V_label = label
	} else {
		val_label := "default-label"
		tx_comp.V_label = &val_label
	}
	if bump != nil {
		tx_comp.V_bump = bump
	} else {
		tx_comp.V_bump = func() {
			tx_comp.V_n++
		}
	}
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_n = 0
	}
	return tx_comp
}

func (tx_comp *tx_H_defaulter) tx_eh1() {
	tx_comp.V_bump()
}

func (tx_comp *tx_H_defaulter) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_defaulter) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p data-test=\"def-label\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_label)))
	tx_w.WriteString("</p> <p data-test=\"def-n\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_n)))
	tx_w.WriteString("</p> <button data-test=\"def-btn\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">bump</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_init_H_derived struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_a int `json:"a"`
	V_b int `json:"-"`
}

func tx_new_tx_H_init_H_derived(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_init_H_derived {
	tx_comp := &tx_H_init_H_derived{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_b = tx_comp.V_a + 1
	} else {
		tx_comp.V_a = 1
		tx_comp.V_b = tx_comp.V_a + 1
		tx_comp.V_a += tx_comp.V_b
		tx_comp.V_b = tx_comp.V_a + 1
	}
	return tx_comp
}

func (tx_comp *tx_H_init_H_derived) tx_eh1() {
	tx_comp.V_a++
	tx_comp.V_b = tx_comp.V_a + 1
}

func (tx_comp *tx_H_init_H_derived) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_init_H_derived) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <p data-test=\"comp-a\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_a)))
	tx_w.WriteString("</p> <p data-test=\"comp-b\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_b)))
	tx_w.WriteString("</p> <button data-test=\"comp-plus\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">+</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_panel struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_heading *string `json:"-"`
}

func tx_new_tx_H_panel(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, heading *string) *tx_H_panel {
	tx_comp := &tx_H_panel{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_heading = heading
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_H_panel) tx_compute(tx_id string) {
	{
		tx_id := tx_id + ":tx-counter-1"
		tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_H_panel) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <section data-test=\"panel\"> <h2 data-test=\"panel-heading\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_heading)))
	tx_w.WriteString("</h2> ")
	{
		tx_id := tx_id + ":tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" </section> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_seeded struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_n int `json:"n"`
}

func tx_new_tx_H_seeded(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string) *tx_H_seeded {
	tx_comp := &tx_H_seeded{}
	tx_comp.tx_target = tx_target
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get(tx_id)
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_n = 42
	}
	return tx_comp
}

func (tx_comp *tx_H_seeded) tx_eh1() {
	tx_comp.V_n++
}

func (tx_comp *tx_H_seeded) tx_compute(tx_id string) {
	if tx_id == tx_comp.tx_trigger {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_H_seeded) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <span data-test=\"seeded-n\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_n)))
	tx_w.WriteString("</span> <button data-test=\"seeded-plus\" data-tx-trigger=\"")
	fmt.Fprint(tx_w, tx_id)
	tx_w.WriteString("\" data-tx-target=\"")
	fmt.Fprint(tx_w, tx_comp.tx_target)
	tx_w.WriteString("\" data-tx-eh1-on=\"click\">+1</button> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_slot_H_card struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_title *string `json:"-"`
}

func tx_new_tx_H_slot_H_card(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, title *string) *tx_H_slot_H_card {
	tx_comp := &tx_H_slot_H_card{}
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

func (tx_comp *tx_H_slot_H_card) tx_render(tx_w *bytes.Buffer, tx_id string, tx_render_fill_ func(), tx_render_fill_footer func()) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <div> <h3 data-test=\"card-title\">")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(*tx_comp.V_title)))
	tx_w.WriteString("</h3> ")
	if tx_render_fill_ != nil {
		tx_render_fill_()
	}
	tx_w.WriteString(" <hr/> <small>")
	if tx_render_fill_footer != nil {
		tx_render_fill_footer()
	}
	tx_w.WriteString("</small> </div> ")
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id+"_e")
		tx_w.WriteString("-->")
	}
}

type tx_H_stat struct {
	tx_target          string         `json:"-"`
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_label *string `json:"-"`
	V_value *int    `json:"-"`
}

func tx_new_tx_H_stat(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, tx_id string, tx_target string, label *string, value *int) *tx_H_stat {
	tx_comp := &tx_H_stat{}
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

func (tx_comp *tx_H_stat) tx_render(tx_w *bytes.Buffer, tx_id string) {
	if tx_comp.tx_target == tx_id {
		tx_w.WriteString("<!--tx:")
		fmt.Fprint(tx_w, tx_id)
		tx_w.WriteString("-->")
	}
	tx_w.WriteString(" <span data-test=\"stat\">")
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

type tx_S_attr_H_interp struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_who string `json:"who"`
	V_cls string `json:"cls"`
}

func tx_new_tx_S_attr_H_interp(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_attr_H_interp {
	tx_comp := &tx_S_attr_H_interp{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_who = "alice"
		tx_comp.V_cls = "active"
	}
	return tx_comp
}

func (tx_comp *tx_S_attr_H_interp) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>attribute interpolation</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/attr-interp\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- interpolation inside an attribute value --> <a data-test=\"link\" href=\"/echo/")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_who)))
	tx_w2.WriteString("\">go</a> <div data-test=\"dyn-class\" class=\"box ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_cls)))
	tx_w2.WriteString("\">x</div> </body></html>")
}

type tx_S_badge struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_badge(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_badge {
	tx_comp := &tx_S_badge{}
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

func (tx_comp *tx_S_badge) tx_compute() {
	{
		tx_id := "tx-badge-1"
		tx_val_name := "A"
		tx_child := tx_new_tx_H_badge(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_name)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-badge-2"
		tx_val_name := "B"
		tx_child := tx_new_tx_H_badge(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_name)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_badge) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Badges</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/badge\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-badge-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_badge)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-badge-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_badge)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_compound struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_pageHits int `json:"pageHits"`
}

func tx_new_tx_S_compound(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_compound {
	tx_comp := &tx_S_compound{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_pageHits = 0
	}
	return tx_comp
}

func (tx_comp *tx_S_compound) bumpPage() {
	tx_comp.V_pageHits++
}

func (tx_comp *tx_S_compound) tx_compute() {
	{
		tx_id := "tx-compound-1"
		tx_child := tx_new_tx_H_compound(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", tx_comp.bumpPage)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_compound) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>compound</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/compound\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"compound-page-hits\">page hits: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_pageHits)))
	tx_w2.WriteString("</p> ")
	{
		tx_id := "tx-compound-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_compound)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_conditionals struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_n    int  `json:"n"`
	V_show bool `json:"show"`
}

func tx_new_tx_S_conditionals(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_conditionals {
	tx_comp := &tx_S_conditionals{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_n = 0
		tx_comp.V_show = true
	}
	return tx_comp
}

func (tx_comp *tx_S_conditionals) tx_eh1() {
	tx_comp.V_show = !tx_comp.V_show
}

func (tx_comp *tx_S_conditionals) tx_eh2() {
	tx_comp.V_n++
}

func (tx_comp *tx_S_conditionals) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		case "eh2":
			tx_comp.tx_eh2()
		}
	}
	if tx_comp.V_show {

	}
	if tx_comp.V_n > 5 {
	} else if tx_comp.V_n > 0 {
	} else {

	}
	if tx_comp.V_show {
		{
			tx_id := "tx-counter-1"
			tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}

	}
}

func (tx_comp *tx_S_conditionals) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>conditionals</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/conditionals\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- tx-if with no else: the node exists only while the condition holds --> ")
	if tx_comp.V_show {
		tx_w2.WriteString("<p data-test=\"cond-show\">visible</p> ")

	}
	tx_w2.WriteString("<button data-test=\"cond-toggle\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">toggle</button> <!-- if / else-if / else chain: exactly one branch renders --> ")
	if tx_comp.V_n > 5 {
		tx_w2.WriteString("<p data-test=\"bucket\">big</p> ")
	} else if tx_comp.V_n > 0 {
		tx_w2.WriteString("<p data-test=\"bucket\">small</p> ")
	} else {
		tx_w2.WriteString("<p data-test=\"bucket\">none</p> ")

	}
	tx_w2.WriteString("<button data-test=\"cond-inc\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh2-on=\"click\">+</button> <!-- a component rendered conditionally (mounted/unmounted with the branch) --> ")
	if tx_comp.V_show {
		tx_w2.WriteString("<div> ")
		{
			tx_id := "tx-counter-1"
			tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
			tx_child.tx_render(tx_w2, tx_id)
		}
		tx_w2.WriteString(" </div> ")

	}
	tx_w2.WriteString("</body></html>")
}

type tx_S_counter_H_comp struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_counter_H_comp(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_counter_H_comp {
	tx_comp := &tx_S_counter_H_comp{}
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

func (tx_comp *tx_S_counter_H_comp) tx_compute() {
	{
		tx_id := "tx-counter-1"
		tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-counter-2"
		tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_counter_H_comp) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Counter Components</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/counter-comp\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-counter-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_counter struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter       int `json:"counter"`
	V_counterDouble int `json:"-"`
}

func tx_new_tx_S_counter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_counter {
	tx_comp := &tx_S_counter{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_counterDouble = tx_comp.V_counter * 2
	} else {
		tx_comp.V_counter = 0
		tx_comp.V_counterDouble = tx_comp.V_counter * 2
	}
	return tx_comp
}

func (tx_comp *tx_S_counter) addAmount(amount int) {
	tx_comp.V_counter += amount
	tx_comp.V_counterDouble = tx_comp.V_counter * 2
}

func (tx_comp *tx_S_counter) tx_eh1() {
	tx_comp.V_counter--
	tx_comp.V_counterDouble = tx_comp.V_counter * 2
}

func (tx_comp *tx_S_counter) tx_eh2() {
	tx_comp.V_counter = 0
	tx_comp.V_counterDouble = tx_comp.V_counter * 2
}

func (tx_comp *tx_S_counter) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		case "eh2":
			tx_comp.tx_eh2()
		}
	}
	{
		tx_id := "tx-stat-1"
		tx_val_label := "Live"
		tx_child := tx_new_tx_H_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_comp.V_counter)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-stat-2"
		tx_val_label := "Doubled"
		tx_child := tx_new_tx_H_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_comp.V_counterDouble)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-button-1"
		tx_val_label := "+3"
		tx_val_amount := 3
		tx_child := tx_new_tx_H_button(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_val_amount, tx_comp.addAmount)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-button-2"
		tx_val_label := "+5"
		tx_val_amount := 5
		tx_child := tx_new_tx_H_button(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_val_amount, tx_comp.addAmount)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	if tx_comp.V_counter > 0 {
	} else if tx_comp.V_counter == 0 {
	} else {

	}
}

func (tx_comp *tx_S_counter) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Counter</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/counter\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <h1 data-test=\"counter-h1\">Counter: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w2.WriteString(" (doubled: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counterDouble)))
	tx_w2.WriteString(")</h1> ")
	{
		tx_id := "tx-stat-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_stat)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-stat-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_stat)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-button-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_button)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-button-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_button)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <button data-test=\"minus\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">-1</button> <button data-test=\"reset\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh2-on=\"click\">reset</button> ")
	if tx_comp.V_counter > 0 {
		tx_w2.WriteString("<p data-test=\"sign\">positive</p> ")
	} else if tx_comp.V_counter == 0 {
		tx_w2.WriteString("<p data-test=\"sign\">zero</p> ")
	} else {
		tx_w2.WriteString("<p data-test=\"sign\">negative</p> ")

	}
	tx_w2.WriteString("</body></html>")
}

type tx_S_defaulter struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_defaulter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_defaulter {
	tx_comp := &tx_S_defaulter{}
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

func (tx_comp *tx_S_defaulter) tx_compute() {
	{
		tx_id := "tx-defaulter-1"
		tx_child := tx_new_tx_H_defaulter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id, nil, nil)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_defaulter) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>defaulter</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/defaulter\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-defaulter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_defaulter)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_derived_H_chain struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_a int `json:"a"`
	V_b int `json:"-"`
	V_c int `json:"-"`
}

func tx_new_tx_S_derived_H_chain(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_derived_H_chain {
	tx_comp := &tx_S_derived_H_chain{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_b = tx_comp.V_a + 1
		tx_comp.V_c = tx_comp.V_b * 10
	} else {
		tx_comp.V_a = 1
		tx_comp.V_b = tx_comp.V_a + 1
		tx_comp.V_c = tx_comp.V_b * 10
	}
	return tx_comp
}

func (tx_comp *tx_S_derived_H_chain) tx_eh1() {
	tx_comp.V_a++
	tx_comp.V_b = tx_comp.V_a + 1
	tx_comp.V_c = tx_comp.V_b * 10
}

func (tx_comp *tx_S_derived_H_chain) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_S_derived_H_chain) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>derived chain</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/derived-chain\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- b derives from a, c derives from b: a change to a must cascade to both --> <p data-test=\"dc-a\">a=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_a)))
	tx_w2.WriteString("</p> <p data-test=\"dc-b\">b=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_b)))
	tx_w2.WriteString("</p> <p data-test=\"dc-c\">c=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_c)))
	tx_w2.WriteString("</p> <button data-test=\"dc-inc\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">+</button> </body></html>")
}

type tx_S_echo_S__L_msg_R_ struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_msg string `json:"-"`
}

func tx_new_tx_S_echo_S__L_msg_R_(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string, msg string) *tx_S_echo_S__L_msg_R_ {
	tx_comp := &tx_S_echo_S__L_msg_R_{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_comp.V_msg = msg
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	}
	return tx_comp
}

func (tx_comp *tx_S_echo_S__L_msg_R_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>echo path var</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/echo/{msg}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- path variable bound from the URL segment {msg} --> <p data-test=\"echo-msg\">msg: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_msg)))
	tx_w2.WriteString("</p> </body></html>")
}

type tx_S_expr_H_prop struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_base int `json:"base"`
}

func tx_new_tx_S_expr_H_prop(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_expr_H_prop {
	tx_comp := &tx_S_expr_H_prop{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_base = 4
	}
	return tx_comp
}

func (tx_comp *tx_S_expr_H_prop) tx_eh1() {
	tx_comp.V_base++
}

func (tx_comp *tx_S_expr_H_prop) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	{
		tx_id := "tx-stat-1"
		tx_val_label := "sum"
		tx_val_value := tx_comp.V_base*2 + 1
		tx_child := tx_new_tx_H_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_val_value)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_expr_H_prop) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"/><title>expr prop</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/expr-prop\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- a component value-prop given a compound expression, not just an ident/literal --> ")
	{
		tx_id := "tx-stat-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_stat)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" <button data-test=\"ep-inc\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">+</button> </body></html>")
}

type tx_S_filter struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_nums []int `json:"nums"`
}

func tx_new_tx_S_filter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_filter {
	tx_comp := &tx_S_filter{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_nums = []int{1, 2, 3, 4, 5, 6}
	}
	return tx_comp
}

func (tx_comp *tx_S_filter) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>conditional rows (tx-if inside tx-for)</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/filter\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- NOTE: tx-for + tx-if on the SAME element is NOT supported (the tx-if can't\n       see the loop var; recorded as a gap). The supported pattern is tx-if on an\n       INNER element, which can reference the enclosing loop variable. --> <ul> ")

	for _, n := range tx_comp.V_nums {
		_ = n
		tx_w2.WriteString("<li> ")
		if n > 3 {
			tx_w2.WriteString("<b data-test=\"big\">")
			tx_w2.WriteString(html.EscapeString(fmt.Sprint(n)))
			tx_w2.WriteString("-big</b> ")
		} else {
			tx_w2.WriteString("<i data-test=\"small\">")
			tx_w2.WriteString(html.EscapeString(fmt.Sprint(n)))
			tx_w2.WriteString("-small</i> ")

		}
		tx_w2.WriteString("</li>")

	}
	tx_w2.WriteString(" </ul> </body></html>")
}

type tx_S_funcprop_H_return struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_funcprop_H_return(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_funcprop_H_return {
	tx_comp := &tx_S_funcprop_H_return{}
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

func (tx_comp *tx_S_funcprop_H_return) triple(x int) int {
	return x * 3
}

func (tx_comp *tx_S_funcprop_H_return) tx_compute() {
	{
		tx_id := "tx-calc-1"
		tx_child := tx_new_tx_H_calc(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", tx_comp.triple)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_funcprop_H_return) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>func-prop return</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/funcprop-return\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-calc-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_calc)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_importer struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_name string `json:"name"`
}

func tx_new_tx_S_importer(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_importer {
	tx_comp := &tx_S_importer{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_name = "hello"
	}
	return tx_comp
}

func (tx_comp *tx_S_importer) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"/><title>import</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/importer\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"upper\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(strings.ToUpper(tx_comp.V_name))))
	tx_w2.WriteString("</p> </body></html>")
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

func (tx_comp *tx_S__EX_) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>tmplx test examples</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/{$}\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <h1>tmplx test examples</h1> <p>Ordered simple → complex, mirroring the compiler feature flow.</p> <h2>text &amp; state</h2> <ul> <li><a href=\"/text\">text</a> — interpolation, expressions, HTML-escaping</li> <li><a href=\"/state-types\">state-types</a> — bool / float / slice state</li> <li><a href=\"/derived-chain\">derived-chain</a> — transitive derived (a → b → c)</li> <li><a href=\"/init-derived-page\">init-derived-page</a> — init() reads derived (page)</li> <li><a href=\"/init-derived-comp\">init-derived-comp</a> — init() reads derived (component)</li> </ul> <h2>handlers</h2> <ul> <li><a href=\"/counter\">counter</a> — handlers, func-prop, derived, if/else-if/else</li> <li><a href=\"/show-counter\">show-counter</a> — handler + pure func in an expression</li> <li><a href=\"/input\">input</a> — input event / event.target.value</li> </ul> <h2>conditions &amp; loops</h2> <ul> <li><a href=\"/conditionals\">conditionals</a> — tx-if / else-if / else, conditional component</li> <li><a href=\"/loops\">loops</a> — index+value, empty, range-over-int</li> <li><a href=\"/nested-loop\">nested-loop</a> — tx-for inside tx-for</li> <li><a href=\"/map-loop\">map-loop</a> — range over a map</li> <li><a href=\"/loop-comps\">loop-comps</a> — keyed components in a loop</li> </ul> <h2>components &amp; props</h2> <ul> <li><a href=\"/counter-comp\">counter-comp</a> — two independent sealed instances</li> <li><a href=\"/badge\">badge</a> — propful component, prop rewiring + morph</li> <li><a href=\"/seeded\">seeded</a> — component init() seeds state</li> <li><a href=\"/nested-comp\">nested-comp</a> — component rendering a component</li> <li><a href=\"/defaulter\">defaulter</a> — prop &amp; func-prop defaults</li> </ul> <h2>func-props &amp; slots</h2> <ul> <li><a href=\"/compound\">compound</a> — child calls a parent func-prop</li> <li><a href=\"/funcprop-return\">funcprop-return</a> — func-prop returning a value</li> <li><a href=\"/slots\">slots</a> — default + named slots</li> <li><a href=\"/slot-fallback\">slot-fallback</a> — slot fallback vs. fill</li> </ul> <h2>captured locals, path vars, swap</h2> <ul> <li><a href=\"/todos\">todos</a> — captured loop locals (removeTodo(i)); tx-action add (known gap)</li> <li><a href=\"/echo/hello\">echo/hello</a> — path variable bound from the URL segment</li> <li><a href=\"/state-survives\">state-survives</a> — page state survives a component swap</li> </ul> <h2>known gaps</h2> <ul> <li><a href=\"/attr-interp\">attr-interp</a> — attribute interpolation (NOT supported; renders literally)</li> </ul> </body></html>")
}

type tx_S_init_H_derived_H_comp struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_init_H_derived_H_comp(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_init_H_derived_H_comp {
	tx_comp := &tx_S_init_H_derived_H_comp{}
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

func (tx_comp *tx_S_init_H_derived_H_comp) tx_compute() {
	{
		tx_id := "tx-init-derived-1"
		tx_child := tx_new_tx_H_init_H_derived(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_init_H_derived_H_comp) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>init reads derived (comp)</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/init-derived-comp\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-init-derived-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_init_H_derived)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_init_H_derived_H_page struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_a int `json:"a"`
	V_b int `json:"-"`
	V_c int `json:"c"`
	V_d int `json:"-"`
}

func tx_new_tx_S_init_H_derived_H_page(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_init_H_derived_H_page {
	tx_comp := &tx_S_init_H_derived_H_page{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_b = tx_comp.V_a + 1
		tx_comp.V_d = tx_comp.V_c + 1
	} else {
		tx_comp.V_a = 1
		tx_comp.V_b = tx_comp.V_a + 1
		tx_comp.V_c = 10
		tx_comp.V_d = tx_comp.V_c + 1
		tx_comp.V_a += tx_comp.V_b
		tx_comp.V_b = tx_comp.V_a + 1
	}
	return tx_comp
}

func (tx_comp *tx_S_init_H_derived_H_page) tx_eh1() {
	tx_comp.V_a++
	tx_comp.V_b = tx_comp.V_a + 1
}

func (tx_comp *tx_S_init_H_derived_H_page) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
}

func (tx_comp *tx_S_init_H_derived_H_page) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>init reads derived (page)</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/init-derived-page\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"a\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_a)))
	tx_w2.WriteString("</p> <p data-test=\"b\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_b)))
	tx_w2.WriteString("</p> <p data-test=\"c\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_c)))
	tx_w2.WriteString("</p> <p data-test=\"d\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_d)))
	tx_w2.WriteString("</p> <button data-test=\"plus\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">+</button> </body></html>")
}

type tx_S_input struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_inputValue string `json:"inputValue"`
}

func tx_new_tx_S_input(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_input {
	tx_comp := &tx_S_input{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_inputValue = ""
	}
	return tx_comp
}

func (tx_comp *tx_S_input) tx_eh1(tx_ev_target_value string) {
	tx_comp.V_inputValue = tx_ev_target_value
}

func (tx_comp *tx_S_input) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var tx_ev_target_value string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("tx_ev_target_value")), &tx_ev_target_value)
			tx_comp.tx_eh1(tx_ev_target_value)
		}
	}
}

func (tx_comp *tx_S_input) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Input</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/input\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <input data-test=\"text-input\" type=\"text\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"input\"/> <p data-test=\"echo\">You typed: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_inputValue)))
	tx_w2.WriteString(" (")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_inputValue))))
	tx_w2.WriteString(" chars)</p> </body></html>")
}

type tx_S_loop_H_comps struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_labels []string `json:"labels"`
}

func tx_new_tx_S_loop_H_comps(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_loop_H_comps {
	tx_comp := &tx_S_loop_H_comps{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_labels = []string{"red", "green", "blue"}
	}
	return tx_comp
}

func (tx_comp *tx_S_loop_H_comps) tx_compute() {

	for _, lbl := range tx_comp.V_labels {
		_ = lbl
		{
			tx_id := ";" + fmt.Sprint(lbl) + ":tx-counter-1"
			tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}

	}
}

func (tx_comp *tx_S_loop_H_comps) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Loop Components</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/loop-comps\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")

	for _, lbl := range tx_comp.V_labels {
		_ = lbl
		tx_w2.WriteString("<div> ")
		{
			tx_id := ";" + fmt.Sprint(lbl) + ":tx-counter-1"
			tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
			tx_child.tx_render(tx_w2, tx_id)
		}
		tx_w2.WriteString(" </div>")

	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_loops struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_items []string `json:"items"`
	V_empty []string `json:"empty"`
}

func tx_new_tx_S_loops(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_loops {
	tx_comp := &tx_S_loops{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_items = []string{"a", "b", "c"}
		tx_comp.V_empty = []string{}
	}
	return tx_comp
}

func (tx_comp *tx_S_loops) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>loops</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/loops\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- loop exposing both index and value --> <ul> ")

	for i, v := range tx_comp.V_items {
		_ = i
		_ = v
		tx_w2.WriteString("<li data-test=\"item\">")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(i)))
		tx_w2.WriteString(":")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(v)))
		tx_w2.WriteString("</li>")

	}
	tx_w2.WriteString(" </ul> <!-- empty slice renders zero rows --> <ul> ")

	for _, v := range tx_comp.V_empty {
		_ = v
		tx_w2.WriteString("<li data-test=\"empty-item\">")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(v)))
		tx_w2.WriteString("</li>")

	}
	tx_w2.WriteString(" </ul> <!-- range over an integer (Go 1.22 form): yields 0..n-1 --> ")

	for k := range 4 {
		_ = k
		tx_w2.WriteString("<span data-test=\"num\">")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(k)))
		tx_w2.WriteString("</span>")

	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_loopvar_H_prop struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_labels []string `json:"labels"`
}

func tx_new_tx_S_loopvar_H_prop(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_loopvar_H_prop {
	tx_comp := &tx_S_loopvar_H_prop{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_labels = []string{"x", "y", "z"}
	}
	return tx_comp
}

func (tx_comp *tx_S_loopvar_H_prop) tx_compute() {

	for i, l := range tx_comp.V_labels {
		_ = i
		_ = l
		{
			tx_id := ";" + fmt.Sprint(i) + ":tx-stat-1"
			tx_val_label := l
			tx_val_value := i
			tx_child := tx_new_tx_H_stat(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_label, &tx_val_value)
			tx_comp.tx_next[tx_id] = tx_child
		}

	}
}

func (tx_comp *tx_S_loopvar_H_prop) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>loop var as prop</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/loopvar-prop\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- tx-for on a component, passing the loop index/value as props --> ")

	for i, l := range tx_comp.V_labels {
		_ = i
		_ = l
		{
			tx_id := ";" + fmt.Sprint(i) + ":tx-stat-1"
			tx_child := tx_comp.tx_next[tx_id].(*tx_H_stat)
			tx_child.tx_render(tx_w2, tx_id)
		}

	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_map_H_loop struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_scores map[string]int `json:"scores"`
}

func tx_new_tx_S_map_H_loop(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_map_H_loop {
	tx_comp := &tx_S_map_H_loop{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_scores = map[string]int{"a": 1, "b": 2, "c": 3}
	}
	return tx_comp
}

func (tx_comp *tx_S_map_H_loop) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>map loop</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/map-loop\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- ranging over a map. Map order is random in Go and conditions/loops run\n       twice (compute + render); keys keep child identity stable, but visual\n       order may not be deterministic. Probing what actually happens. --> ")

	for k, v := range tx_comp.V_scores {
		_ = k
		_ = v
		tx_w2.WriteString("<span data-test=\"score\">")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(k)))
		tx_w2.WriteString("=")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(v)))
		tx_w2.WriteString(";</span>")

	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_nested_H_comp struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_nested_H_comp(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_nested_H_comp {
	tx_comp := &tx_S_nested_H_comp{}
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

func (tx_comp *tx_S_nested_H_comp) tx_compute() {
	{
		tx_id := "tx-panel-1"
		tx_val_heading := "Alpha"
		tx_child := tx_new_tx_H_panel(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_heading)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-panel-2"
		tx_val_heading := "Beta"
		tx_child := tx_new_tx_H_panel(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_heading)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_nested_H_comp) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>nested components</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/nested-comp\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- a component (panel) that itself renders another component (counter);\n       two instances must keep independent state --> ")
	{
		tx_id := "tx-panel-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_panel)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" ")
	{
		tx_id := "tx-panel-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_panel)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_nested_H_loop struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_rows []string `json:"rows"`
	V_cols []int    `json:"cols"`
}

func tx_new_tx_S_nested_H_loop(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_nested_H_loop {
	tx_comp := &tx_S_nested_H_loop{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_rows = []string{"x", "y"}
		tx_comp.V_cols = []int{1, 2, 3}
	}
	return tx_comp
}

func (tx_comp *tx_S_nested_H_loop) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>nested loop</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/nested-loop\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- a tx-for nested inside another tx-for (cartesian product) --> ")

	for _, r := range tx_comp.V_rows {
		_ = r
		tx_w2.WriteString("<div data-test=\"row\"> ")

		for _, c := range tx_comp.V_cols {
			_ = c
			tx_w2.WriteString("<span data-test=\"cell\">")
			tx_w2.WriteString(html.EscapeString(fmt.Sprint(r)))
			tx_w2.WriteString(html.EscapeString(fmt.Sprint(c)))
			tx_w2.WriteString("</span>")

		}
		tx_w2.WriteString(" </div>")

	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_render_H_edges struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_f     float64 `json:"f"`
	V_g     float64 `json:"g"`
	V_neg   int     `json:"neg"`
	V_emoji string  `json:"emoji"`
	V_zero  int     `json:"zero"`
}

func tx_new_tx_S_render_H_edges(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_render_H_edges {
	tx_comp := &tx_S_render_H_edges{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_f = 1.0
		tx_comp.V_g = 0.5
		tx_comp.V_neg = -42
		tx_comp.V_emoji = "hello world"
		tx_comp.V_zero = 0
	}
	return tx_comp
}

func (tx_comp *tx_S_render_H_edges) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"/><title>render edges</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/render-edges\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"re-f\">[")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_f)))
	tx_w2.WriteString("]</p> <p data-test=\"re-g\">[")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_g)))
	tx_w2.WriteString("]</p> <p data-test=\"re-neg\">[")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_neg)))
	tx_w2.WriteString("]</p> <p data-test=\"re-emoji\">[")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_emoji)))
	tx_w2.WriteString("]</p> <p data-test=\"re-zero\">[")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_zero)))
	tx_w2.WriteString("]</p> </body></html>")
}

type tx_S_seeded struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_seeded(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_seeded {
	tx_comp := &tx_S_seeded{}
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

func (tx_comp *tx_S_seeded) tx_compute() {
	{
		tx_id := "tx-seeded-1"
		tx_child := tx_new_tx_H_seeded(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_seeded) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>seeded comp</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/seeded\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-seeded-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_seeded)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_show_H_counter struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_counter       int `json:"counter"`
	V_counterDouble int `json:"-"`
}

func tx_new_tx_S_show_H_counter(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_show_H_counter {
	tx_comp := &tx_S_show_H_counter{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
		tx_comp.V_counterDouble = tx_comp.V_counter * 2
	} else {
		tx_comp.V_counter = 0
		tx_comp.V_counterDouble = tx_comp.V_counter * 2
	}
	return tx_comp
}

func (tx_comp *tx_S_show_H_counter) showCounter() int {
	return tx_comp.V_counter + 1
}

func (tx_comp *tx_S_show_H_counter) tx_eh1() {
	tx_comp.V_counter--
	tx_comp.V_counterDouble = tx_comp.V_counter * 2
}

func (tx_comp *tx_S_show_H_counter) tx_eh2() {
	tx_comp.V_counter = 0
	tx_comp.V_counterDouble = tx_comp.V_counter * 2
}

func (tx_comp *tx_S_show_H_counter) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		case "eh2":
			tx_comp.tx_eh2()
		}
	}
}

func (tx_comp *tx_S_show_H_counter) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>showCounter</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/show-counter\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <h1 data-test=\"sc-h1\">Counter: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counter)))
	tx_w2.WriteString(" (doubled: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_counterDouble)))
	tx_w2.WriteString(")</h1> <div data-test=\"sc-show\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.showCounter())))
	tx_w2.WriteString("</div> <button data-test=\"sc-minus\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">-1</button> <button data-test=\"sc-reset\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh2-on=\"click\">reset</button> </body></html>")
}

type tx_S_slot_H_comp_H_swap struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_slot_H_comp_H_swap(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_slot_H_comp_H_swap {
	tx_comp := &tx_S_slot_H_comp_H_swap{}
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

func (tx_comp *tx_S_slot_H_comp_H_swap) tx_compute() {
	{
		tx_id := "tx-box-1"
		tx_child := tx_new_tx_H_box(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
		{
			tx_id := tx_id + "@tx-counter-1"
			tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
			tx_child.tx_compute(tx_id)
			tx_comp.tx_next[tx_id] = tx_child
		}
	}
}

func (tx_comp *tx_S_slot_H_comp_H_swap) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>slot-wrapped component swap</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/slot-comp-swap\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- a SEALED component placed inside a slot: its instance id uses the slot\n       '@' separator, so the swap path must parse '@' (regression: tx_dispatch\n       used to split the type name on ':' only -> empty swap -> dead demo) --> ")
	{
		tx_id := "tx-box-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_box)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_box_1_(tx_w2, tx_id)
		})
	}
	tx_w2.WriteString(" </body></html>")
}

func (tx_comp *tx_S_slot_H_comp_H_swap) tx_render_fill_tx_H_box_1_(tx_w *bytes.Buffer, tx_id string) {
	tx_w.WriteString(" ")
	{
		tx_id := tx_id + "@tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w, tx_id)
	}
	tx_w.WriteString(" ")
}

type tx_S_slot_H_fallback struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_slot_H_fallback(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_slot_H_fallback {
	tx_comp := &tx_S_slot_H_fallback{}
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

func (tx_comp *tx_S_slot_H_fallback) tx_compute() {
	{
		tx_id := "tx-box-1"
		tx_child := tx_new_tx_H_box(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
	{
		tx_id := "tx-box-2"
		tx_child := tx_new_tx_H_box(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_slot_H_fallback) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>slot fallback</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/slot-fallback\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- first box: no fill -> slot renders its fallback content --> ")
	{
		tx_id := "tx-box-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_box)
		tx_child.tx_render(tx_w2, tx_id, nil)
	}
	tx_w2.WriteString(" <!-- second box: a fill overrides the fallback --> ")
	{
		tx_id := "tx-box-2"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_box)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_box_2_(tx_w2)
		})
	}
	tx_w2.WriteString(" </body></html>")
}

func (tx_comp *tx_S_slot_H_fallback) tx_render_fill_tx_H_box_2_(tx_w *bytes.Buffer) {
	tx_w.WriteString("<span data-test=\"box-fill\">custom</span>")
}

type tx_S_slot_H_state struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_msg string `json:"msg"`
}

func tx_new_tx_S_slot_H_state(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_slot_H_state {
	tx_comp := &tx_S_slot_H_state{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_msg = "hi"
	}
	return tx_comp
}

func (tx_comp *tx_S_slot_H_state) tx_eh1() {
	tx_comp.V_msg = "bye"
}

func (tx_comp *tx_S_slot_H_state) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	{
		tx_id := "tx-box-1"
		tx_child := tx_new_tx_H_box(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_slot_H_state) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>dynamic slot fill</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/slot-state\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- a slot fill that interpolates PARENT state, and updates when it changes --> ")
	{
		tx_id := "tx-box-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_box)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_box_1_(tx_w2)
		})
	}
	tx_w2.WriteString(" <button data-test=\"change\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">change</button> </body></html>")
}

func (tx_comp *tx_S_slot_H_state) tx_render_fill_tx_H_box_1_(tx_w *bytes.Buffer) {
	tx_w.WriteString(" <span data-test=\"dynfill\">filled: ")
	tx_w.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_msg)))
	tx_w.WriteString("</span> ")
}

type tx_S_slots struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`
}

func tx_new_tx_S_slots(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_slots {
	tx_comp := &tx_S_slots{}
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

func (tx_comp *tx_S_slots) tx_compute() {
	{
		tx_id := "tx-slot-card-1"
		tx_val_title := "Recipe"
		tx_child := tx_new_tx_H_slot_H_card(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, "page", &tx_val_title)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_slots) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Slots</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/slots\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> ")
	{
		tx_id := "tx-slot-card-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_slot_H_card)
		tx_child.tx_render(tx_w2, tx_id, func() {
			tx_comp.tx_render_fill_tx_H_slot_H_card_1_(tx_w2)
		}, func() {
			tx_comp.tx_render_fill_tx_H_slot_H_card_1_footer(tx_w2)
		})
	}
	tx_w2.WriteString(" </body></html>")
}

func (tx_comp *tx_S_slots) tx_render_fill_tx_H_slot_H_card_1_(tx_w *bytes.Buffer) {
	tx_w.WriteString(" <p data-test=\"default-fill\">Mix and bake.</p>   ")
}

func (tx_comp *tx_S_slots) tx_render_fill_tx_H_slot_H_card_1_footer(tx_w *bytes.Buffer) {
	tx_w.WriteString("<em data-test=\"footer-fill-1\">Approved</em><em data-test=\"footer-fill-2\">By chef</em>")
}

type tx_S_state_H_survives struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_n int `json:"n"`
}

func tx_new_tx_S_state_H_survives(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_state_H_survives {
	tx_comp := &tx_S_state_H_survives{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_n = 0
	}
	return tx_comp
}

func (tx_comp *tx_S_state_H_survives) tx_eh1() {
	tx_comp.V_n++
}

func (tx_comp *tx_S_state_H_survives) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	{
		tx_id := "tx-counter-1"
		tx_child := tx_new_tx_H_counter(tx_comp.tx_prev, tx_comp.tx_next, tx_comp.tx_trigger, tx_comp.tx_trigger_handler, tx_id, tx_id)
		tx_child.tx_compute(tx_id)
		tx_comp.tx_next[tx_id] = tx_child
	}
}

func (tx_comp *tx_S_state_H_survives) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>state survives</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/state-survives\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"ss-page-n\">page: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_n)))
	tx_w2.WriteString("</p> <button data-test=\"ss-page-plus\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">+page</button> ")
	{
		tx_id := "tx-counter-1"
		tx_child := tx_comp.tx_next[tx_id].(*tx_H_counter)
		tx_child.tx_render(tx_w2, tx_id)
	}
	tx_w2.WriteString(" </body></html>")
}

type tx_S_state_H_types struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_flag  bool    `json:"flag"`
	V_ratio float64 `json:"ratio"`
	V_nums  []int   `json:"nums"`
}

func tx_new_tx_S_state_H_types(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_state_H_types {
	tx_comp := &tx_S_state_H_types{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_flag = true
		tx_comp.V_ratio = 1.5
		tx_comp.V_nums = []int{10, 20, 30}
	}
	return tx_comp
}

func (tx_comp *tx_S_state_H_types) tx_eh1() {
	tx_comp.V_flag = !tx_comp.V_flag
}

func (tx_comp *tx_S_state_H_types) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			tx_comp.tx_eh1()
		}
	}
	if tx_comp.V_flag {

	}
}

func (tx_comp *tx_S_state_H_types) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>state types</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/state-types\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- non-int/string state: bool, float64, slice --> <p data-test=\"flag\">flag: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_flag)))
	tx_w2.WriteString("</p> <p data-test=\"ratio\">ratio: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_ratio)))
	tx_w2.WriteString("</p> <p data-test=\"nums-len\">count: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_nums))))
	tx_w2.WriteString("</p> <!-- a bool used directly as a tx-if condition --> ")
	if tx_comp.V_flag {
		tx_w2.WriteString("<p data-test=\"when-on\">flag is on</p> <!-- handler mutating a bool via negation --> ")

	}
	tx_w2.WriteString("<button data-test=\"toggle\" data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-eh1-on=\"click\">toggle</button> </body></html>")
}

type tx_S_text struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_name  string `json:"name"`
	V_count int    `json:"count"`
	V_raw   string `json:"raw"`
}

func tx_new_tx_S_text(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_text {
	tx_comp := &tx_S_text{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_name = "world"
		tx_comp.V_count = 3
		tx_comp.V_raw = "<b>bold</b> & 'q'"
	}
	return tx_comp
}

func (tx_comp *tx_S_text) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>text interpolation</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/text\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <!-- multiple interpolations + a string literal in one text node --> <p data-test=\"greet\">Hello, ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_name)))
	tx_w2.WriteString("! You have ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count)))
	tx_w2.WriteString(" item")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint("s")))
	tx_w2.WriteString(".</p> <!-- arbitrary Go expressions, builtins, arithmetic --> <p data-test=\"expr\">2+2=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(2 + 2)))
	tx_w2.WriteString(" len=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_name))))
	tx_w2.WriteString(" up=")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count * 10)))
	tx_w2.WriteString("</p> <!-- two interpolations with NO text between them must stay separate --> <p data-test=\"adjacent\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_name)))
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_count)))
	tx_w2.WriteString("</p> <!-- interpolations must be HTML-escaped (no markup injection) --> <p data-test=\"escape\">")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_raw)))
	tx_w2.WriteString("</p> <!-- tx-ignore: the element's text is rendered verbatim (no { } interpolation,\n       skipped by the probe) — the escape hatch for literal braces --> <p data-test=\"ignored\">literal { name } and { undefinedVar } stay</p> </body></html>")
}

type tx_S_todos struct {
	tx_prev            url.Values     `json:"-"`
	tx_next            map[string]any `json:"-"`
	tx_trigger         string         `json:"-"`
	tx_trigger_handler string         `json:"-"`

	V_todos     []string `json:"todos"`
	V_lastAdded string   `json:"lastAdded"`
}

func tx_new_tx_S_todos(tx_prev url.Values, tx_next map[string]any, tx_trigger string, tx_trigger_handler string) *tx_S_todos {
	tx_comp := &tx_S_todos{}
	tx_comp.tx_prev = tx_prev
	tx_comp.tx_next = tx_next
	tx_comp.tx_trigger = tx_trigger
	tx_comp.tx_trigger_handler = tx_trigger_handler
	tx_prev_str := tx_prev.Get("page")
	if tx_prev_str != "" {
		json.Unmarshal([]byte(tx_prev_str), tx_comp)
	} else {
		tx_comp.V_todos = []string{"first", "second"}
		tx_comp.V_lastAdded = ""
	}
	return tx_comp
}

func (tx_comp *tx_S_todos) removeTodo(i int) {
	tx_comp.V_todos = append(tx_comp.V_todos[:i], tx_comp.V_todos[i+1:]...)
}

func (tx_comp *tx_S_todos) addTodo(text string) {
	tx_comp.V_todos = append(tx_comp.V_todos, text)
	tx_comp.V_lastAdded = text
}

func (tx_comp *tx_S_todos) tx_eh1(i int) {
	tx_comp.removeTodo(i)
}

func (tx_comp *tx_S_todos) tx_eh2(text string) {
	tx_comp.addTodo(text)
}

func (tx_comp *tx_S_todos) tx_compute() {
	if tx_comp.tx_trigger == "page" {
		switch tx_comp.tx_trigger_handler {
		case "eh1":
			var i int
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("i")), &i)
			tx_comp.tx_eh1(i)
		case "eh2":
			var text string
			json.Unmarshal([]byte(tx_comp.tx_prev.Get("text")), &text)
			tx_comp.tx_eh2(text)
		}
	}

	for i, t := range tx_comp.V_todos {
		_ = i
		_ = t

	}
}

func (tx_comp *tx_S_todos) tx_render(tx_w1 *bytes.Buffer, tx_w2 *bytes.Buffer) {
	tx_w1.WriteString("<!DOCTYPE html><html lang=\"en\"><head> <meta charset=\"UTF-8\"/> <title>Todos</title>  <script type=\"application/json\" id=\"tx-saved\" data-tx-page=\"/todos\">")
	tx_w2.WriteString("</script><script id=\"tx-runtime\">")
	fmt.Fprint(tx_w2, tx_runtime_script)
	tx_w2.WriteString("</script></head> <body> <p data-test=\"count\">total: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(len(tx_comp.V_todos))))
	tx_w2.WriteString("</p> <ul data-test=\"todo-list\"> ")

	for i, t := range tx_comp.V_todos {
		_ = i
		_ = t
		tx_w2.WriteString("<li> <span data-test=\"todo-text\">")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(t)))
		tx_w2.WriteString("</span> <button data-test=\"remove\" data-tx-trigger=\"page\" data-tx-target=\"")
		fmt.Fprint(tx_w2, "page")
		tx_w2.WriteString("\" data-tx-eh1-on=\"click\" data-tx-eh1-arg-i=\"")
		tx_w2.WriteString(html.EscapeString(fmt.Sprint(func() string { tx_b, _ := json.Marshal(i); return string(tx_b) }())))
		tx_w2.WriteString("\">x</button> </li>")

	}
	tx_w2.WriteString(" </ul> <form data-tx-trigger=\"page\" data-tx-target=\"")
	fmt.Fprint(tx_w2, "page")
	tx_w2.WriteString("\" data-tx-action=\"%2Ftodos/eh2\"> <input data-test=\"new-todo\" name=\"text\" type=\"text\"/> <button data-test=\"add\" type=\"submit\">add</button> </form> <p data-test=\"last-added\">Last added: ")
	tx_w2.WriteString(html.EscapeString(fmt.Sprint(tx_comp.V_lastAdded)))
	tx_w2.WriteString("</p> </body></html>")
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
	case "/attr-interp":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_attr_H_interp(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/badge":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_badge(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/compound":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_compound(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/conditionals":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_conditionals(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/counter-comp":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_counter_H_comp(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/counter":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_counter(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/defaulter":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_defaulter(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/derived-chain":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_derived_H_chain(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/echo/{msg}":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_echo_S__L_msg_R_(tx_prev, tx_next, tx_trigger, tx_trigger_handler, "")
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/expr-prop":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_expr_H_prop(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/filter":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_filter(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/funcprop-return":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_funcprop_H_return(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/importer":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_importer(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
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
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/init-derived-comp":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_init_H_derived_H_comp(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/init-derived-page":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_init_H_derived_H_page(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/input":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_input(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/loop-comps":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_loop_H_comps(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/loops":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_loops(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/loopvar-prop":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_loopvar_H_prop(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/map-loop":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_map_H_loop(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/nested-comp":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_nested_H_comp(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/nested-loop":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_nested_H_loop(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/render-edges":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_render_H_edges(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/seeded":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_seeded(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/show-counter":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_show_H_counter(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/slot-comp-swap":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_slot_H_comp_H_swap(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/slot-fallback":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_slot_H_fallback(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/slot-state":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_slot_H_state(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/slots":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_slots(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/state-survives":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_state_H_survives(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/state-types":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_state_H_types(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_compute()
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/text":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_text(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
		tx_next["page"] = tx_comp
		tx_comp.tx_render(&tx_buf1, &tx_buf2)
		tx_json, _ := json.Marshal(tx_next)
		tx_w.Write(tx_buf1.Bytes())
		tx_w.Write(tx_json)
		tx_w.Write(tx_buf2.Bytes())
		return
	case "/todos":
		var tx_buf1, tx_buf2 bytes.Buffer
		tx_comp := tx_new_tx_S_todos(tx_prev, tx_next, tx_trigger, tx_trigger_handler)
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
	case "tx-box":
		tx_comp := tx_new_tx_H_box(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_render(&buf, tx_target, nil)
	case "tx-counter":
		tx_comp := tx_new_tx_H_counter(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-defaulter":
		tx_comp := tx_new_tx_H_defaulter(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target, nil, nil)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-init-derived":
		tx_comp := tx_new_tx_H_init_H_derived(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
		tx_next[tx_target] = tx_comp
		tx_comp.tx_compute(tx_target)
		tx_comp.tx_render(&buf, tx_target)
	case "tx-seeded":
		tx_comp := tx_new_tx_H_seeded(tx_prev, tx_next, tx_trigger, tx_trigger_handler, tx_target, tx_target)
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
		Pattern: "GET /attr-interp",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_attr_H_interp(nil, tx_next, "", "")
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
		Pattern: "GET /badge",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_badge(nil, tx_next, "", "")
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
		Pattern: "GET /compound",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_compound(nil, tx_next, "", "")
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
		Pattern: "GET /conditionals",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_conditionals(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fconditionals/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/%2Fconditionals/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /counter-comp",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_counter_H_comp(nil, tx_next, "", "")
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
		Pattern: "GET /counter",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_counter(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fcounter/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/%2Fcounter/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /defaulter",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_defaulter(nil, tx_next, "", "")
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
		Pattern: "GET /derived-chain",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_derived_H_chain(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fderived-chain/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /echo/{msg}",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_echo_S__L_msg_R_(nil, tx_next, "", "", tx_r.PathValue("msg"))
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
		Pattern: "GET /expr-prop",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_expr_H_prop(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fexpr-prop/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /filter",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_filter(nil, tx_next, "", "")
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
		Pattern: "GET /funcprop-return",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_funcprop_H_return(nil, tx_next, "", "")
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
		Pattern: "GET /importer",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_importer(nil, tx_next, "", "")
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
			var tx_buf1, tx_buf2 bytes.Buffer
			tx_comp.tx_render(&tx_buf1, &tx_buf2)
			tx_json, _ := json.Marshal(tx_next)
			tx_w.Write(tx_buf1.Bytes())
			tx_w.Write(tx_json)
			tx_w.Write(tx_buf2.Bytes())
		},
	},
	{
		Pattern: "GET /init-derived-comp",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_init_H_derived_H_comp(nil, tx_next, "", "")
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
		Pattern: "GET /init-derived-page",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_init_H_derived_H_page(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Finit-derived-page/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /input",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_input(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Finput/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /loop-comps",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_loop_H_comps(nil, tx_next, "", "")
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
		Pattern: "GET /loops",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_loops(nil, tx_next, "", "")
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
		Pattern: "GET /loopvar-prop",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_loopvar_H_prop(nil, tx_next, "", "")
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
		Pattern: "GET /map-loop",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_map_H_loop(nil, tx_next, "", "")
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
		Pattern: "GET /nested-comp",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_nested_H_comp(nil, tx_next, "", "")
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
		Pattern: "GET /nested-loop",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_nested_H_loop(nil, tx_next, "", "")
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
		Pattern: "GET /render-edges",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_render_H_edges(nil, tx_next, "", "")
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
		Pattern: "GET /seeded",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_seeded(nil, tx_next, "", "")
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
		Pattern: "GET /show-counter",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_show_H_counter(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fshow-counter/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/%2Fshow-counter/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /slot-comp-swap",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_slot_H_comp_H_swap(nil, tx_next, "", "")
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
		Pattern: "GET /slot-fallback",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_slot_H_fallback(nil, tx_next, "", "")
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
		Pattern: "GET /slot-state",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_slot_H_state(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fslot-state/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /slots",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_slots(nil, tx_next, "", "")
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
		Pattern: "GET /state-survives",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_state_H_survives(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fstate-survives/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /state-types",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_state_H_types(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Fstate-types/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "GET /text",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_text(nil, tx_next, "", "")
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
		Pattern: "GET /todos",
		Handler: func(tx_w http.ResponseWriter, tx_r *http.Request) {
			tx_next := map[string]any{}
			tx_comp := tx_new_tx_S_todos(nil, tx_next, "", "")
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
		Pattern: "POST /tx/%2Ftodos/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/%2Ftodos/eh2",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-badge/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-button/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-compound/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-counter/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-defaulter/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-init-derived/eh1",
		Handler: tx_dispatch,
	},
	{
		Pattern: "POST /tx/tx-seeded/eh1",
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
