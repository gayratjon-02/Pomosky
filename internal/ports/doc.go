// Package ports declares the interfaces through which the application talks
// to the outside world (database, mail, time, randomness, broadcasting).
//
// Adapters live in store/postgres, platform/*, realtime and so on and are
// wired together only in internal/app.
package ports
