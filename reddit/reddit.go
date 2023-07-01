package reddit

import (
	"fmt"
	"net/http"
	"rat/model"
	"rat/util"
	"regexp"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const redditUrl = "https://www.reddit.com"
const userAgent = "RUDD/1.0"

// Get comments recursively from a given link with a given method, visited keeps track of previously visited links
func GetComments(link string, method string, visited *map[string]bool) []*model.Comment {
	url := redditUrl + link
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("User-Agent", userAgent)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	roots, err := html.ParseFragment(resp.Body, nil)
	if err != nil {
		panic(err)
	}

	comms := []*model.Comment{}
	links := make(map[string]bool)
	for _, e := range roots {
		for _, l := range getDeepCommentLinks(e) {
			links[l] = true
		}
		comments := scrape.FindAllNested(e, util.ScrapeByTagType("shreddit-comment"))
		for _, comment := range comments {
			c, ok := GetCommentFromNode(comment)
			if ok {
				comms = append(comms, c)
			}
		}
	}
	for k := range links {
		_, present := (*visited)[k]
		if !present {
			fmt.Println(k)
			comms = append(comms, GetComments(k, "POST", &links)...)
		}
	}
	return comms
}

// Extract comment from a given node
func GetCommentFromNode(node *html.Node) (*model.Comment, bool) {
	if node == nil {
		return nil, false
	}
	textNode := func(n *html.Node) bool {
		return n.Parent.Data == "p"
	}

	userNameNode := func(n *html.Node) bool {
		r := regexp.MustCompile(`\/user\/\w+\/`)
		link := scrape.Attr(n, "href")
		isLink := n.Data == "a"
		isUserLink := r.MatchString(link)
		hasTextChildNode := false
		if n.FirstChild != nil {
			child := n.FirstChild
			hasTextChildNode = child.Type == html.TextNode
		}

		return isLink && isUserLink && hasTextChildNode
	}

	score, _ := util.AttrToInt(node, "score")
	depth, _ := util.AttrToInt(node, "depth")
	thing := scrape.Attr(node, "thingid")
	post := scrape.Attr(node, "postid")
	parent := scrape.Attr(node, "parentid")

	p, _ := scrape.Find(node, scrape.ById("-post-rtjson-content"))
	user, userOk := scrape.Find(node, userNameNode)
	if !userOk || user == nil || user.FirstChild == nil {
		return nil, false
	}
	timeNode, _ := scrape.Find(node, util.ScrapeByTagType("faceplate-timeago"))
	userName := scrape.Text(user.FirstChild)
	timeStr := scrape.Attr(timeNode, "ts")
	t, _ := time.Parse("2006-01-02T15:04:05-0700", timeStr)
	text := []string{}
	for _, node := range scrape.FindAll(p, textNode) {
		s := scrape.Text(node)
		text = append(text, s)
	}
	return &model.Comment{
		Author:       userName,
		PostDateTime: t,
		Score:        int(score),
		Text:         text,
		Post:         post,
		Thing:        thing,
		Depth:        int(depth),
		Parent:       parent,
	}, true
}

// Extract post from a given node
func GetPostFromNode(node *html.Node) model.Post {
	title := scrape.Attr(node, "post-title")
	score, _ := util.AttrToInt(node, "score")
	author := scrape.Attr(node, "author")
	timeStr := scrape.Attr(node, "created-timestamp")
	t, _ := time.Parse("2006-01-02T15:04:05-0700", timeStr)
	commentCount, _ := util.AttrToInt(node, "comment-count")
	post := scrape.Attr(node, "id")

	p := scrape.FindAll(node, util.ScrapeByTagTypeHierarchy("p", "div", "div", "shreddit-post"))
	text := []string{}
	for _, e := range p {
		text = append(text, scrape.Text(e))
	}

	id, _ := strings.CutPrefix(post, "t3_")
	url := "/svc/shreddit/comments/arbeitsleben/" + id
	comments := GetComments(url, "GET", new(map[string]bool))

	return model.Post{
		Author:       author,
		Score:        int(score),
		Title:        title,
		PostDateTime: t,
		CommentCount: int(commentCount),
		Post:         post,
		Comments:     transformCommentList(comments),
		Text:         text,
	}
}

func GetPost(link string) (*model.Post, bool) {
	// request and parse the front page
	resp, err := http.Get(link)
	if err != nil {
		return nil, false
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, false
	}

	matcher := func(n *html.Node) bool {
		return n.Data == "shreddit-post"
	}

	postNode, _ := scrape.Find(root, matcher)
	post := GetPostFromNode(postNode)
	return &post, true
}

// Extract comment links leading to unloaded comments from nodes
func getDeepCommentLinks(node *html.Node) []string {
	m := func(n *html.Node) bool {
		return util.HasAttr(n, "src") && util.HasAttr(n, "method")
	}
	foo := scrape.FindAllNested(node, m)
	links := []string{}

	for _, l := range foo {
		link := scrape.Attr(l, "src")
		links = append(links, link)
	}
	return links
}

// Transform flat list into trees of comments
func transformCommentList(comments []*model.Comment) []*model.Comment {
	r := map[string]*model.Comment{}
	for _, comment := range comments {
		r[comment.Thing] = comment
	}
	return hierarchizeComments(r)
}

// Build a comment tree from a map of comments
func hierarchizeComments(comments map[string]*model.Comment) []*model.Comment {
	r := []*model.Comment{}
	for k := range comments {
		item := comments[k]
		if item.Depth == 0 {
			r = append(r, item)
		}
	}

	for _, v := range comments {
		if v.Parent != "" {
			parent := comments[v.Parent]
			if parent.Children == nil {
				parent.Children = []*model.Comment{}
			}
			parent.Children = append(parent.Children, v)
		}
	}
	return r
}
