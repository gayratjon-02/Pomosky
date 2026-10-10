// Package room is the pure core of Pomosky.
//
// It contains the room State, the Events that can happen to it and the Reduce
// function that applies one to the other.
//
// Rules for this package (enforced by internal/archtest):
//
//   - It imports only the Go standard library.
//   - No clocks, randomness, goroutines, I/O or logging: time and IDs are
//     passed in as arguments.
//   - Every behavioural rule of a room lives here, never in controllers or
//     adapters.
package room
