package core

type errorConstant string

func (e errorConstant) Error() string { return string(e) }

// DisconnectedError represents Error when a peer is disconnected
const DisconnectedError = errorConstant("disconnected")

// ErrInvalidNetwork represents an Error when an invalid network was specified
const ErrInvalidNetwork = errorConstant("invalid network")

// ErrPipeStopProcessing is returned when the pipe has stopped it execution
var ErrPipeStopProcessing = errorConstant("pipe stopped")
