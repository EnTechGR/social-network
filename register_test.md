## Test 1: Valid Registration ✅
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "SecurePass123",
    "first_name": "John",
    "last_name": "Doe",
    "age": 25,
    "gender": "male"
  }'
```
### Expected Response (Status 201):
```bash
json{
  "user": {
    "id": "some-uuid",
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "age": 25,
    "gender": "male",
    "created_at": "2025-11-04T21:30:00Z"
  },
  "session_id": "session-uuid",
  "csrf_token": "csrf-token-here"
}
```

## Test 2: Missing Required Fields ❌
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "janedoe",
    "email": "jane@example.com",
    "password": "SecurePass456"
  }'
```
### Expected Response (Status 400):
```bash
json{
  "error": "All fields are required: username, email, password, first_name, last_name, age, gender"
}
```
## Test 3: Invalid Age ❌
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "younguser",
    "email": "young@example.com",
    "password": "SecurePass789",
    "first_name": "Young",
    "last_name": "User",
    "age": 10,
    "gender": "male"
  }'
```
### Expected Response (Status 400):
```bash
json{
  "error": "Age must be between 13 and 120"
}
```
## Test 4: Invalid Gender ❌
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass000",
    "first_name": "Test",
    "last_name": "User",
    "age": 25,
    "gender": "invalid_gender"
  }'
```
### Expected Response (Status 400):
```bash
json{
  "error": "Gender must be one of: male, female, other, prefer_not_to_say"
}
```
## Test 5: Duplicate Email ❌
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe2",
    "email": "john@example.com",
    "password": "SecurePass123",
    "first_name": "John",
    "last_name": "Doe",
    "age": 30,
    "gender": "male"
  }'
```
### Expected Response (Status 409):
```bash
json{
  "error": "Email is already taken"
}
```
## Test 6: Duplicate Username ❌
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john2@example.com",
    "password": "SecurePass123",
    "first_name": "John",
    "last_name": "Doe",
    "age": 30,
    "gender": "male"
  }'
```  
### Expected Response (Status 409):
```bash
json{
  "error": "Username is already taken"
}
```
## Step 3: Test Login Endpoint
- First, register a test user:
```bash
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alicetest",
    "email": "alice@example.com",
    "password": "AlicePass123",
    "first_name": "Alice",
    "last_name": "Smith",
    "age": 28,
    "gender": "female"
  }'
```
## Test 7: Login with Email ✅
```bash
curl -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "alice@example.com",
    "password": "AlicePass123"
  }'
```
### Expected Response (Status 200):
```bash
json{
  "user": {
    "id": "some-uuid",
    "username": "alicetest",
    "email": "alice@example.com",
    "first_name": "Alice",
    "last_name": "Smith",
    "age": 28,
    "gender": "female",
    "created_at": "2025-11-04T21:35:00Z"
  },
  "session_id": "session-uuid",
  "csrf_token": "csrf-token-here"
}
```
## Test 8: Login with Username ✅
```bash
curl -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "alicetest",
    "password": "AlicePass123"
  }'
```
### Expected Response (Status 200):
```bash
json{
  "user": {
    "id": "some-uuid",
    "username": "alicetest",
    "email": "alice@example.com",
    "first_name": "Alice",
    "last_name": "Smith",
    "age": 28,
    "gender": "female",
    "created_at": "2025-11-04T21:35:00Z"
  },
  "session_id": "session-uuid",
  "csrf_token": "csrf-token-here"
}
```
## Test 9: Login with Wrong Password ❌
```bash
curl -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "alicetest",
    "password": "WrongPassword"
  }'
```
### Expected Response (Status 401):
```bash
json{
  "error": "Invalid username/email or password"
}
```
## Test 10: Login with Non-existent User ❌
```bash
curl -X POST http://localhost:8080/forum/api/session/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "nonexistent@example.com",
    "password": "SomePassword123"
  }'
```
### Expected Response (Status 401):
```bash
json{
  "error": "Invalid username/email or password"
}
```
## Step 4: Verify Database
- Check that the data was stored correctly:
```bash
sqlite3 ./database/forum.db
```
Then run:
sql-- Check if migration ran
SELECT * FROM database_version ORDER BY version DESC LIMIT 1;
-- Should show version 9

-- Check user table structure
PRAGMA table_info(user);
-- Should show: user_id, username, email, first_name, last_name, age, gender, created_at

-- Check registered users
SELECT user_id, username, email, first_name, last_name, age, gender FROM user;
-- Should show all registered users with all fields

-- Exit sqlite
.exit

## Step 5: Test All Gender Options
```bash
# Test male
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user1", "email": "user1@example.com", "password": "Pass1234", "first_name": "User", "last_name": "One", "age": 25, "gender": "male"}'

# Test female
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user2", "email": "user2@example.com", "password": "Pass1234", "first_name": "User", "last_name": "Two", "age": 30, "gender": "female"}'

# Test other
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user3", "email": "user3@example.com", "password": "Pass1234", "first_name": "User", "last_name": "Three", "age": 22, "gender": "other"}'

# Test prefer_not_to_say
curl -X POST http://localhost:8080/forum/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user4", "email": "user4@example.com", "password": "Pass1234", "first_name": "User", "last_name": "Four", "age": 35, "gender": "prefer_not_to_say"}'
```