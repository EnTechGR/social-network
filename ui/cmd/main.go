package main

import (
	"encoding/json" // Add this import
	"log"
	"net/http"
)

// Add a struct to unmarshal the session verify response
type SessionVerifyResponse struct {
	User *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
	CSRFToken string `json:"csrf_token"` // Ensure this matches your backend's field name
}

func checkSession(r *http.Request) (bool, string) { // Modified to return CSRF token
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Println("No session cookie:", err)
		return false, ""
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/forum/api/session/verify", nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return false, ""
	}
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Session verify request failed:", err)
		return false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Session verify returned status:", resp.StatusCode)
		return false, ""
	}

	// Read and parse the response body to get the CSRF token
	var sessionResp SessionVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		log.Println("Failed to decode session verify response:", err)
		return false, ""
	}

	return true, sessionResp.CSRFToken
}

func router(w http.ResponseWriter, r *http.Request) {
	// ... (unchanged code for handling paths)
	switch r.URL.Path {
	case "/index":
		http.ServeFile(w, r, "./static/templates/index.html")
	case "/":
		http.ServeFile(w, r, "./static/templates/index.html")
	case "/login":
		ok, _ := checkSession(r) // We don't need CSRF token here
		if ok {
			http.Redirect(w, r, "/user/feed", http.StatusFound) // Changed to user/feed for consistency
			return
		}
		http.ServeFile(w, r, "./static/templates/login.html")
	case "/register":
		ok, _ := checkSession(r) // We don't need CSRF token here
		if ok {
			http.Redirect(w, r, "/user/feed", http.StatusFound) // Changed to user/feed for consistency
			return
		}
		http.ServeFile(w, r, "./static/templates/register.html")
	case "/guest":
		http.ServeFile(w, r, "./static/templates/guest/guest_mainpage.html")
	case "/guest/feed":
		http.ServeFile(w, r, "./static/templates/guest/guest_feed.html")
	case "/guest/category":
		http.ServeFile(w, r, "./static/templates/guest/guest_category.html")
	case "/guest/post":
		http.ServeFile(w, r, "./static/templates/guest/guest_post.html")
	case "/user":
		if ok, _ := checkSession(r); !ok { // We don't need CSRF token here
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_mainpage.html")
	case "/user/feed":
		if ok, _ := checkSession(r); !ok { // Get CSRF token here
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_feed.html")
	case "/user/category":
		if ok, _ := checkSession(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_category.html")
	case "/user/post":
		if ok, _ := checkSession(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_post.html")
	case "/user/liked-posts":
		if ok, _ := checkSession(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_liked_posts.html")
	case "/user/created-posts":
		if ok, _ := checkSession(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./static/templates/user/user_created_posts.html")
	default:
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, "./static/templates/error.html")
	}
}

func main() {
	// Static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Use the custom router for all other paths
	http.HandleFunc("/", router)

	log.Println("Serving on http://localhost:8081/")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}