package chat

import (
	"encoding/json"
	"fmt"
)

//Msg is a message sent by other
type Msg jsonMsg
type jsonMsg struct {
	Text string `json:"text,omitempty"`

	Bold          bool   `json:"bold,omitempty"`          //粗体
	Italic        bool   `json:"Italic,omitempty"`        //斜体
	UnderLined    bool   `json:"underlined,omitempty"`    //下划线
	StrikeThrough bool   `json:"strikethrough,omitempty"` //删除线
	Obfuscated    bool   `json:"obfuscated,omitempty"`    //随机
	Color         string `json:"color,omitempty"`

	Translate string `json:"translate,omitempty"`
	With      []Msg  `json:"with,omitempty"`
	Extra     []Msg  `json:"extra,omitempty"`

	Pos byte `json:"-"` //消息显示位置
}

//UnmarshalJSON parse a []byte into msg
func (m *Msg) UnmarshalJSON(input []byte) (err error) {
	if len(input) < 1 {
		return fmt.Errorf("json too short")
	}
	if input[0] == '"' {
		err = json.Unmarshal(input, &m.Text)
	} else {
		err = json.Unmarshal(input, (*jsonMsg)(m))
	}
	return
}

var colors = map[string]int{
	"black":        30,
	"dark_blue":    34,
	"dark_green":   32,
	"dark_aqua":    36,
	"dark_red":     31,
	"dark_purple":  35,
	"gold":         33,
	"gray":         37,
	"dark_gray":    90,
	"blue":         94,
	"green":        92,
	"aqua":         96,
	"red":          91,
	"light_purple": 95,
	"yellow":       93,
	"white":        97,
}

// String return the message with escape sequence for ansi color.
// On windows, you may want print this string using
// github.com/mattn/go-colorable.
// func (m Msg) String() (s string) {
// 	var format string
// 	if m.Bold {
// 		format += "1;"
// 	}
// 	if m.Italic {
// 		format += "3;"
// 	}
// 	if m.UnderLined {
// 		format += "4;"
// 	}
// 	if m.StrikeThrough {
// 		format += "9;"
// 	}
// 	if m.Color != "" {
// 		format += fmt.Sprintf("%d;", colors[m.Color])
// 	}

// 	if format != "" {
// 		s = "\033[" + format[:len(format)-1] + "m"
// 	}

// 	s += m.Text

// 	//handle translate
// 	if m.Translate != "" {
// 		args := make([]interface{}, len(m.With))
// 		for i, v := range m.With {
// 			args[i] = v
// 		}

// 		s += fmt.Sprintf(lang.Translate[m.Translate], args...)
// 	}

// 	if format != "" {
// 		s += "\033[0m"
// 	}

// 	if m.Extra != nil {
// 		for i := range m.Extra {
// 			s += Msg(m.Extra[i]).String()
// 		}
// 	}
// 	return
// }

//NewTranslateMsg create a translated message
func NewTranslateMsg(translate string, args ...Msg) *Msg {
	return &Msg{
		Translate: translate,
		With:      args,
	}
}
