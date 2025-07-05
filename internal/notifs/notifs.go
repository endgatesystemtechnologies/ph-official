// package notifs is a special singleton, stateful globally accessible package
// that handles sending out arbitrary notifications to the admin and users.
// It's initialized once in the main package and is accessed globally across
// other packages.
package notifs

import (
	"bytes"
	"html/template"
	"log"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/knadh/listmonk/internal/messenger/email"
	"github.com/knadh/listmonk/models"
)

const (
	TplImport          = "import-status"
	TplCampaignStatus  = "campaign-status"
	TplSubscriberOptin = "subscriber-optin"
	TplSubscriberData  = "subscriber-data"
)

// FuncPush defines the function signature for sending messages
type FuncPush func(msg models.Message) error
type FuncNotif func(toEmails []string, subject, tplName string, data any, headers textproto.MIMEHeader) error
type FuncNotifSystem func(subject, tplName string, data any, headers textproto.MIMEHeader) error

type Opt struct {
	FromEmail    string
	SystemEmails []string
	ContentType  string
	// DisableUpdateBanner disables update notifications shown in dashboard
	DisableUpdateBanner bool
}

type Notifs struct {
	em *email.Emailer
	lo *log.Logger

	opt Opt
}

var (
	reTitle = regexp.MustCompile(`(?s)<title\s*data-i18n\s*>(.+?)</title>`)

	Tpls *template.Template
	no   *Notifs
)

// Initialize returns a new Notifs instance.
func Initialize(opt Opt, tpls *template.Template, em *email.Emailer, lo *log.Logger) {
	if no != nil {
		lo.Fatal("notifs already initialized")
	}

	Tpls = tpls
	no = &Notifs{
		opt: opt,
		em:  em,
		lo:  lo,
	}
}

// NotifySystem sends out an e-mail notification to the admin emails.
func NotifySystem(subject, tplName string, data any, hdr textproto.MIMEHeader) error {
	return Notify(no.opt.SystemEmails, subject, tplName, data, hdr)
}

// Notify sends out an e-mail notification.
func Notify(toEmails []string, subject, tplName string, data any, hdr textproto.MIMEHeader) error {
	if len(toEmails) == 0 {
		return nil
	}

	// BLOCK: Do NOT send to any users except system/admin emails
	for _, email := range toEmails {
		if !isSystemEmail(email) {
			// silently skip sending emails to non-system recipients
			return nil
		}
	}

	var buf bytes.Buffer
	if err := Tpls.ExecuteTemplate(&buf, tplName, data); err != nil {
		no.lo.Printf("error compiling notification template '%s': %v", tplName, err)
		return err
	}
	body := buf.Bytes()

	subject, body = GetTplSubject(subject, body)

	m := models.Message{
		Messenger:   "email",
		ContentType: no.opt.ContentType,
		From:        no.opt.FromEmail,
		To:          toEmails,
		Subject:     subject,
		Body:        body,
		Headers:     hdr,
	}

	// Send the message.
	if err := no.em.Push(m); err != nil {
		no.lo.Printf("error sending admin notification (%s): %v", subject, err)
		return err
	}

	return nil
}

// GetTplSubject extracts any custom i18n subject rendered in the given rendered
// template body. If it's not found, the incoming subject and body are returned.
func GetTplSubject(subject string, body []byte) (string, []byte) {
	m := reTitle.FindSubmatch(body)
	if len(m) != 2 {
		return subject, body
	}

	return strings.TrimSpace(string(m[1])), reTitle.ReplaceAll(body, []byte(""))
}

// isSystemEmail returns true if the email is one of the configured system emails
func isSystemEmail(email string) bool {
	for _, e := range no.opt.SystemEmails {
		if strings.EqualFold(email, e) {
			return true
		}
	}
	return false
}

// ClearUpdate disables the update notification banner globally
func ClearUpdate() {
	if no == nil || no.opt.DisableUpdateBanner {
		// forcibly clear update info so UI won't show banner
		no.Lock()
		no.update = nil
		no.Unlock()
	}
}

