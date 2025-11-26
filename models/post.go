package models

import "time"

// Post represents a Cyberspace post
type Post struct {
	ID             string    `json:"id"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	RepliesCount   int       `json:"repliesCount"`
	BookmarksCount int       `json:"bookmarksCount"`
	Topics         []string  `json:"topics"`
	Deleted        bool      `json:"deleted"`
}

// Reply represents a reply to a post
type Reply struct {
	ID             string    `json:"id"`
	PostID         string    `json:"postId"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	Deleted        bool      `json:"deleted"`
}
