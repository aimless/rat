package reddit

import (
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"rat/util"
	"regexp"
)

func commentUserNameNode(n *html.Node) bool {
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

func commentTextNode(n *html.Node) bool {
	return n.Parent.Data == "p"
}

func subEntryHeadline(n *html.Node) bool {
	return n.Data == "h3"
}

func subEntryLink(n *html.Node) bool {
	lre := regexp.MustCompile(`\/r\/\w*\/comments\/\w*\/\w*\/`)
	return n.Data == "a" && util.HasAttr(n, "href") && lre.MatchString(scrape.Attr(n, "href"))
}

func subEntryNode(n *html.Node) bool {
	return util.HasAttr(n, "data-testid") && scrape.Attr(n, "data-testid") == "post-container"
}

func commentDeepLink(n *html.Node) bool {
	return util.HasAttr(n, "src") && util.HasAttr(n, "method")
}
