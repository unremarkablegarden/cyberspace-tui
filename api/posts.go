package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/euklides/cyberspace-cli/models"
)

// FirestoreClient handles Firestore REST API calls
type FirestoreClient struct {
	ProjectID string
	IDToken   string
}

// NewFirestoreClient creates a new Firestore client
func NewFirestoreClient(projectID, idToken string) *FirestoreClient {
	return &FirestoreClient{
		ProjectID: projectID,
		IDToken:   idToken,
	}
}

func (c *FirestoreClient) baseURL() string {
	return fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents", c.ProjectID)
}

// FetchPosts retrieves posts from Firestore
func (c *FirestoreClient) FetchPosts(limit int) ([]models.Post, error) {
	query := map[string]interface{}{
		"structuredQuery": map[string]interface{}{
			"from": []map[string]string{
				{"collectionId": "posts"},
			},
			"where": map[string]interface{}{
				"fieldFilter": map[string]interface{}{
					"field": map[string]string{"fieldPath": "deleted"},
					"op":    "EQUAL",
					"value": map[string]bool{"booleanValue": false},
				},
			},
			"orderBy": []map[string]interface{}{
				{
					"field":     map[string]string{"fieldPath": "createdAt"},
					"direction": "DESCENDING",
				},
			},
			"limit": limit,
		},
	}

	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s:runQuery", c.baseURL())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firestore error: %s", string(body))
	}

	// Parse Firestore response
	var results []firestoreQueryResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(results))
	for _, r := range results {
		if r.Document == nil {
			continue
		}
		post, err := parsePost(r.Document)
		if err != nil {
			continue // Skip malformed posts
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// FetchPost retrieves a single post by ID
func (c *FirestoreClient) FetchPost(postID string) (*models.Post, error) {
	url := fmt.Sprintf("%s/posts/%s", c.baseURL(), postID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firestore error: %s", string(body))
	}

	var doc firestoreDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	post, err := parsePost(&doc)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// FetchReplies retrieves replies for a post
func (c *FirestoreClient) FetchReplies(postID string) ([]models.Reply, error) {
	query := map[string]interface{}{
		"structuredQuery": map[string]interface{}{
			"from": []map[string]string{
				{"collectionId": "replies"},
			},
			"where": map[string]interface{}{
				"compositeFilter": map[string]interface{}{
					"op": "AND",
					"filters": []map[string]interface{}{
						{
							"fieldFilter": map[string]interface{}{
								"field": map[string]string{"fieldPath": "postId"},
								"op":    "EQUAL",
								"value": map[string]string{"stringValue": postID},
							},
						},
						{
							"fieldFilter": map[string]interface{}{
								"field": map[string]string{"fieldPath": "deleted"},
								"op":    "EQUAL",
								"value": map[string]bool{"booleanValue": false},
							},
						},
					},
				},
			},
			"orderBy": []map[string]interface{}{
				{
					"field":     map[string]string{"fieldPath": "createdAt"},
					"direction": "ASCENDING",
				},
			},
			"limit": 100,
		},
	}

	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s:runQuery", c.baseURL())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.IDToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firestore error: %s", string(body))
	}

	var results []firestoreQueryResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	replies := make([]models.Reply, 0, len(results))
	for _, r := range results {
		if r.Document == nil {
			continue
		}
		reply, err := parseReply(r.Document)
		if err != nil {
			continue
		}
		replies = append(replies, reply)
	}

	return replies, nil
}

// parseReply converts a Firestore document to a Reply
func parseReply(doc *firestoreDocument) (models.Reply, error) {
	reply := models.Reply{}

	// Extract ID from document name
	name := doc.Name
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			reply.ID = name[i+1:]
			break
		}
	}

	fields := doc.Fields

	if v, ok := fields["postId"]; ok && v.StringValue != nil {
		reply.PostID = *v.StringValue
	}
	if v, ok := fields["authorId"]; ok && v.StringValue != nil {
		reply.AuthorID = *v.StringValue
	}
	if v, ok := fields["authorUsername"]; ok && v.StringValue != nil {
		reply.AuthorUsername = *v.StringValue
	}
	if v, ok := fields["content"]; ok && v.StringValue != nil {
		reply.Content = *v.StringValue
	}
	if v, ok := fields["createdAt"]; ok && v.TimestampValue != nil {
		t, _ := time.Parse(time.RFC3339Nano, *v.TimestampValue)
		reply.CreatedAt = t
	}
	if v, ok := fields["deleted"]; ok && v.BooleanValue != nil {
		reply.Deleted = *v.BooleanValue
	}

	return reply, nil
}

// Firestore response structures
type firestoreQueryResult struct {
	Document *firestoreDocument `json:"document"`
}

type firestoreDocument struct {
	Name   string                     `json:"name"`
	Fields map[string]firestoreValue `json:"fields"`
}

type firestoreValue struct {
	StringValue    *string                    `json:"stringValue,omitempty"`
	IntegerValue   *string                    `json:"integerValue,omitempty"`
	BooleanValue   *bool                      `json:"booleanValue,omitempty"`
	TimestampValue *string                    `json:"timestampValue,omitempty"`
	ArrayValue     *firestoreArrayValue       `json:"arrayValue,omitempty"`
	MapValue       *firestoreMapValue         `json:"mapValue,omitempty"`
}

type firestoreArrayValue struct {
	Values []firestoreValue `json:"values"`
}

type firestoreMapValue struct {
	Fields map[string]firestoreValue `json:"fields"`
}

// parsePost converts a Firestore document to a Post
func parsePost(doc *firestoreDocument) (models.Post, error) {
	post := models.Post{}

	// Extract ID from document name (last segment)
	name := doc.Name
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			post.ID = name[i+1:]
			break
		}
	}

	fields := doc.Fields

	if v, ok := fields["authorId"]; ok && v.StringValue != nil {
		post.AuthorID = *v.StringValue
	}
	if v, ok := fields["authorUsername"]; ok && v.StringValue != nil {
		post.AuthorUsername = *v.StringValue
	}
	if v, ok := fields["content"]; ok && v.StringValue != nil {
		post.Content = *v.StringValue
	}
	if v, ok := fields["createdAt"]; ok && v.TimestampValue != nil {
		t, _ := time.Parse(time.RFC3339Nano, *v.TimestampValue)
		post.CreatedAt = t
	}
	if v, ok := fields["repliesCount"]; ok && v.IntegerValue != nil {
		n, _ := strconv.Atoi(*v.IntegerValue)
		post.RepliesCount = n
	}
	if v, ok := fields["bookmarksCount"]; ok && v.IntegerValue != nil {
		n, _ := strconv.Atoi(*v.IntegerValue)
		post.BookmarksCount = n
	}
	if v, ok := fields["topics"]; ok && v.ArrayValue != nil {
		for _, tv := range v.ArrayValue.Values {
			if tv.StringValue != nil {
				post.Topics = append(post.Topics, *tv.StringValue)
			}
		}
	}
	if v, ok := fields["deleted"]; ok && v.BooleanValue != nil {
		post.Deleted = *v.BooleanValue
	}

	return post, nil
}
