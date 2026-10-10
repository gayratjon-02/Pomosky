// Package realtime implements the WebSocket layer: the message protocol, one
// RoomHub per live room, and the Client abstraction.
//
// The package does not depend on a WebSocket library. A transport adapter
// wraps the library's connection into the Conn interface (see conn.go), so the
// hub logic can be tested without any network.
package realtime
