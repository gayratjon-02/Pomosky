package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/i18n"
	"github.com/gayratjon-02/Pomosky/internal/models"
	"github.com/gayratjon-02/Pomosky/internal/ports"
	"github.com/gayratjon-02/Pomosky/internal/room"
)

// Limits that keep memory and outgoing traffic bounded.
const (
	maxPerRoom   = 200
	maxTotal     = 20000
	maxSends     = 16 // concurrent requests to push services
	messageTTL   = 5 * 60
	sendTimeout  = 10 * time.Second
	maxFieldSize = 512
)

// Service implements ports.PushSubscriptions and ports.PhaseNotifier.
type Service struct {
	log     *slog.Logger
	pub     string
	priv    string
	subject string
	client  *http.Client
	// allowHost says whether a push service host may be called. The browser
	// supplies the address, so only the known push services are accepted.
	allowHost func(host string) bool

	slots chan struct{}

	mu    sync.Mutex
	rooms map[room.ID]map[string]models.PushSubscription // room -> endpoint -> subscription
	total int
}

var (
	_ ports.PushSubscriptions = (*Service)(nil)
	_ ports.PhaseNotifier     = (*Service)(nil)
)

// New creates the service. subject is a mailto: or https: contact.
func New(log *slog.Logger, publicKey, privateKey, subject string) *Service {
	return &Service{
		log:       log,
		pub:       publicKey,
		priv:      privateKey,
		subject:   subject,
		client:    &http.Client{Timeout: sendTimeout},
		allowHost: knownPushHost,
		slots:     make(chan struct{}, maxSends),
		rooms:     map[room.ID]map[string]models.PushSubscription{},
	}
}

// knownPushHost accepts the push services of the big browsers: Chrome/Edge/Opera
// (Google FCM), Firefox (Mozilla), Safari (Apple) and Windows (WNS).
func knownPushHost(host string) bool {
	host = strings.ToLower(host)
	for _, suffix := range []string{".googleapis.com", ".mozilla.com", ".push.apple.com", ".notify.windows.com"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

// PublicKey implements ports.PushSubscriptions.
func (s *Service) PublicKey() string { return s.pub }

// Subscribe implements ports.PushSubscriptions.
func (s *Service) Subscribe(id room.ID, sub models.PushSubscription) error {
	u, err := url.Parse(sub.Endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Hostname() == "" || !s.allowHost(u.Hostname()) {
		return fmt.Errorf("%w: unknown push service", apperr.ErrInvalid)
	}
	if len(sub.Endpoint) > 2048 || sub.P256dh == "" || sub.Auth == "" || len(sub.P256dh) > maxFieldSize || len(sub.Auth) > maxFieldSize {
		return fmt.Errorf("%w: bad subscription", apperr.ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.rooms[id]
	if _, known := subs[sub.Endpoint]; !known {
		if len(subs) >= maxPerRoom || s.total >= maxTotal {
			return fmt.Errorf("%w: too many subscriptions", apperr.ErrInvalid)
		}
		if subs == nil {
			subs = map[string]models.PushSubscription{}
			s.rooms[id] = subs
		}
		s.total++
	}
	subs[sub.Endpoint] = sub
	return nil
}

// Unsubscribe implements ports.PushSubscriptions.
func (s *Service) Unsubscribe(endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, subs := range s.rooms {
		s.removeLocked(id, subs, endpoint)
	}
}

func (s *Service) removeLocked(id room.ID, subs map[string]models.PushSubscription, endpoint string) {
	if _, ok := subs[endpoint]; !ok {
		return
	}
	delete(subs, endpoint)
	s.total--
	if len(subs) == 0 {
		delete(s.rooms, id)
	}
}

// Prune forgets the subscriptions of rooms that no longer exist.
func (s *Service) Prune(live func(room.ID) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, subs := range s.rooms {
		if !live(id) {
			s.total -= len(subs)
			delete(s.rooms, id)
		}
	}
}

// Count returns how many subscriptions are held (for tests and logs).
func (s *Service) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.total
}

// message is what the service worker receives.
type message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Room  string `json:"room"`
}

// EndMessage is the text of the "phase ended" alert in a language (the page shows the same one).
func EndMessage(lang i18n.Lang, phase room.Phase, label string) (title, body string) {
	label = strings.TrimSpace(label)
	if phase == room.PhaseBreak {
		return i18n.T(lang, "Break is over"), i18n.T(lang, "Ready for the next focus session?")
	}
	if label != "" {
		return i18n.T(lang, "Focus session finished"), strings.ReplaceAll(i18n.T(lang, "{label}. Time for a break."), "{label}", label)
	}
	return i18n.T(lang, "Focus session finished"), i18n.T(lang, "Time for a break.")
}

// PhaseEnded implements ports.PhaseNotifier. It returns at once; the messages
// are sent in the background.
func (s *Service) PhaseEnded(id room.ID, phase room.Phase, label string) {
	s.mu.Lock()
	subs := make([]models.PushSubscription, 0, len(s.rooms[id]))
	for _, sub := range s.rooms[id] {
		subs = append(subs, sub)
	}
	s.mu.Unlock()
	if len(subs) == 0 {
		return
	}
	for _, sub := range subs {
		title, body := EndMessage(i18n.Parse(sub.Lang), phase, label)
		payload, err := json.Marshal(message{Title: title, Body: body, Room: string(id)})
		if err != nil {
			continue
		}
		go s.send(id, sub, payload)
	}
}

func (s *Service) send(id room.ID, sub models.PushSubscription, payload []byte) {
	s.slots <- struct{}{}
	defer func() { <-s.slots }()

	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		HTTPClient:      s.client,
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.pub,
		VAPIDPrivateKey: s.priv,
		TTL:             messageTTL,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		s.log.Warn("push send failed", "room", id, "err", err)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		s.mu.Lock()
		s.removeLocked(id, s.rooms[id], sub.Endpoint) // the browser unsubscribed
		s.mu.Unlock()
	case resp.StatusCode >= 300:
		s.log.Warn("push service refused a message", "room", id, "status", resp.StatusCode)
	}
}
