package conversation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFSMStore(t *testing.T) {
	store := NewFSMStore(5 * time.Minute)

	sess := store.GetSession(100)
	assert.NotNil(t, sess)
	assert.Equal(t, StateNone, sess.State)

	store.SetState(100, StateWaitingVPNProtocol)
	sess = store.GetSession(100)
	assert.Equal(t, StateWaitingVPNProtocol, sess.State)

	store.Reset(100)
	sess = store.GetSession(100)
	assert.Equal(t, StateNone, sess.State)
}
