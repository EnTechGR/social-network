# Post Privacy and Visibility Testing Guide

This document contains curl commands to test post creation, privacy settings, and visibility controls.

## Prerequisites
- Alice and Bob are registered users
- Cookies stored in: `alice1.cookies`, `bob1.cookies`
- CSRF tokens obtained from login responses or stored in `.txt` files

## Test Scenarios

### 1. Alice Creates a Public Post (Default)
By default, all posts are public and visible to all users.

```bash
curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -F "title=My Public Post" \
  -F "content=This is a public post visible to everyone!" \
  -F "image=@/home/entech/win_documents/social-network/3551739.jpg"
```

**Expected Response:**
```json
{
  "post": {
    "id": "post-id-here",
    "user_id": "alice-id",
    "visibility": "public",
    "title": "My Public Post",
    "content": "This is a public post visible to everyone!",
    "created_at": "2025-01-25T..."
  },
  "images_uploaded": 1,
  "message": "Post created with 1 image(s)"
}
```

---

### 2. Alice Creates a "Followers Only" Post
Bob will first try to view it (should fail), then follow Alice, then view it again (should succeed).

**Step 2.1: Alice creates a followers-only post**
```bash
curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -F "title=My Followers Post" \
  -F "content=Only my followers can see this!" \
  -F "visibility=followers"
```

Store the `post_id` from response:
```
FOLLOWERS_POST_ID="<post-id-from-response>"
```

**Step 2.2: Bob tries to view the post (should see 403 Forbidden)**
```bash
curl -X GET http://localhost:8080/api/v1/posts/$FOLLOWERS_POST_ID \
  -b bob1.cookies
```

Expected: 403 Forbidden or post not visible

**Step 2.3: Bob follows Alice**
```bash
curl -X POST http://localhost:8080/api/v1/follow/request \
  -H "X-CSRF-Token: $(cat bob.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b bob1.cookies \
  -H "Content-Type: application/json" \
  -d '{"followee_id": "alice-id"}'
```

**Step 2.4: Alice accepts Bob's follow request**
```bash
curl -X POST http://localhost:8080/api/v1/follow/accept \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"follower_id": "bob-id"}'
```

**Step 2.5: Bob now tries to view the post again (should succeed)**
```bash
curl -X GET http://localhost:8080/api/v1/posts/$FOLLOWERS_POST_ID \
  -b bob1.cookies
```

Expected: 200 OK with post data

---

### 3. Alice Creates a Private Post with Specific Allowed Users
Alice creates a private post and explicitly allows Bob to see it.

**Step 3.1: Alice creates a private post**
```bash
curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -F "title=My Secret Post" \
  -F "content=Only Bob can see this!" \
  -F "visibility=private"
```

Store the `post_id`:
```
PRIVATE_POST_ID="<post-id-from-response>"
```

**Step 3.2: Add Bob to the allowed users list**
```bash
curl -X POST http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID/allow-user \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"user_id": "bob-id"}'
```

**Step 3.3: Bob can now view the private post**
```bash
curl -X GET http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID \
  -b bob1.cookies
```

Expected: 200 OK with post data

**Step 3.4: Charlie (another user) tries to view - should fail**
```bash
curl -X GET http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID \
  -b charlie1.cookies
```

Expected: 403 Forbidden or post not visible

---

### 4. Alice Edits Post Visibility

**Step 4.1: Change public post to followers-only**
```bash
curl -X PUT http://localhost:8080/api/v1/posts/$PUBLIC_POST_ID/visibility \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"visibility": "followers"}'
```

Expected: 200 OK with status message

**Step 4.2: Change followers-only to private**
```bash
curl -X PUT http://localhost:8080/api/v1/posts/$FOLLOWERS_POST_ID/visibility \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"visibility": "private"}'
```

Expected: 200 OK

**Step 4.3: Change private back to public**
```bash
curl -X PUT http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID/visibility \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"visibility": "public"}'
```

Expected: 200 OK - now everyone can see the post again

---

### 5. Remove User from Allowed List

**Step 5.1: Remove Bob from private post's allowed users**
```bash
curl -X DELETE http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID/allow-user/bob-id \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'YOUR_CSRF_TOKEN')" \
  -b alice1.cookies
```

Expected: 200 OK with status message

**Step 5.2: Bob tries to view - should now fail**
```bash
curl -X GET http://localhost:8080/api/v1/posts/$PRIVATE_POST_ID \
  -b bob1.cookies
```

Expected: 403 Forbidden

---

## Complete Test Sequence Script

Save this as `test_post_privacy.sh`:

```bash
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function for curl requests
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local cookies=$4
    local csrf=$5
    
    if [ -z "$csrf" ]; then
        csrf=$(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' 2>/dev/null || echo "test-token")
    fi
    
    if [ -z "$data" ]; then
        curl -s -X $method "$url" -b "$cookies"
    else
        curl -s -X $method "$url" \
            -H "X-CSRF-Token: $csrf" \
            -H "Content-Type: application/json" \
            -b "$cookies" \
            -d "$data"
    fi
}

echo -e "${YELLOW}=== Post Privacy Testing ===${NC}\n"

# Test 1: Create public post
echo -e "${YELLOW}Test 1: Creating public post...${NC}"
PUBLIC_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'test')" \
  -b alice1.cookies \
  -F "title=Public Post" \
  -F "content=Visible to everyone!")

PUBLIC_POST_ID=$(echo $PUBLIC_RESPONSE | jq -r '.post.id')
echo -e "${GREEN}✓ Public post created: $PUBLIC_POST_ID${NC}\n"

# Test 2: Create followers-only post
echo -e "${YELLOW}Test 2: Creating followers-only post...${NC}"
FOLLOWERS_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'test')" \
  -b alice1.cookies \
  -F "title=Followers Only" \
  -F "content=Only followers can see!" \
  -F "visibility=followers")

FOLLOWERS_POST_ID=$(echo $FOLLOWERS_RESPONSE | jq -r '.post.id')
echo -e "${GREEN}✓ Followers post created: $FOLLOWERS_POST_ID${NC}\n"

# Test 3: Create private post
echo -e "${YELLOW}Test 3: Creating private post...${NC}"
PRIVATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'test')" \
  -b alice1.cookies \
  -F "title=Private Post" \
  -F "content=Only selected users!" \
  -F "visibility=private")

PRIVATE_POST_ID=$(echo $PRIVATE_RESPONSE | jq -r '.post.id')
echo -e "${GREEN}✓ Private post created: $PRIVATE_POST_ID${NC}\n"

# Test 4: Update visibility
echo -e "${YELLOW}Test 4: Updating post visibility...${NC}"
curl -s -X PUT http://localhost:8080/api/v1/posts/$PUBLIC_POST_ID/visibility \
  -H "X-CSRF-Token: $(cat alice.txt | grep -oP 'X-CSRF-Token: \K[^"]+' || echo 'test')" \
  -b alice1.cookies \
  -H "Content-Type: application/json" \
  -d '{"visibility": "followers"}' | jq '.'
echo -e "${GREEN}✓ Visibility updated${NC}\n"

echo -e "${YELLOW}=== All tests complete ===${NC}"
```

Make it executable and run:
```bash
chmod +x test_post_privacy.sh
./test_post_privacy.sh
```

---

## Notes

- Replace `alice-id`, `bob-id`, `charlie-id` with actual user IDs from registration
- Replace CSRF tokens with actual values from login responses
- The `post_id` format is UUID (e.g., `7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d`)
- Visibility options: `public`, `followers`, `private`
- Private posts with no allowed users are only visible to the creator
- Follower posts require an accepted follow relationship
