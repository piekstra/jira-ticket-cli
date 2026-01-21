package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetFields returns all field definitions
func (c *Client) GetFields() ([]Field, error) {
	urlStr := fmt.Sprintf("%s/field", c.BaseURL)
	body, err := c.get(urlStr)
	if err != nil {
		return nil, err
	}

	var fields []Field
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, fmt.Errorf("failed to parse fields: %w", err)
	}

	return fields, nil
}

// GetCustomFields returns only custom field definitions
func (c *Client) GetCustomFields() ([]Field, error) {
	fields, err := c.GetFields()
	if err != nil {
		return nil, err
	}

	var customFields []Field
	for _, f := range fields {
		if f.Custom {
			customFields = append(customFields, f)
		}
	}

	return customFields, nil
}

// FindFieldByName finds a field by name (case-insensitive)
func FindFieldByName(fields []Field, name string) *Field {
	nameLower := strings.ToLower(name)
	for i := range fields {
		if strings.ToLower(fields[i].Name) == nameLower {
			return &fields[i]
		}
	}
	return nil
}

// FindFieldByID finds a field by ID
func FindFieldByID(fields []Field, id string) *Field {
	for i := range fields {
		if fields[i].ID == id {
			return &fields[i]
		}
	}
	return nil
}

// ResolveFieldID resolves a field name or ID to its ID
func ResolveFieldID(fields []Field, nameOrID string) (string, error) {
	// First try exact ID match
	if f := FindFieldByID(fields, nameOrID); f != nil {
		return f.ID, nil
	}

	// Then try name match
	if f := FindFieldByName(fields, nameOrID); f != nil {
		return f.ID, nil
	}

	return "", fmt.Errorf("field not found: %s", nameOrID)
}

// FormatFieldValue formats a field value based on its type for the Jira API.
// It handles special cases like:
//   - option fields: wraps value as {"value": "..."}
//   - array fields: wraps value as [{"value": "..."}] or []string{...}
//   - user fields: wraps value as {"accountId": "..."}
//   - number fields: converts string to float64
//   - textarea custom fields: converts to ADF document
func FormatFieldValue(field *Field, value string) interface{} {
	if field == nil {
		return value
	}

	// Check for textarea custom fields that require ADF format
	if field.Schema.Custom == "com.atlassian.jira.plugin.system.customfieldtypes:textarea" {
		return NewADFDocument(value)
	}

	// Handle different field types
	switch field.Schema.Type {
	case "option":
		// Single select fields need {"value": "..."} format
		return map[string]string{"value": value}
	case "array":
		// Multi-select options need [{"value": "..."}] format
		if field.Schema.Items == "option" {
			return []map[string]string{{"value": value}}
		}
		// Other arrays (like labels) are just string arrays
		return []string{value}
	case "user":
		// User fields need {"accountId": "..."} format
		return map[string]string{"accountId": value}
	case "number":
		// Number fields need to be sent as JSON numbers, not strings
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			return n
		}
		return value
	default:
		return value
	}
}
