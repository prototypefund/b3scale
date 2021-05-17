package templates

import (
	"bytes"
	"html/template"
	// Use go16 embedding instead of inline templates
	_ "embed"
)

var (
	//go:embed html/redirect.html
	tmplRedirectHTML string

	//go:embed html/retry-join.html
	tmplRetryJoinHTML string

	//go:embed html/meeting-not-found.html
	tmplMeetingNotFoundHTML string

	//go:embed xml/default-presentation-body.xml
	tmplDefaultPresentationBodyXML string

	tmplRedirect                *template.Template
	tmplRetryJoin               *template.Template
	tmplMeetingNotFound         *template.Template
	tmplDefaultPresentationBody *template.Template
)

// Initialize templates
func init() {
	tmplRedirect, _ = template.New("redirect").Parse(tmplRedirectHTML)
	tmplRetryJoin, _ = template.New("retry_join").Parse(tmplRetryJoinHTML)
	tmplMeetingNotFound, _ = template.New("meeting_not_found").
		Parse(tmplMeetingNotFoundHTML)
	tmplDefaultPresentationBody, _ = template.New("default_presentation").
		Parse(tmplDefaultPresentationBodyXML)
}

// Redirect applies the redirect template
func Redirect(url string) []byte {
	res := new(bytes.Buffer)
	tmplRedirect.Execute(res, url)
	return res.Bytes()
}

// RetryJoin applies the retry join template
func RetryJoin(url string) []byte {
	res := new(bytes.Buffer)
	tmplRetryJoin.Execute(res, url)
	return res.Bytes()
}

// MeetingNotFound applies the meeting not found template
func MeetingNotFound() []byte {
	res := new(bytes.Buffer)
	tmplMeetingNotFound.Execute(res, nil)
	return res.Bytes()
}

// DefaultPresentationBody renders the xml body for
// a default presentation.
func DefaultPresentationBody(u, filename string) []byte {
	res := new(bytes.Buffer)
	tmplDefaultPresentationBody.Execute(res, struct{ URL, Filename string }{
		URL:      u,
		Filename: filename,
	})
	return res.Bytes()
}
