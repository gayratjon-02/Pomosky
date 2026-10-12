package models

// PushSubscription is a browser's Web Push address: where to send a message
// and the keys that encrypt it. It is held in memory only, for as long as the
// room it belongs to is alive.
type PushSubscription struct {
	Endpoint string
	P256dh   string // the browser's public key, base64url
	Auth     string // the shared secret, base64url
	Lang     string // language of the browser's page, for the text of the alert
}
