# Instructions

## Guest view
curl  http://localhost:8080/forum/api/guest

## Get categories
curl http://localhost:8080/forum/api/categories

## Register a new user:

curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

## Login

curl -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -c cookies.txt

## Logout

curl -X POST http://localhost:8080/forum/api/session/logout \
  -b cookies.txt


## Front

fetch("http://localhost:8080/forum/api/session/login", {
    method: "POST",
    credentials: "include", // IMPORTANT
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        email: "user@example.com",
        password: "supersecret"
    })
})



 ## Create a Post

 curl -X POST http://localhost:8080/forum/api/posts \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"category_id":1,"title":"My first post","content":"Hello forum!"}'

  curl -X POST http://localhost:8080/forum/api/posts/create \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"title":"My first TITLE","content":"Hello new forum!","category_ids":[1,2]}'

## Create a comment

curl -X POST http://localhost:8080/forum/api/comments \
  -H "Content-Type: application/json" \
  -d '{"post_id":"<POST_ID>","content":"Nice post!"}' \
  -b cookies.txt

## React to a post or comment

To like or dislike a post or comment you must be logged in. Use the ID of the
target post or comment along with the reaction type:

```
curl -X POST http://localhost:8080/forum/api/react \
  -H "Content-Type: application/json" \
  -d '{"target_id":"<TARGET_ID>","target_type":"post","reaction_type":1}' \
  -b cookies.txt

curl -X POST http://localhost:8080/forum/api/react \
  -H "Content-Type: application/json" \
  -d '{"target_id":"<TARGET_ID>","target_type":"comment","reaction_type":2}' \
  -b cookies.txt
```
Reaction type `1` represents a like and `2` represents a dislike. Running the
command again with the same parameters will toggle the reaction off.

# DOCKER

- On root directory

`
docker compose up --build
docker compose up
`

curl -X PUT http://localhost:8080/forum/api/posts/edit-title/b502e547-419a-453d-8037-b6672718964a \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=7c586478-5cbf-4148-8487-28308ed3b77a; csrf_token=109660a3824fcc81edbf9f0b78d30ed733bc147aa62aa6575b2221f01ea3a93d" \
  -H "X-CSRF-Token: 109660a3824fcc81edbf9f0b78d30ed733bc147aa62aa6575b2221f01ea3a93d" \
  -d '{"title":"Updated Title"}'

curl -X PUT http://localhost:8080/forum/api/posts/edit-content/b502e547-419a-453d-8037-b6672718964a \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=7c586478-5cbf-4148-8487-28308ed3b77a; csrf_token=109660a3824fcc81edbf9f0b78d30ed733bc147aa62aa6575b2221f01ea3a93d" \
  -H "X-CSRF-Token: 109660a3824fcc81edbf9f0b78d30ed733bc147aa62aa6575b2221f01ea3a93d" \
  -d '{"content":"My new content goes here."}'

  curl -X DELETE http://localhost:8080/forum/api/posts/delete/{b3378809-3aa6-45c1-a43a-555128c857d2} \
  -H "Cookie: session_id=b5f7815a-9990-4814-8f90-9a8a04cb0c29; csrf_token=69114c6d46778cd32e58f511a9ccca73aed94e6517320f8c0515aeaa411c5da5" \
  -H "X-CSRF-Token: 69114c6d46778cd32e58f511a9ccca73aed94e6517320f8c0515aeaa411c5da5"

curl -X PUT http://localhost:8080/forum/api/comments/edit/{ID} \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=YOUR_SESSION_ID" \
  -d '{"content":"Updated comment"}'

curl -X DELETE http://localhost:8080/forum/api/comments/delete/{ID} \
  -H "Cookie: session_id=YOUR_SESSION_ID"    


  curl -i -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{"email":"pat@pat.com","password":"pat123456"}'