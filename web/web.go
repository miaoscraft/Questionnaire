package web

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

//Listen 监听80端口并响应http请求
func Listen(addr string) {
	http.HandleFunc("/", renderQuestionnaire)
	http.HandleFunc("/check", checkCode)
	http.HandleFunc("/submit", onSubmit)
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

func checkCode(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	code := r.URL.Query().Get("code")

	Codes.Lock()
	defer Codes.Unlock()

	if p, ok := Codes.M[code]; ok {
		fmt.Fprint(rw, p.Name)
	}
}

//接收提交的表单，判定分数
func onSubmit(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	log.Println(r.PostForm)
}

//Questions 题目列表
var Questions struct {
	ChoiceQuestions []struct {
		Question string `xml:"value,attr"`
		Options  []struct {
			Value string `xml:",chardata"`
			Point int    `xml:"p,attr"`
		} `xml:"Option"`
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
	log.Println(Questions)
}
