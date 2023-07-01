package util

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func ScrapeByTagType(tagType string) scrape.Matcher {
	return func(n *html.Node) bool {
		return n.Data == tagType
	}
}

func ScrapeByTagTypeHierarchy(tagTypes ...string) scrape.Matcher {
	return func(n *html.Node) bool {
		if len(tagTypes) >= 1 {
			if n.Parent != nil {
				return ScrapeByTagType(tagTypes[0])(n) && ScrapeByTagTypeHierarchy(tagTypes[1:]...)(n.Parent)
			} else {
				return false
			}
		} else {
			return true
		}
	}
}

func HasAttr(node *html.Node, attrKey string) bool {
	for _, attr := range node.Attr {
		if attr.Key == attrKey {
			return true
		}
	}
	return false
}

func AttrToInt(node *html.Node, attr string) (int64, error) {
	s := scrape.Attr(node, attr)
	return strconv.ParseInt(s, 10, 64)
}

func PrintNodeDebug(node *html.Node, pad string) {

	if node == nil {
		fmt.Println(pad, "< nil node >")
		return
	}

	r, _ := regexp.Compile("[a-z]+")

	if r.MatchString(node.Data) {
		fmt.Print(pad, node.Data)
	} else {
		fmt.Print(pad, "( strange node )")
	}
	for d := node.Parent; d != nil; d = d.Parent {
		if r.MatchString(d.Data) {
			fmt.Print(" < ", d.Data)
		} else {
			fmt.Print(" < ", "( strange node )")
		}
	}
	fmt.Println()
	fmt.Println(pad, "ns:", "<", node.Namespace, len(node.Namespace), ">")
	for _, attribute := range node.Attr {
		fmt.Println(pad, "attr:", "<", attribute.Key, ":", attribute.Val, ">")
	}

	fmt.Println("--")
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		PrintNodeDebug(c, pad+" ")
	}
}
