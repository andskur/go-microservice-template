// Package version provides build and VCS metadata formatting.
package version

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// unspecified placeholder used to indicate unspecified values.
const unspecified = "unspecified"

// dateLayout is string layout for parsing datetime.
const dateLayout = "2006-01-02T15:04:05ZMST"

// verTemplate is template for version message.
const verTemplate = `{{.GetName}} {{.Tag}}
{{if .IfSpecified}}Branch {{.Branch}}, commit hash: {{.Commit}}
Origin repository: {{.URL}}
Compiled at: {{.Date}}
©{{.Date.Year}} {{end}}
`

// This variables supposed to be bound during compilation using -ldflags.
var (
	ServiceName  = unspecified
	CommitTag    = unspecified
	CommitSHA    = unspecified
	CommitBranch = unspecified
	OriginURL    = unspecified
	BuildDate    = unspecified
	Release      = unspecified
)

const defaultTag = "v0.0.0"

// Version represent git version structure.
type Version struct {
	Service string
	Tag     string
	Commit  string
	Branch  string
	URL     string
	Date    time.Time
	msg     bytes.Buffer
}

// NewVersion create new girt Version instance.
func NewVersion() (*Version, error) {
	var date time.Time

	if BuildDate == "unspecified" {
		date = time.Now()
	} else {
		parsedDate, parseErr := time.Parse(dateLayout, BuildDate)
		if parseErr != nil {
			return nil, fmt.Errorf("parse build date: %w", parseErr)
		}
		date = parsedDate
	}

	tag := CommitTag
	if tag == "" || tag == unspecified {
		tag = defaultTag
	}

	ver := &Version{
		Service: ServiceName,
		Tag:     tag,
		Commit:  CommitSHA,
		Branch:  CommitBranch,
		URL:     strings.TrimSuffix(OriginURL, ".git"),
		Date:    date,
	}

	if Release != unspecified {
		ver.Tag = Release
	}

	if err := ver.initTemplate(); err != nil {
		return nil, err
	}

	return ver, nil
}

// initTemplate initialize Version message from template.
func (v *Version) initTemplate() (err error) {
	tmpl, err := template.New("version").Parse(verTemplate)
	if err != nil {
		return fmt.Errorf("create version template: %w", err)
	}

	err = tmpl.Execute(&v.msg, v)
	return
}

// IfSpecified check if version specified.
func (v Version) IfSpecified() bool {
	if v.Service == unspecified || v.Commit == unspecified || v.Branch == unspecified || v.URL == unspecified {
		return false
	}
	return true
}

// String return version message.
func (v Version) String() string {
	return v.msg.String()
}

// GetName return service name.
func (v Version) GetName() string {
	return cases.Title(language.English).String(v.Service)
}
