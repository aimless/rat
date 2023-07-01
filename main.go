package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"rat/model"
	"rat/reddit"

	"github.com/julienschmidt/httprouter"
)

func getRoot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	io.WriteString(w, "This is my website!\n")
}

func getPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sub := ps.ByName("sub")
	post := ps.ByName("post")
	title := ps.ByName("title")
	url := "https://www.reddit.com/r/" + sub + "/comments/" + post + "/" + title + "/"
	postInstance, _ := reddit.GetPost(url)

	t := template.New("post.template")
	tem, err := t.ParseFiles("post.template")
	fmt.Println(err)
	err = tem.Execute(w, postInstance)
	fmt.Println(err)
}

func getStyling(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	b, _ := ioutil.ReadFile("styling.css")
	buf := bytes.NewBuffer(b)
	buf.WriteTo(w)
}

func foo() {
	var comfn func(*model.Comment, string)
	comfn = func(comment *model.Comment, pad string) {
		fmt.Println(pad, comment.Score, comment.Author, comment.PostDateTime)
		fmt.Println()
		for _, line := range comment.Text {
			fmt.Println(pad, line)
			fmt.Println()
		}
		for _, cc := range comment.Children {
			comfn(cc, ""+" ")
		}
	}

	p, _ := reddit.GetPost("https://www.reddit.com/r/arbeitsleben/comments/14e0ywi/unversch%C3%A4mte_vorgesetzte/")

	fmt.Println(p.Score, p.Author, p.PostDateTime)
	fmt.Println()
	fmt.Println(p.Title)
	fmt.Println()
	for _, line := range p.Text {
		fmt.Println(line)
		fmt.Println()
	}
	for _, c := range p.Comments {
		comfn(c, "")
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", getRoot)
	router.GET("/styling", getStyling)
	router.GET("/r/:sub/comments/:post/:title/", getPost)
	http.ListenAndServe(":3333", router)
}
