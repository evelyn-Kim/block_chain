const fs = require('fs');
var http = require('http');
var server = http.createServer();

var port = 3000;

server.listen( port, function() {
    console.log('Web server is started. : %d', port);
});

server.on( 'connection', function(socket) {
    var addr = socket.address();
    console.log('Client is coming. : %s %d', addr.address, addr.port);
});

server.on('request', function(req, res){
    console.log('Request has come.');
    //console.dir(req);

    // res.writeHead(200, {"Content-Type":"text/html; charset=utf-8"});
    // res.write('<!DOCTYPE hrml>')
    // res.write('<html>');
    // res.write('<head>');
    // res.write('<title> 응답 페이지 </title>');
    // res.write('</head>');
    // res.write('<body>');
    // res.write('<h1> 노드제이에스로부터의 응답페이지</h1>');
    // res.write('</body>');
    // res.write('</html>');
    // res.end();

    var filename = "./images/cat.jpg"
    fs.readFile(filename, function(err, data){
        res.writeHead(200, {"Content-Type":"image/jpeg"});
        res.write(data);
        res.end();
    });

});