# go2p
> golang p2p framework

By [v-braun - viktor-braun.de](https://viktor-braun.de).

[![](https://img.shields.io/github/license/v-braun/go2p.svg?style=flat-square)](https://github.com/v-braun/go2p/blob/master/LICENSE)
[![Build Status](https://img.shields.io/travis/v-braun/go2p.svg?style=flat-square)](https://travis-ci.org/v-braun/go2p)
[![codecov](https://codecov.io/gh/v-braun/go2p/branch/master/graph/badge.svg)](https://codecov.io/gh/v-braun/go2p)
[![Go Report Card](https://goreportcard.com/badge/github.com/v-braun/go2p)](https://goreportcard.com/report/github.com/v-braun/go2p)
[![Documentation](https://godoc.org/github.com/v-braun/go2p?status.svg)](http://godoc.org/github.com/v-braun/go2p)
![PR welcome](https://img.shields.io/badge/PR-welcome-green.svg?style=flat-square)


<p align="center">
<img width="70%" src="./idea/logo-1.svg" />
</p>


## Description

GO2P is a P2P framework, designed with flexibility and simplicity in mind. 
You can use a pre configured stack (encryption, compression, etc.) or built your own based on the existing modules. 

GO2P use the [middleware pattern](https://dzone.com/articles/understanding-middleware-pattern-in-expressjs) as a core pattern for messages. 
If you have used expressJS, OWIN or other HTTP/Web based frameworks you should be familiar with that.   
The basic idea is that an outgoing message is passed through multiple middleware functions. Each of this functions can manipulate the message.  
A middleware function could encrypt, compress, log or sign the message.  
Outgoing messages will be processed through the middleware functions and incomming messages in the inverted order:  

<p align="center">
<img width="70%" src="./idea/middleware-overview.svg" />
</p>




## Installation
```sh
go get github.com/v-braun/go2p
```



## Usage

```
todo...
```

## Configuration

```
todo...
```



## Authors

![image](https://avatars3.githubusercontent.com/u/4738210?v=3&amp;s=50)  
[v-braun](https://github.com/v-braun/)



## Contributing

Make sure to read these guides before getting started:
- [Contribution Guidelines](https://github.com/v-braun/go2p/blob/master/CONTRIBUTING.md)
- [Code of Conduct](https://github.com/v-braun/go2p/blob/master/CODE_OF_CONDUCT.md)

## License
**go2p** is available under the MIT License. See [LICENSE](https://github.com/v-braun/go2p/blob/master/LICENSE) for details.
