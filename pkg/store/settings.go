package store

import (
	"sort"
	"strings"
)

// Tags are a list of strings with labels to declare
// for example backend capabilities
type Tags []string

// Eq checks if a list of tags is equal. Warning
// this mutates the list.
func (t Tags) Eq(other Tags) bool {
	sort.Strings(t)
	sort.Strings(other)
	return strings.Join(t, " ") == strings.Join(other, " ")
}

// BackendSettings hold per backend runtime configuration.
type BackendSettings struct {
	Tags Tags `json:"tags"`
}

// Merge with a partial update. Nil fields are ignored.
// If a field was updated this will return true
func (s *BackendSettings) Merge(update *BackendSettings) bool {
	updated := false
	if update.Tags != nil && !s.Tags.Eq(update.Tags) {
		s.Tags = update.Tags
		updated = true
	}
	return updated
}

// FrontendSettings hold all well known settings for a
// frontend.
type FrontendSettings struct {
	RequiredTags        Tags                         `json:"required_tags"`
	DefaultPresentation *DefaultPresentationSettings `json:"default_presentation"`
}

// Merge with a partial update. Fields that are nil
// will be ignored.
func (s *FrontendSettings) Merge(update *FrontendSettings) bool {
	updated := false
	if update.RequiredTags != nil && !s.RequiredTags.Eq(update.RequiredTags) {
		s.RequiredTags = update.RequiredTags
		updated = true
	}
	if update.DefaultPresentation != nil {
		s.DefaultPresentation = update.DefaultPresentation
		updated = true
	}
	return updated
}

// DefaultPresentationSettings configure a per frontend
// default configuration
type DefaultPresentationSettings struct {
	URL   string `json:"url,omitempty"`
	Force bool   `json:"force,omitempty"`
}
