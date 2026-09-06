package middleware

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v3"
)

type mockContext struct {
	tele.Context
	sender    *tele.User
	chat      *tele.Chat
	message   *tele.Message
	callback  *tele.Callback
	responded bool
}

func (m *mockContext) Sender() *tele.User {
	return m.sender
}

func (m *mockContext) Chat() *tele.Chat {
	return m.chat
}

func (m *mockContext) Message() *tele.Message {
	return m.message
}

func (m *mockContext) Callback() *tele.Callback {
	return m.callback
}

func (m *mockContext) Respond(resp ...*tele.CallbackResponse) error {
	m.responded = true
	return nil
}

func (m *mockContext) Notify(action tele.ChatAction) error {
	return nil
}

func (m *mockContext) Reply(what interface{}, opts ...interface{}) error {
	return nil
}

func TestLoggerMiddleware(t *testing.T) {
	// 1. Development mode logging
	mwDev := Logger(true)
	called := false
	handler := mwDev(func(c tele.Context) error {
		called = true
		return nil
	})

	ctx := &mockContext{
		sender:  &tele.User{ID: 12345, FirstName: "TestUser", Username: "test"},
		chat:    &tele.Chat{ID: 12345, Type: tele.ChatPrivate},
		message: &tele.Message{Text: "/start"},
	}

	err := handler(ctx)
	assert.NoError(t, err)
	assert.True(t, called)

	// Callback update
	cbCtx := &mockContext{
		sender:   &tele.User{ID: 12345, FirstName: "TestUser"},
		chat:     &tele.Chat{ID: 12345, Type: tele.ChatPrivate},
		callback: &tele.Callback{Data: "\fc_vpn", Unique: "c_vpn"},
	}
	err = handler(cbCtx)
	assert.NoError(t, err)

	// Error handling
	errHandler := mwDev(func(c tele.Context) error {
		return errors.New("boom")
	})
	err = errHandler(ctx)
	assert.Error(t, err)

	// 2. Production mode (isDev = false)
	mwProd := Logger(false)
	prodCalled := false
	prodHandler := mwProd(func(c tele.Context) error {
		prodCalled = true
		return nil
	})
	err = prodHandler(ctx)
	assert.NoError(t, err)
	assert.True(t, prodCalled)
}

func TestTypingMiddleware(t *testing.T) {
	mw := Typing()

	// For message, calls next
	called := false
	handler := mw(func(c tele.Context) error {
		called = true
		return nil
	})

	msgCtx := &mockContext{message: &tele.Message{Text: "hi"}}
	err := handler(msgCtx)
	assert.NoError(t, err)
	assert.True(t, called)

	// For callback, skips notify but calls next
	cbCalled := false
	cbHandler := mw(func(c tele.Context) error {
		cbCalled = true
		return nil
	})
	cbCtx := &mockContext{callback: &tele.Callback{Data: "c_vpn"}}
	err = cbHandler(cbCtx)
	assert.NoError(t, err)
	assert.True(t, cbCalled)
}

func TestPrivateOnlyMiddleware(t *testing.T) {
	mw := PrivateOnly()

	called := false
	handler := mw(func(c tele.Context) error {
		called = true
		return nil
	})

	// Private chat allows
	privCtx := &mockContext{chat: &tele.Chat{Type: tele.ChatPrivate}}
	err := handler(privCtx)
	assert.NoError(t, err)
	assert.True(t, called)

	// Group chat filters
	called = false
	groupCtx := &mockContext{chat: &tele.Chat{Type: tele.ChatGroup}}
	err = handler(groupCtx)
	assert.NoError(t, err)
	assert.False(t, called)
}

func TestAdminOnlyMiddleware(t *testing.T) {
	mw := AdminOnly(999)

	called := false
	handler := mw(func(c tele.Context) error {
		called = true
		return nil
	})

	// Non-admin
	ctx := &mockContext{sender: &tele.User{ID: 111}}
	err := handler(ctx)
	assert.NoError(t, err)
	assert.False(t, called)

	// Admin
	adminCtx := &mockContext{sender: &tele.User{ID: 999}}
	err = handler(adminCtx)
	assert.NoError(t, err)
	assert.True(t, called)
}