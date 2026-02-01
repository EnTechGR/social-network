curl -s -X POST http://localhost:8080/api/v1/register \
  -F 'email=alice@example.com' \
  -F 'password=Pass1234' \
  -F 'first_name=Alice' \
  -F 'last_name=One' \
  -F 'date_of_birth=1990-01-01' \
  -F 'gender=female' \
  -F 'nickname=alice' \
  -F 'is_private=false' > ./alice_register.json

  curl -s -X POST http://localhost:8080/api/v1/register \
  -F 'email=bob@example.com' \
  -F 'password=Pass1234' \
  -F 'first_name=Bob' \
  -F 'last_name=Two' \
  -F 'date_of_birth=1991-02-02' \
  -F 'gender=male' \
  -F 'nickname=bob' \
  -F 'is_private=false' > ./bob_register.json

curl -s -X POST http://localhost:8080/api/v1/register \
  -F 'email=paul@example.com' \
  -F 'password=Pass1234' \
  -F 'first_name=Paul' \
  -F 'last_name=Three' \
  -F 'date_of_birth=1991-02-02' \
  -F 'gender=male' \
  -F 'nickname=paul' \
  -F 'is_private=false' > ./paul_register.json

curl -s -c ./alice.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"alice@example.com","password":"Pass1234"}' > alice_login.json

curl -s -c ./bob.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"bob@example.com","password":"Pass1234"}' > bob_login.json

curl -s -c ./paul.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"paul@example.com","password":"Pass1234"}' > paul_login.json


curl -i -b ./alice.cookies \
  -H "X-CSRF-Token: Z3MSt-g3dcg3D_Ru1F1fpzlDZnW-tRXfe9f6clSnkdI=" \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/follow \
  -d '{"followee_id":"c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f"}'

curl -i -b ./bob.cookies \
  -H "X-CSRF-Token: MVI5ZDZImTdIgmvt2JeeTJLCXl2y8uUdXQTg6dNv_o8=" \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/follow \
  -d '{"followee_id":"adc591b3-0fcc-43d1-bfcc-9b4bec9b80ae"}'

curl -i -b ./alice.cookies \
  -H "X-CSRF-Token: AWcyqjRX-0ZQFIUITk-PL2CpI27asTFoqXnCc-1DMdU=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/v1/user/privacy \
  -d '{"is_private": true}' 

curl -i -b ./bob.cookies \
  -H "X-CSRF-Token: MVI5ZDZImTdIgmvt2JeeTJLCXl2y8uUdXQTg6dNv_o8=" \
  -X DELETE http://localhost:8080/api/v1/followee/delete/adc591b3-0fcc-43d1-bfcc-9b4bec9b80ae

curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: 5x-CpcwbtBtWBIAq3aE5yMrYeyqmh6cfijvc5wGQedQ=" \
  -X DELETE http://localhost:8080/api/v1/followee/delete/c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f


curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: 5x-CpcwbtBtWBIAq3aE5yMrYeyqmh6cfijvc5wGQedQ=" \
  -X DELETE http://localhost:8080/api/v1/follower/delete/c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f

curl -i -b ./bob.cookies \
  -H "X-CSRF-Token: MVI5ZDZImTdIgmvt2JeeTJLCXl2y8uUdXQTg6dNv_o8=" \
  -X DELETE http://localhost:8080/api/v1/follower/delete/adc591b3-0fcc-43d1-bfcc-9b4bec9b80ae

curl -i -b ./bob1.cookies \
  -H "X-CSRF-Token: RxVbRfMHdpSEPxpGddFux_ZeeDQtNsOjjWfasZl1UVg=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/v1/user/privacy \
  -d '{"is_private": true}'

curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: 5x-CpcwbtBtWBIAq3aE5yMrYeyqmh6cfijvc5wGQedQ=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/v1/user/privacy \
  -d '{"is_private": true}'

curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: AWcyqjRX-0ZQFIUITk-PL2CpI27asTFoqXnCc-1DMdU=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/v1/follow/accept/b354fbb9-868c-4de7-b289-00a9d572b04d

curl -X POST http://localhost:8080/api/v1/groups/create \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: kKlTASjRe5j2Ka0iMkxtMCg9mHXlkpyDbVoKW-fOHm8=" \
  -b alice.cookies \
  -d '{"title": "Photography Club", "description": "For photography enthusiasts"}'

curl -X POST http://localhost:8080/api/v1/groups/create \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: gptJAu-_2nIsmZkKkG68mZYyM0hsJtXhTgN-sJF7hMU=" \
  -b alice.cookies \
  -d "{
    \"title\": \"Photography Club\",
    \"description\": \"For photography enthusiasts\",
    \"invitees\": [\"7d08e50f-ae3b-4b79-a352-fc2d4c883958\"]
  }"

curl -X PUT http://localhost:8080/api/v1/groups/invites/accept/ef2bcb2e-0187-47c6-95e9-653919770ab6 \
  -H "X-CSRF-Token: DKMB-9HPJSmZ5i1bVAJXnXJNbVlhKGVyRqNQzWQzVR0=" \
  -b bob.cookies

curl -X POST http://localhost:8080/api/v1/groups/invite/514a4942-be40-4fb0-851e-ed6c3d228b26 \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: DKMB-9HPJSmZ5i1bVAJXnXJNbVlhKGVyRqNQzWQzVR0=" \
  -b bob.cookies \
  -d "{\"user_id\": \"c6a8dc2c-6136-4068-8b1e-91b95e524905\"}"

curl -X POST http://localhost:8080/api/v1/groups/invite/514a4942-be40-4fb0-851e-ed6c3d228b26 \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: W9liVUS1ufD1UcNl_8pst8ve_xcDErbhMufmSYR2VIc=" \
  -b alice.cookies \
  -d "{\"user_id\": \"7d08e50f-ae3b-4b79-a352-fc2d4c883958\"}"

curl -X GET http://localhost:8080/api/v1/groups \
  -b bob.cookies

curl -X POST http://localhost:8080/api/v1/groups/request/ea7c7ac8-6df2-4ec9-afec-9785d5a9b36e \
  -H "X-CSRF-Token: DKMB-9HPJSmZ5i1bVAJXnXJNbVlhKGVyRqNQzWQzVR0=" \
  -b bob.cookies

curl -X GET http://localhost:8080/api/v1/groups/requests/ea7c7ac8-6df2-4ec9-afec-9785d5a9b36e \
  -b alice.cookies

curl -X PUT http://localhost:8080/api/v1/groups/requests/approve/7d6005d4-687e-40b9-953a-e34edd49d618 \
  -H "X-CSRF-Token: 6HoCMR82r-p0bLoe3SNQT5LZPuLi2y4iQ4ey7njJJSo=" \
  -b alice.cookies

curl -X GET http://localhost:8080/api/v1/groups/members/ea7c7ac8-6df2-4ec9-afec-9785d5a9b36e \
  -b alice.cookies