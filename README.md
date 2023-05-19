# Timers Scheduler demo
Demo project that exposes API to create timers which invoke a URL after specified time passed.
Using mysql DB with [ent](https://entgo.io/) as ORM, and rabbitMQ as queue to proccess the messages.

It works by storing each time due date in DB, and run local cron every second that
checks which timers should run now, and adds them to a queue for processing.


## Run locally
```bash
docker-compose up -d
go run main.go
```

## Usage
Create a new timer by issuing a POST request to `localhost:8081/timers`.

We have a dummy endpoint `/test-webhook` to simulate successful webhook call.
For example if we want to add a new timer that will call the dummy URL after 2 minutes and 10 seconds:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"hours":0,"minutes":2,"seconds":10,"url":"http://localhost:8081/test-webhook"}' \
  http://localhost:8081/timers
```
Success response return the id of the timer:
```JSON
{
  "id": 5
}
```

Get the time left of a specific timer by issuing a GET request to `localhost:8081/timers/:id`, success response will 
contain the id and time left in seconds, for example:
```bash
curl http://localhost:8081/timers/5
```
Success response:
```JSON
{
  "id": 5,
  "time_left": 246
}
```

## Run tests
`make tests`

## Change default config
You can change default config by creating `.env` file
and set the values from `.env.example`
```text
PORT=
QUEUE_NAME=
MYSQL_CONNECTION=
RABBITMQ_CONNECTION=
```