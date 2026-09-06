package conversation

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/cloudflare"
	tele "gopkg.in/telebot.v3"
)

var wildcardDomainRegex = regexp.MustCompile(`^.+\.\D+$`)

// WildcardWizard manages wildcard domain registrations.
type WildcardWizard struct {
	userRepo     repository.UserRepository
	serverRepo   repository.ServerRepository
	wildcardRepo repository.WildcardRepository
	cfClient     *cloudflare.Client
	store        *FSMStore
}

// NewWildcardWizard creates a new WildcardWizard.
func NewWildcardWizard(
	userRepo repository.UserRepository,
	serverRepo repository.ServerRepository,
	wildcardRepo repository.WildcardRepository,
	cfClient *cloudflare.Client,
	store *FSMStore,
) *WildcardWizard {
	return &WildcardWizard{
		userRepo:     userRepo,
		serverRepo:   serverRepo,
		wildcardRepo: wildcardRepo,
		cfClient:     cfClient,
		store:        store,
	}
}

// StartWizard checks donator status and prompts domain.
func (w *WildcardWizard) StartWizard(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := w.userRepo.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun belum terdaftar", ShowAlert: true})
	}

	if time.Now().After(user.Expired) {
		return c.Respond(&tele.CallbackResponse{
			Text:      "YDDA\nYang Donasi Donasi Ajah :>",
			ShowAlert: true,
		})
	}

	w.store.SetState(c.Sender().ID, StateWaitingWildcardDomain)

	_ = c.Respond()
	return c.EditCaption("OK, mau domain apa ?\n\ncontoh: zoom.us")
}

// HandleTextMessage handles the text message with the wildcard domain name.
func (w *WildcardWizard) HandleTextMessage(c tele.Context) error {
	domainName := strings.ToLower(strings.TrimSpace(c.Message().Text))

	if !wildcardDomainRegex.MatchString(domainName) {
		w.store.Reset(c.Sender().ID)
		return c.Reply("Waduh error!\nProses batal!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	exists, err := w.wildcardRepo.Exists(ctx, domainName)
	if err != nil || exists {
		w.store.Reset(c.Sender().ID)
		return c.Reply("Wildcard sudah ada!\nProses batal!")
	}

	_ = c.Reply("OK proses...")

	if err := w.wildcardRepo.Create(ctx, domainName); err != nil {
		w.store.Reset(c.Sender().ID)
		return c.Reply("Gagal menyimpan wildcard ke database!\nProses batal!")
	}

	servers, _ := w.serverRepo.GetAll(ctx)
	wildcards, _ := w.wildcardRepo.GetAll(ctx)
	_ = w.cfClient.RegisterWorkersSubdomains(ctx, servers, wildcards)

	w.store.Reset(c.Sender().ID)
	return c.Reply("Done!")
}
