package web

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

//Listen 监听80端口并响应http请求
func Listen(addr string) {
	http.HandleFunc("/", renderQuestionnaire)
	http.ListenAndServe(addr, nil)
}

var questionnaire = template.Must(template.
	ParseFiles("questionnaire.html"))

func renderQuestionnaire(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if err := questionnaire.Execute(rw, Questions); err != nil {
		log.Println(err)
	}
}

//Questions 题目列表
var Questions struct {
	ChoiceQuestions []struct {
		Question string   `xml:"value,attr"`
		Options  []string `xml:"Option"`
	} `xml:"ChoiceQuestions>Question"`
}

func init() {
	//读取问卷
	q, err := ioutil.ReadFile("questionnaire.xml")
	if err != nil {
		panic(err)
	}

	err = xml.Unmarshal(q, &Questions)
	if err != nil {
		panic(err)
	}
}
