package web

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//Listen 监听指定地址并响应http请求
func Listen(addr string) {
	http.HandleFunc("/", renderQuestionnaire)
	http.HandleFunc("/check", handleCheckCode)
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

func handleCheckCode(rw http.ResponseWriter, r *http.Request) {
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

	if p, ok := checkCode(r.PostFormValue("code")); ok {
		//验证通过
		//判题
		var score int
		for i, v := range Questions.ChoiceQuestions {
			choice := r.PostForm.Get(strconv.Itoa(i))
			if o, err := strconv.Atoi(choice); err == nil {
				if o >= 0 && o < len(v.Options) {
					score += v.Options[o].Point
				}
			}
		}
		log.Printf("%v, %v\n总分:%d", p, r.PostForm, score)
		if score >= Questions.MinScore {
			fmt.Fprint(rw, "您成功通过")
		} else {
			fmt.Fprint(rw, "您没有通过")
		}
	} else {
		//验证不通过
		fmt.Fprint(rw, "验证码错误")
	}

}

//检查验证码，若验证通过，返回true，并使验证码失效
func checkCode(code string) (p Player, ok bool) {
	Codes.Lock()
	defer Codes.Unlock()
	if p, ok = Codes.M[code]; ok {
		delete(Codes.M, code)
		delete(Codes.U, p)
	}
	return
}

//Questions 题目列表
var Questions struct {
	MinScore        int `xml:"MinScore,attr"`
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
	// log.Println(Questions)
}
