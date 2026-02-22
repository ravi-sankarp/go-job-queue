scheduled_at=$(date -u -d "+1 second" +"%Y-%m-%dT%H:%M:%S.%3NZ")

for i in {1..5}; do
  curl --request POST \
    --url http://localhost:8000/jobs/ \
    --header 'content-type: application/json' \
    --data "{\"title\":\"Test\",\"endpoint\":\"http://localhost:3000/\",\"payload\":\"{}\",\"method\":\"GET\",\"scheduled_at\":\"$scheduled_at\"}" &
done
wait
