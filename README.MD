curl --location 'http://book-foto-art-back-app-1:8080/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "i.plc.rus@gmail.com",
    "password": "Shalimov123!!"
}'

curl --location 'http://localhost:8080/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "i.plc.rus@gmail.com",
    "password": "Shalimov123!!"
}'