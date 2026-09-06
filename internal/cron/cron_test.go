package cron

import (
	"testing"

	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestScheduler_Init(t *testing.T) {
	sched := NewScheduler(Config{
		Repos: &repository.Repositories{},
	})
	assert.NotNil(t, sched)
	assert.Len(t, sched.cron.Entries(), 3)
}
