for i in {1..10}; do
curl -X POST http://localhost:8080/v1/campaigns \
-H "Content-Type: application/json" \
-d '{
  "name": "Campaign Test Chịu Tải '"$i"'",
  "content": "Test redis",
  "target_users": 5
}'
done
