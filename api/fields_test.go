package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestFields() []Field {
	return []Field{
		{ID: "summary", Name: "Summary", Custom: false},
		{ID: "description", Name: "Description", Custom: false},
		{ID: "customfield_10001", Name: "Story Points", Custom: true},
		{ID: "customfield_10002", Name: "Sprint", Custom: true},
		{ID: "customfield_10003", Name: "Epic Link", Custom: true},
	}
}

func TestFindFieldByName(t *testing.T) {
	fields := getTestFields()

	tests := []struct {
		name       string
		searchName string
		wantID     string
		wantNil    bool
	}{
		{
			name:       "exact match",
			searchName: "Summary",
			wantID:     "summary",
		},
		{
			name:       "case insensitive",
			searchName: "story points",
			wantID:     "customfield_10001",
		},
		{
			name:       "not found",
			searchName: "NonExistent",
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
			result := FindFieldByName(fields, tt.searchName)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantID, result.ID)
			}
		})
	}
}

func TestFindFieldByID(t *testing.T) {
	fields := getTestFields()

	tests := []struct {
		name     string
		searchID string
		wantName string
		wantNil  bool
	}{
		{
			name:     "exact match",
			searchID: "summary",
			wantName: "Summary",
		},
		{
			name:     "custom field",
			searchID: "customfield_10001",
			wantName: "Story Points",
		},
		{
			name:     "not found",
			searchID: "nonexistent",
			wantNil:  true,
		},
		{
			name:     "empty id",
			searchID: "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindFieldByID(fields, tt.searchID)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
			}
		})
	}
}

func TestResolveFieldID(t *testing.T) {
	fields := getTestFields()

	tests := []struct {
		name      string
		nameOrID  string
		wantID    string
		wantError bool
	}{
		{
			name:     "by exact ID",
			nameOrID: "summary",
			wantID:   "summary",
		},
		{
			name:     "by name",
			nameOrID: "Story Points",
			wantID:   "customfield_10001",
		},
		{
			name:     "by name case insensitive",
			nameOrID: "epic link",
			wantID:   "customfield_10003",
		},
		{
			name:      "not found",
			nameOrID:  "NonExistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ResolveFieldID(fields, tt.nameOrID)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}
}

func TestFindFieldByName_EmptySlice(t *testing.T) {
	result := FindFieldByName([]Field{}, "Summary")
	assert.Nil(t, result)
}

func TestFindFieldByID_EmptySlice(t *testing.T) {
	result := FindFieldByID([]Field{}, "summary")
	assert.Nil(t, result)
}

func TestResolveFieldID_EmptySlice(t *testing.T) {
	_, err := ResolveFieldID([]Field{}, "summary")
	assert.Error(t, err)
}
