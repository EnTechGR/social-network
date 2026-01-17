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
  -H "X-CSRF-Token: Qdbs_g_sWPhOYcMZFY3_TQvVvB-g9p4Md9Vhgsg8qJo=" \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/follow \
  -d '{"followee_id":"ce8715ad-27f1-409e-8261-8f9d68a36221"}'

curl -i -b ./alice.cookies \
  -H "X-CSRF-Token: 1BGnF2882u5tyx3Rpn_99G2zVwaH-zvMXurbS0ITQ3k=" \
  -H 'Content-Type: application/json' \
  -X PUT http://localhost:8080/api/user/privacy \
  -d '{"is_private": true}'

curl -s -c ./bob.cookies \
  -H 'Content-Type: application/json' \
  -X POST http://localhost:8080/api/v1/login \
  -d '{"login":"bob@example.com","password":"Pass1234"}' 

curl -i -b ./bob.cookies \
  -H "X-CSRF-Token: a_jgfUyRS_BjRxTJaY_5CflkYi4vGeMSiYSpp5XhLsM=" \
  -X DELETE http://localhost:8080/api/v1/follow/delete/b9abfb51-245c-4621-9885-1865ccc2aeb5
