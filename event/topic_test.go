package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hadroncorp/geck/event"
)

func TestNewTopic(t *testing.T) {
	type (
		inVals struct {
			org    string
			entity string
			action string
			opts   []event.TopicOption
		}
		expVals struct {
			strVal string
		}
	)

	tests := []struct {
		name string
		in   inVals
		exp  expVals
	}{
		{
			name: "empty",
			in:   inVals{},
			exp:  expVals{},
		},
		{
			name: "no platform",
			in: inVals{
				org:    "acme-corp",
				entity: "some-entity",
				action: "some-action",
				opts:   nil,
			},
			exp: expVals{
				strVal: "acme-corp.some-entity.some-action",
			},
		},
		{
			name: "with platform",
			in: inVals{
				org:    "acme-corp",
				entity: "some-entity",
				action: "some-action",
				opts: []event.TopicOption{
					event.WithPlatform("some-platform"),
				},
			},
			exp: expVals{
				strVal: "acme-corp.some-platform.some-entity.some-action",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := event.NewTopic(tt.in.org, tt.in.entity, tt.in.action, tt.in.opts...)
			assert.Equal(t, tt.exp.strVal, out.String())
		})
	}
}

func TestParseTopic(t *testing.T) {
	type (
		inVals struct {
			strVal string
		}
		expVals struct {
			org      string
			platform string
			entity   string
			action   string
			err      error
		}
	)

	tests := []struct {
		name string
		in   inVals
		exp  expVals
	}{
		{
			name: "empty",
			in:   inVals{},
			exp: expVals{
				err: event.ErrInvalidTopicName,
			},
		},
		{
			name: "no platform",
			in: inVals{
				strVal: "acme-corp.some-entity.some-action",
			},
			exp: expVals{
				org:    "acme-corp",
				entity: "some-entity",
				action: "some-action",
			},
		},
		{
			name: "with platform",
			in: inVals{
				strVal: "acme-corp.some-platform.some-entity.some-action",
			},
			exp: expVals{
				org:      "acme-corp",
				platform: "some-platform",
				entity:   "some-entity",
				action:   "some-action",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := event.ParseTopic(tt.in.strVal)
			assert.Equal(t, tt.exp.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.exp.org, out.Organization())
			assert.Equal(t, tt.exp.platform, out.Platform())
			assert.Equal(t, tt.exp.entity, out.Entity())
			assert.Equal(t, tt.exp.action, out.Action())
		})
	}
}
