package go2p

import "errors"

// InvalidNetworkError represents an invalid network part in the given address
var InvalidNetworkError = errors.New("invalid network")

// PeerOperator connect peers to the current network connection
// I provides functionalities for dialing (active connection)
// and listening (passive connections) over a protocol (tcp/udp/etc)
type PeerOperator interface {

	// Dial connects to the given address by the given network
	Dial(network string, addr string) error

	// OnPeer registers a handler function that should be called
	// when a new peer connection is established
	OnPeer(handler func(p Adapter))

	// Start the background listening jobs for the operator
	Start() error

	// Stop the background listening jobs for the operator
	Stop()
}
