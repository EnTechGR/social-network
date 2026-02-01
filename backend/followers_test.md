curl -s -X POST http://localhost:8080/api/v1/register \
  -F 'email=alice1@example.com' \
  -F 'password=Pass1234' \
  -F 'first_name=Alice1' \
  -F 'last_name=One1' \
  -F 'date_of_birth=1990-01-01' \
  -F 'gender=female' \
  -F 'nickname=alice1' \
  -F 'is_private=false' > ./alice_register.json

  curl -s -X POST http://localhost:8080/api/v1/register \
  -F 'email=bob1@example.com' \
  -F 'password=Pass1234' \
  -F 'first_name=Bob1' \
  -F 'last_name=Two1' \
  -F 'date_of_birth=1991-02-02' \
  -F 'gender=male' \
  -F 'nickname=bob1' \
  -F 'is_private=false' > ./bob_register.json

curl -s -c ./alice1.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"alice1@example.com","password":"Pass1234"}'

curl -s -c ./bob1.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"bob1@example.com","password":"Pass1234"}'


curl -i -b ./alice.cookies \
  -H "X-CSRF-Token: Z3MSt-g3dcg3D_Ru1F1fpzlDZnW-tRXfe9f6clSnkdI=" \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/follow \
  -d '{"followee_id":"c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f"}'

curl -i -b ./bob1.cookies \
  -H "X-CSRF-Token: S0XOOLHwvm4gUOycu1grPyO-zy3k1sY0PhImiXAAaUQ=" \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/follow \
  -d '{"followee_id":"e140e66d-4ce3-4727-b421-eb9ca1481605"}'

curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: AWcyqjRX-0ZQFIUITk-PL2CpI27asTFoqXnCc-1DMdU=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/v1/user/privacy \
  -d '{"is_private": true}'

curl -s -c ./bob.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"bob@example.com","password":"Pass1234"}' 

curl -i -b ./bob1.cookies \
  -H "X-CSRF-Token: RxVbRfMHdpSEPxpGddFux_ZeeDQtNsOjjWfasZl1UVg=" \
  -X DELETE http://localhost:8080/api/v1/followee/delete/01e97e7d-7a57-4df4-8623-e98977e5bead

curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: 5x-CpcwbtBtWBIAq3aE5yMrYeyqmh6cfijvc5wGQedQ=" \
  -X DELETE http://localhost:8080/api/v1/followee/delete/c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f


curl -i -b ./alice1.cookies \
  -H "X-CSRF-Token: 5x-CpcwbtBtWBIAq3aE5yMrYeyqmh6cfijvc5wGQedQ=" \
  -X DELETE http://localhost:8080/api/v1/follower/delete/c7f62be6-0cbd-4f46-bed5-dc27ef9ef49f

curl -i -b ./bob1.cookies \
  -H "X-CSRF-Token: Zehh8mtg-F9nwDEJKH2FNC7ZGp6FNZ9Qa-AHvo4lGTM=" \
  -X DELETE http://localhost:8080/api/v1/follower/delete/01e97e7d-7a57-4df4-8623-e98977e5bead

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