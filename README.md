# go2p
> golang p2p framework

By [v-braun - viktor-braun.de](https://viktor-braun.de).

[![Build Status](https://img.shields.io/travis/v-braun/go2p.svg?style=flat-square)](https://travis-ci.org/v-braun/go2p)
[![codecov](https://codecov.io/gh/v-braun/go2p/branch/master/graph/badge.svg)](https://codecov.io/gh/v-braun/go2p)
[![Go Report Card](https://goreportcard.com/badge/github.com/v-braun/go2p)](https://goreportcard.com/report/github.com/v-braun/go2p)
[![Documentation](https://godoc.org/github.com/v-braun/go2p?status.svg)](http://godoc.org/github.com/v-braun/go2p)
![PR welcome](https://img.shields.io/badge/PR-welcome-green.svg?style=flat-square)
[![](https://img.shields.io/github/license/v-braun/go2p.svg?style=flat-square)](https://github.com/v-braun/go2p/blob/master/LICENSE)



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

> You like code? Checkout the [chat example](https://github.com/v-braun/go2p/blob/master/examples/chat/main.go)

The simplest way to use this framework is to create a new instance of the full configured TCP based network stack:

``` go
    localAddr := "localhost:7077"
	net := go2p.NewNetworkConnectionTCP(*localAddr, &map[string]func(peer *go2p.Peer, msg *go2p.Message){
		"msg": func(peer *go2p.Peer, msg *go2p.Message) {
			fmt.Printf("%s > %s\n", peer.RemoteAddress(), msg.PayloadGetString())
		},
    })
    
    net.OnPeer(func(p *go2p.Peer) {
		fmt.Printf("new peer: %s\n", p.RemoteAddress())
    })
    
    err := net.Start()
	if err != nil {
		panic(err)
    }

    defer net.Stop()
    

    // connects to another peer via tcp
    net.ConnectTo("tcp", "localhost:7077")

    // send a message to the 'msg' route 
    net.SendBroadcast(go2p.NewMessageRoutedFromString("msg", "hello go2p"))



```

## Advanced Usage

The function NewNetworkConnectionTCP is a shorthand for the advanced configuration of a network stack. 

``` go
	op := go2p.NewTCPOperator("tcp", localAddr) // creates a tcp based operator (net.Dialer and net.Listener)
	peerStore := go2p.NewDefaultPeerStore(10) // creates a simple peer store that limits connections to 10

	conn := go2p.NewNetworkConnection(). // creates a new instance of the builder
		WithOperator(op). // adds the operator to the network stack
		WithPeerStore(peerStore). // adds the peer store to the network stack
		WithMiddleware(go2p.Routes(routes)). // adds the routes middleware
		WithMiddleware(go2p.Headers()). // adds the headers middleware
		WithMiddleware(go2p.Crypt()). // adds encryption
		WithMiddleware(go2p.Log()). // adds logging
		Build() // creates the network 
```

This code creates a new NetworkConnection that use tcp communication, a default PeerStore and some middlewares.  
Outgoing messages will now pass the following middlewares:  
``` 
(app logic) -> Routing -> Headers -> Crypt -> Log -> (network) 
``` 

Incomming messages will pass the following middlewares  
``` 
(app logic) <- Routing <- Headers <- Crypt <- Log <- (network)
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
