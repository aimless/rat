package model

import "time"

type Post struct {
	Author       string
	Score        int
	PostDateTime time.Time
	CommentCount int
	Post         string

	Title    string
	Text     []string
	Comments []*Comment
}

type Comment struct {
	Author       string
	Score        int
	PostDateTime time.Time
	Depth        int
	Parent       string
	Thing        string
	Post         string
	Text         []string
	Children     []*Comment
}

type SubEntry struct {
	Title string
	Link  string
}

type Sub struct {
	Entries []*SubEntry
}
