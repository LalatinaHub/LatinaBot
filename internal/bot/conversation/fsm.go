package conversation

import (
	"sync"
	"time"
)

// State represents the current dialog state of a user.
type State string

const (
	StateNone                  State = ""
	StateWaitingVPNProtocol    State = "waiting_vpn_protocol"
	StateWaitingVPNServer      State = "waiting_vpn_server"
	StateWaitingVPNRelay       State = "waiting_vpn_relay"
	StateWaitingVPNConfirm     State = "waiting_vpn_confirm"
	StateWaitingWildcardDomain State = "waiting_wildcard_domain"
)

// VPNSession holds temporary data while user is configuring a VPN profile.
type VPNSession struct {
	VPN        string   `json:"vpn"`
	ServerCode string   `json:"server_code"`
	Relay      string   `json:"relay"`
	Page       int      `json:"page"`
	RelaysCC   []string `json:"-"`
}

// Session holds state and ephemeral data for a user.
type Session struct {
	UserID     int64
	State      State
	VPNData    VPNSession
	LastActive time.Time
}

// FSMStore is a thread-safe in-memory session manager with TTL.
type FSMStore struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewFSMStore creates a new FSM session store.
func NewFSMStore(ttl time.Duration) *FSMStore {
	store := &FSMStore{
		sessions: make(map[int64]*Session),
		ttl:      ttl,
	}

	// Background routine to clean expired sessions every 2 minutes
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			store.cleanup()
		}
	}()

	return store
}

// GetSession returns session for userID, creating a new one if not found.
func (s *FSMStore) GetSession(userID int64) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[userID]
	if !ok || time.Since(sess.LastActive) > s.ttl {
		sess = &Session{
			UserID:     userID,
			State:      StateNone,
			LastActive: time.Now(),
		}
		s.sessions[userID] = sess
	} else {
		sess.LastActive = time.Now()
	}

	return sess
}

// SetState updates the user's conversation state.
func (s *FSMStore) SetState(userID int64, state State) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[userID]
	if !ok {
		sess = &Session{UserID: userID}
		s.sessions[userID] = sess
	}
	sess.State = state
	sess.LastActive = time.Now()
}

// Reset clears the state and VPN temporary data for a user.
func (s *FSMStore) Reset(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, userID)
}

func (s *FSMStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, sess := range s.sessions {
		if now.Sub(sess.LastActive) > s.ttl {
			delete(s.sessions, id)
		}
	}
}
