package models

// Category represents a discussion category
// swagger:model Category
type Category struct {
    // The unique identifier for the category.
    // example: 1
    ID int `json:"id"`
    // The name of the category.
    // example: Technology
    Name string `json:"name"`
}

// CategoryWithPosts represents a category along with all posts associated with it.
// swagger:model CategoryWithPosts
type CategoryWithPosts struct {
    // The unique identifier for the category.
    // example: 1
    ID int `json:"id"`
    // The name of the category.
    // example: Technology
    Name string `json:"name"`
    // A list of posts that belong to this category.
    Posts []PostWithUser `json:"posts"`
}