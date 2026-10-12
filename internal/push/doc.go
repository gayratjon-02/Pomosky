// Package push sends Web Push messages: a phone or browser that subscribed
// gets an alert when a phase of its room ends, also while the page is closed.
//
// Subscriptions live in memory only and are dropped when the room goes away
// (a server restart does the same to the rooms); browsers subscribe again on
// every visit. Nothing about them is stored in the database.
package push
