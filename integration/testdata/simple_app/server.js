const http = require('http');
const https = require('https');
const port = process.env.PORT || 8080;

const server = http.createServer((request, response) => {
  switch (request.url) {
    case '/test-openssl-ca':
      https.get('https://google.com', (res) => {
        res.setEncoding('utf8');
        let data = '';
        res.on('data', (chunk) => { data += chunk; });
        res.on('end', () => { response.end(data); });
      }).on('error', (e) => {
        response.end(e.toString());
      });
      break;

    case '/version':
      response.end(process.version);
      break;

    default:
      response.end("hello world");
  }
});

server.listen(port, (err) => {
    if (err) {
        return console.log('something bad happened', err);
    }

    console.log(`server is listening on ${port}`);
});
