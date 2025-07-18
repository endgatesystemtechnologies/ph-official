// package notifs is a special singleton, stateful globally accessible package
// that handles sending out arbitrary notifications to the admin and users.
// It's initialized once in the main package and is accessed globally across
// other packages.
package notifs

import (
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

type FuncPush func(msg models.Message) error
type FuncNotif func(toEmails []string, subject, tplName string, data any, headers textproto.MIMEHeader) error
type FuncNotifSystem func(subject, tplName string, data any, headers textproto.MIMEHeader) error

type Opt struct {
	FromEmail    string
	SystemEmails []string
	ContentType  string
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
	// Notifications disabled. Instantly return, nothing sent.
	return nil
}

// Notify sends out an e-mail notification.
func Notify(toEmails []string, subject, tplName string, data any, hdr textproto.MIMEHeader) error {
	// Notifications disabled. Instantly return, nothing sent.
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

