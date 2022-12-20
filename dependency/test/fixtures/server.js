const http = require('http')
const port = process.argv[2]

const requestHandler = (request, response) => {
    response.end("hello world")
}

const server = http.createServer(requestHandler)

server.listen(port, (err) => {
    if (err) {
        return console.log('server failed to start', err)
    }

    console.log(`server is listening on ${port}`)
})
