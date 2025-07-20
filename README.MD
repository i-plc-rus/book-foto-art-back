curl --location 'http://book-foto-art-back-app-1:8080/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "i.plc.rus@gmail.com",
    "password": "Shalimov123!!"
}'

docker exec -it book-foto-art-back-app-1 \
curl --location 'http://book-foto-art-back-app-1:8080/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "i.plc.rus@gmail.com",
    "password": "Shalimov123!!"
}'