package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTransitionByName(t *testing.T) {
	transitions := []Transition{
		{ID: "11", Name: "To Do", To: Status{Name: "To Do"}},
		{ID: "21", Name: "In Progress", To: Status{Name: "In Progress"}},
		{ID: "31", Name: "Done", To: Status{Name: "Done"}},
	}

	tests := []struct {
		name       string
		searchName string
		wantID     string
		wantNil    bool
	}{
		{
			name:       "exact match",
			searchName: "In Progress",
			wantID:     "21",
		},
		{
			name:       "case insensitive",
			searchName: "in progress",
			wantID:     "21",
		},
		{
			name:       "uppercase",
			searchName: "DONE",
			wantID:     "31",
		},
		{
			name:       "not found",
			searchName: "Blocked",
			wantNil:    true,
		},
		{
			name:       "empty name",
			searchName: "",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindTransitionByName(transitions, tt.searchName)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantID, result.ID)
			}
		})
	}
}

func TestFindTransitionByName_EmptySlice(t *testing.T) {
	result := FindTransitionByName([]Transition{}, "In Progress")
	assert.Nil(t, result)
}

func TestFindTransitionByName_NilSlice(t *testing.T) {
	result := FindTransitionByName(nil, "In Progress")
	assert.Nil(t, result)
}
