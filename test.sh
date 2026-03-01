
for i in {1..10000}; do
  scheduled_at=$(date -u -d "+${i} seconds" +"%Y-%m-%dT%H:%M:%S.%3NZ")
  curl --request POST \
    --url http://localhost:8000/jobs/ \
    --header 'content-type: application/json' \
    --data "{\"title\":\"Test\",\"endpoint\":\"http://localhost:3000/\",\"payload\":\"{}\",\"method\":\"GET\",\"scheduled_at\":\"$scheduled_at\"}" &
done
wait
