var express = require('express');
var expressErrorHandler = require('express-error-handler');

var cookieParser = require('cookie-parser');
var expressSession = require('express-session');

var path = require('path');
var fs = require('fs');

var app = express();

var errorHandler = expressErrorHandler({
    static: {
        '404': './public/404.html'
    }
});

app.set('port', process.env.PORT || 3000 );

app.use(expressSession({
    secret: 'blockmaster',
    resave:true,
    saveUninitialized:true
}));

app.use(cookieParser());

app.use(express.json());
app.use(express.urlencoded({extended:true}));

app.use(express.static(path.join(__dirname, 'public')));

app.use( function(req, res, next) {
    console.log("1st M/W");

    // res.writeHead(200, {'Content-Type':'text/html;charset=utf8'});
    // res.end('<h1>Express 서버에서 응답한 결과입니다 <br>미들웨어 TEST</h1>');
    next();
    
});

app.post( '/process/login', function(req, res) {
    console.log("2nd M/W");
    // res.send({name:'BSTUDENT', age:6});

    // var userAgent = req.header('User-Agent');
    // var paramName = req.query.name;
    var paramId = req.body.id;
    var paramPassword = req.body.password;

    if(req.session.user) {
        console.log('이미 로그인 되어 상품정보 페이지로 이동합니다.');

        res.redirect('/process/product');
        
    } else {
        //(TO DO) 검증코드 OK -> 아래부분, NOK면 -> 오류페이지 전송
        if( paramId == "bstudent" && paramPassword == "1234"){
            
            req.session.user = {
                id: paramId,
                name: paramId,
                authorized: true
            };
            
            res.writeHead(200, {'Content-Type':'text/html;charset=utf8'});
            res.write('<h1>Express 서버에서 응답한 결과입니다. </h1>');
            res.write('<div><p>Param name: '+paramId+ '</p></div>');
            res.write('<div><p>Param password: '+paramPassword+ '</p></div>');
            res.end();  
        }
        else {
            res.writeHead(401, {'Content-Type':'text/html;charset=utf8'});
            res.write('<h1>Express 서버에서 응답한 결과입니다. </h1>');
            res.write('<div><p>로그인정보가 올바르지 않습니다.</p></div>');
            res.end();  
        }
    }
});

app.get('/process/logout', (req, res) => {
    console.log('/process/logout called.');

    if(req.session.user) {
        console.log('로그아웃합니다.');

        req.session.destroy( function(err) {
            if(err) {throw err;}
            
            console.log('세션을 삭제하고 로그아웃합니다.');
            res.redirect('/login.html')
        });
    } else {
        console.log('아직 로그인되어있지 않습니다.');
        res.redirect('/login.html')
    }
});

app.get('/process/product', (req, res) => {
    console.log('/process/product called.');
    
    if(req.session.user) {
        res.status(200).sendFile(__dirname+'/product.html');
    } else {
        res.redirect('/login.html');
    }
});

app.get('/process/setUserCookie', (req, res) => {
    console.log('/process/setUserCookie called.');

    res.cookie( 'user', {
        id: 'bstudent',
        name: 'blockchain',
        authorized: true
    });

    res.redirect('/process/showCookie');
});

app.get('/process/showCookie', (req, res) => {
    console.log('/process/showCookie called.');

    res.send(req.cookies);
});

app.get('productlist', (req, res) => {
    console.log('productlist called.')

    var obj = JSON.parse('[{"name":"apple","price":1000},{"name":"mango","price":3500},{"name":"mandarin","price":3000}');
    res.status(200).send(obj);
})
app.use( expressErrorHandler.httpError(404) );
app.use( errorHandler )

app.listen(app.get('port'), function() {
    console.log(' Express server is started: ' + app.get('port'));
});