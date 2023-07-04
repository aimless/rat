package main

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"rat/reddit"

	"github.com/julienschmidt/httprouter"
)

func getRoot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	io.WriteString(w, "This is my website!\n")
}

func getSubreddit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sub := ps.ByName("sub")
	url := "https://www.reddit.com/r/" + sub
	subredditInstance, ok := reddit.GetSubreddit(url)
	if ok {
		t := template.New("subreddit.template")
		tem, err := t.ParseFiles("subreddit.template")
		if err != nil {
			panic(err)
		}
		err = tem.Execute(w, subredditInstance)
		if err != nil {
			panic(err)
		}
	} else {
		io.WriteString(w, "Subreddit not found!\n")
	}
}

func getPost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sub := ps.ByName("sub")
	post := ps.ByName("post")
	title := ps.ByName("title")
	url := "https://www.reddit.com/r/" + sub + "/comments/" + post + "/" + title + "/"
	postInstance, _ := reddit.GetPost(url)

	t := template.New("post.template")
	tem, err := t.ParseFiles("post.template")
	if err != nil {
		panic(err)
	}
	err = tem.Execute(w, postInstance)
	if err != nil {
		panic(err)
	}
}

func getStyling(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	b, _ := ioutil.ReadFile("styling.css")
	buf := bytes.NewBuffer(b)
	buf.WriteTo(w)
}

func main() {
	router := httprouter.New()
	router.GET("/", getRoot)
	router.GET("/styling", getStyling)
	router.GET("/r/:sub/comments/:post/:title", getPost)
	router.GET("/r/:sub/comments/:post/:title/", getPost)
	router.GET("/r/:sub", getSubreddit)
	router.GET("/r/:sub/", getSubreddit)
	http.ListenAndServe(":3333", router)
}
