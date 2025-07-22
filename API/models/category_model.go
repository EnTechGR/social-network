package models

// Category represents a discussion category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CategoryWithPosts struct {
	ID    int            `json:"id"`
	Name  string         `json:"name"`
	Posts []PostWithUser `json:"posts"`
}