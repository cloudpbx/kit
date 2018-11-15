# AMQP Transport Example

Example publisher and subscriber using AMQP transport.
Since AMQP transport requires me to call several AMQP functions, and since using publisher and subscriber requires me to export `StringService`, I am putting this example in a subfolder for now. I am more than happy to integrate this to the original `stringsvc4` folder.

## Run
`go run subscriber/main.go`
`go run publisher/main.go`

### Comments
The above commands assume that there is an AMQP server running in `amqp://localhost:5672`.