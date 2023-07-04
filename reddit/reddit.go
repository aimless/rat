package reddit

import (
	"fmt"
	"net/http"
	"rat/model"
	"rat/util"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const redditUrl = "https://www.reddit.com"
const userAgent = "RAT/1.0"

func request(method string, link string) (*http.Response, error) {
	url := redditUrl + link
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("User-Agent", userAgent)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetComments Get comments recursively from a given link with a given method, visited keeps track of previously visited links
func GetComments(link string, method string, visited *map[string]bool) []*model.Comment {
	resp, err := request(method, link)
	if err != nil {
		panic(err)
	}
	roots, err := html.ParseFragment(resp.Body, nil)
	if err != nil {
		panic(err)
	}

	var r []*model.Comment
	links := make(map[string]bool)
	for _, e := range roots {
		for _, l := range getDeepCommentLinks(e) {
			links[l] = true
		}
		comments := scrape.FindAllNested(e, util.ScrapeByTagType("shreddit-comment"))
		for _, comment := range comments {
			c, ok := GetCommentFromNode(comment)
			if ok {
				r = append(r, c)
			}
		}
	}
	for k := range links {
		_, present := (*visited)[k]
		if !present {
			fmt.Println(k)
			r = append(r, GetComments(k, "POST", &links)...)
		}
	}
	return r
}

// GetCommentFromNode Extract comment from a given node
func GetCommentFromNode(node *html.Node) (*model.Comment, bool) {
	if node == nil {
		return nil, false
	}
	score, _ := util.AttrToInt(node, "score")
	depth, _ := util.AttrToInt(node, "depth")
	thing := scrape.Attr(node, "thingid")
	post := scrape.Attr(node, "postid")
	parent := scrape.Attr(node, "parentid")

	p, _ := scrape.Find(node, scrape.ById("-post-rtjson-content"))
	user, userOk := scrape.Find(node, commentUserNameNode)
	if !userOk || user == nil || user.FirstChild == nil {
		return nil, false
	}
	timeNode, _ := scrape.Find(node, util.ScrapeByTagType("faceplate-timeago"))
	userName := scrape.Text(user.FirstChild)
	timeStr := scrape.Attr(timeNode, "ts")
	t, _ := time.Parse("2006-01-02T15:04:05-0700", timeStr)
	var text []string
	for _, node := range scrape.FindAll(p, commentTextNode) {
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

// GetPostFromNode Extract post from a given node
func GetPostFromNode(node *html.Node) model.Post {
	title := scrape.Attr(node, "post-title")
	score, _ := util.AttrToInt(node, "score")
	author := scrape.Attr(node, "author")
	timeStr := scrape.Attr(node, "created-timestamp")
	postDateTime, _ := time.Parse("2006-01-02T15:04:05-0700", timeStr)
	commentCount, _ := util.AttrToInt(node, "comment-count")
	post := scrape.Attr(node, "id")
	subredditPrefixed := scrape.Attr(node, "subreddit-prefixed-name")
	subreddit, _ := strings.CutPrefix(subredditPrefixed, "r/")

	p := scrape.FindAll(node, util.ScrapeByTagTypeHierarchy("p", "div", "div", "shreddit-post"))
	var text []string
	for _, e := range p {
		text = append(text, scrape.Text(e))
	}

	id, _ := strings.CutPrefix(post, "t3_")
	url := "/svc/shreddit/comments/" + subreddit + "/" + id
	comments := transformCommentList(GetComments(url, "GET", new(map[string]bool)))

	return model.Post{
		Author:       author,
		Score:        int(score),
		Title:        title,
		PostDateTime: postDateTime,
		CommentCount: int(commentCount),
		Post:         post,
		Comments:     comments,
		Text:         text,
	}
}

func GetPost(link string) (*model.Post, bool) {
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

func GetSubEntriesFromNode(node *html.Node) []*model.SubEntry {
	postNodes := scrape.FindAll(node, subEntryNode)

	var r []*model.SubEntry

	for _, post := range postNodes {
		h, headerOk := scrape.Find(post, subEntryHeadline)
		l, linkOk := scrape.Find(post, subEntryLink)
		if linkOk && headerOk {
			title := scrape.Text(h)
			r = append(r, &model.SubEntry{Title: title, Link: scrape.Attr(l, "href")})
		}
	}
	return r
}

func GetSubreddit(url string) (*model.Sub, bool) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, false
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, false
	}
	entries := GetSubEntriesFromNode(root)
	sub := model.Sub{Entries: entries}
	pretty.Println(sub)
	return &sub, true
}

// Extract comment links leading to unloaded comments from nodes
func getDeepCommentLinks(node *html.Node) []string {
	foo := scrape.FindAllNested(node, commentDeepLink)
	var links []string

	for _, l := range foo {
		link := scrape.Attr(l, "src")
		links = append(links, link)
	}
	return links
}

func sort(comments []*model.Comment) []*model.Comment {
	for _, i := range comments {
		for _, j := range comments {
			if i.Score > j.Score {
				t := *j
				*j = *i
				*i = t
			}
		}
		sort(i.Children)
	}
	return comments
}

// Transform flat list into trees of comments, build a comment tree from a map of comments
func transformCommentList(comments []*model.Comment) []*model.Comment {
	c := map[string]*model.Comment{}
	var r []*model.Comment

	for _, comment := range comments {
		c[comment.Thing] = comment
	}

	for k := range c {
		item := c[k]
		if item.Depth == 0 {
			r = append(r, item)
		}
		if item.Depth != 0 && item.Parent == "" {
			pretty.Println(item)
		}
	}

	for _, v := range c {
		if v.Parent != "" {
			parent := c[v.Parent]
			if parent.Children == nil {
				parent.Children = []*model.Comment{}
			}
			parent.Children = append(parent.Children, v)
		}
	}
	return sort(r)
}
