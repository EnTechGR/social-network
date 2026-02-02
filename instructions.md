# Instructions

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


## Create a Post

 curl -s -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: 4ysqu61r4Djik2jPBrhUyvWta_HYVHVZxJs1j6SCd20=" \
  -H "Content-Type: application/json" \
  -b alice.cookies \
  -d '{
    "title": "My First Post",
    "content": "This is the content of my post"
  }'

curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: 4ysqu61r4Djik2jPBrhUyvWta_HYVHVZxJs1j6SCd20=" \
  -b alice.cookies \
  -F "title=My Post with Images" \
  -F "content=Check out these amazing pictures" \
  -F "image=@./3551739.jpg"

curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: 4ysqu61r4Djik2jPBrhUyvWta_HYVHVZxJs1j6SCd20=" \
  -b alice.cookies \
  -F "title=Followers only" \
  -F "content=Only followers can see" \
  -F "visibility=followers"

curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: 4ysqu61r4Djik2jPBrhUyvWta_HYVHVZxJs1j6SCd20=" \
  -b alice.cookies \
  -F "title=Private Post" \
  -F "content=Only selected users" \
  -F "visibility=private"

curl -X PUT http://localhost:8080/api/v1/posts/update-visibility/c5fd747f-42f4-4851-b486-0ec489d43c05 \
  -H "X-CSRF-Token: 4ysqu61r4Djik2jPBrhUyvWta_HYVHVZxJs1j6SCd20=" \
  -b alice.cookies \
  -H "Content-Type: application/json" \
  -d '{"visibility": "followers"}'

## Create a comment

curl -s -X POST http://localhost:8080/api/v1/comments/create \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: S0XOOLHwvm4gUOycu1grPyO-zy3k1sY0PhImiXAAaUQ=" \
  -b bob1.cookies \
  -d '{
    "post_id": "6651178e-fcf5-4443-9687-d328b7a508b4",
    "content": "Great post, Alice! Love the pictures!"
  }'

curl -s -X POST http://localhost:8080/api/v1/comments/create \
  -H "X-CSRF-Token: -ZwK3QNUxESO-br7eV1s6C_JnmLDiJhnFJHikHy3maU=" \
  -b bob1.cookies \
  -F "post_id=0a5aba7f-40ef-4675-abaf-3fdfa660e2bd" \
  -F "content=Great post Alice! Here's my feedback image:" \
  -F "image=@/home/entech/win_documents/social-network/3551739.jpg"
  
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


# Add allowed users

curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: KPWe4foeCTjMm0RHYwfq3gXNKFO7JRXb_hnAOXZM9XY=" \
  -b alice.cookies \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"Private Post for Bob\",
    \"content\": \"This post is private and only Bob can see it\",
    \"visibility\": \"private\",
    \"allowed_user_ids\": [\"7d08e50f-ae3b-4b79-a352-fc2d4c883958\"]
  }"

curl -X POST http://localhost:8080/api/v1/posts/remove-allowed-user/cf24575c-f69e-40ec-b0bb-647d2f8d0ef3 \
  -H "X-CSRF-Token: KPWe4foeCTjMm0RHYwfq3gXNKFO7JRXb_hnAOXZM9XY=" \
  -b alice.cookies \
  -H "Content-Type: application/json" \
  -d '{"user_id": "7d08e50f-ae3b-4b79-a352-fc2d4c883958"}'

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


  curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: nZDzwtvdZ3I0Ppab-M0_AMouafxmOVo0Ffc-Fe4QRAs=" \
  -H "Content-Type: application/json" \
  -d '{"category_id":1,"title":"My first post","content":"Hello forum!"}'