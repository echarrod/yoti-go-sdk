package attribute

import (
	"log"
	"time"

	"github.com/getyoti/yoti-go-sdk/v2/anchor"
	"github.com/getyoti/yoti-go-sdk/v2/yotiprotoattr"
)

// TimeAttribute is a Yoti attribute which returns a time as its value
type TimeAttribute struct {
	attributeDetails
	value *time.Time
}

// NewTime creates a new Time attribute
func NewTime(a *yotiprotoattr.Attribute) (*TimeAttribute, error) {
	parsedTime, err := time.Parse("2006-01-02", string(a.Value))
	if err != nil {
		log.Printf("Unable to parse time value of: %q. Error: %q", a.Value, err)
		parsedTime = time.Time{}
		return nil, err
	}

	parsedAnchors := anchor.ParseAnchors(a.Anchors)

	return &TimeAttribute{
		attributeDetails: attributeDetails{
			name:        a.Name,
			contentType: a.ContentType.String(),
			anchors:     parsedAnchors,
		},
		value: &parsedTime,
	}, nil
}

// Value returns the value of the TimeAttribute as *time.Time
func (a *TimeAttribute) Value() *time.Time {
	return a.value
}
