for i in {1..30}; do
    curl --request POST \
      --url http://localhost:8000/jobs/ \
      --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbi10eXBlIjoiVEVOQU5UX0VYQ0hBTkdFIiwiZW1haWwiOiJ0ZXN0QGdtYWlsLmNvbSIsIngtaGFzdXJhLWRlZmF1bHQtcm9sZSI6ImRheS1hZG1pbiIsIngtaGFzdXJhLWFsbG93ZWQtcm9sZXMiOlsiZGF5LWFkbWluIl0sIngtaGFzdXJhLXVzZXItaWQiOiI4ODY1MzY3OS1jNTA2LTQ4ZGEtOGNlNy1mOTljNTg4NjQxZTkiLCJ4LWhhc3VyYS10ZW5hbnQtaWQiOiIwODU2YjNiNy0wYTA4LTQ5NWQtYjQ4ZC1jODA1NjFmMGQ5ZTAiLCJiZWFjb24tYWRtaW4iOmZhbHNlLCJmaXJzdF9uYW1lIjoiUkFWSSIsImxhc3RfbmFtZSI6IlNBTktBUiIsImlhdCI6MTc1ODMzNTQxOCwiZXhwIjozNjE3NTgzMzU0MTh9.8mldjDjl3Vfj0PoNj656U8LpyOvGu9Yx6oIKtUXIZK8' \
      --header 'content-type: application/json' \
      --data '{"title":"Test","endpoint":"http://localhost:3000/","payload":"{}","method":"GET","scheduled_at":"2026-01-31T09:59:12.877Z"}' &
done
wait
