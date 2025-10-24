package httpbox

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPathParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	req.SetPathValue("id", "123")

	param := NewPathParam(req, "id")

	assert.Equal(t, "123", param.String())
	assert.Equal(t, fromPath, param.from)
	assert.Equal(t, "id", param.name)
}

func TestNewPathParam_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)

	param := NewPathParam(req, "id")

	assert.Equal(t, "", param.String())
}

func TestNewQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users?name=john", nil)

	param := NewQueryParam(req, "name")

	assert.Equal(t, "john", param.String())
	assert.Equal(t, fromQuery, param.from)
	assert.Equal(t, "name", param.name)
}

func TestNewQueryParam_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	param := NewQueryParam(req, "name")

	assert.Equal(t, "", param.String())
}

func TestNewDefaultQueryParam(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		paramName    string
		defaultValue string
		expected     string
	}{
		{
			name:         "param exists",
			url:          "/users?limit=50",
			paramName:    "limit",
			defaultValue: "10",
			expected:     "50",
		},
		{
			name:         "param missing - uses default",
			url:          "/users",
			paramName:    "limit",
			defaultValue: "10",
			expected:     "10",
		},
		{
			name:         "param empty - uses default",
			url:          "/users?limit=",
			paramName:    "limit",
			defaultValue: "10",
			expected:     "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			param := NewDefaultQueryParam(req, tt.paramName, tt.defaultValue)

			assert.Equal(t, tt.expected, param.String())
		})
	}
}

func TestNewRequiredQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		paramName string
		expectErr bool
	}{
		{
			name:      "param exists",
			url:       "/users?id=123",
			paramName: "id",
			expectErr: false,
		},
		{
			name:      "param missing",
			url:       "/users",
			paramName: "id",
			expectErr: true,
		},
		{
			name:      "param empty",
			url:       "/users?id=",
			paramName: "id",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			param, err := NewRequiredQueryParam(req, tt.paramName)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "is required")
				assert.Contains(t, err.Error(), tt.paramName)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, param.String())
			}
		})
	}
}

func TestParam_String(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?value=hello", nil)
	param := NewQueryParam(req, "value")

	assert.Equal(t, "hello", param.String())
}

func TestParam_Int(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  int
		expectErr bool
	}{
		{"positive integer", "123", 123, false},
		{"negative integer", "-456", -456, false},
		{"zero", "0", 0, false},
		{"invalid - text", "abc", 0, true},
		{"invalid - float", "12.34", 0, true},
		{"invalid - empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?value="+tt.value, nil)
			param := NewQueryParam(req, "value")

			result, err := param.Int()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be an integer")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParam_Float(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  float64
		expectErr bool
	}{
		{"float with decimal", "3.14", 3.14, false},
		{"integer as float", "42", 42.0, false},
		{"negative float", "-2.5", -2.5, false},
		{"zero", "0", 0.0, false},
		{"scientific notation", "1e3", 1000.0, false},
		{"invalid - text", "abc", 0, true},
		{"invalid - empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?value="+tt.value, nil)
			param := NewQueryParam(req, "value")

			result, err := param.Float()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be a float")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParam_Bool(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  bool
		expectErr bool
	}{
		{"true lowercase", "true", true, false},
		{"TRUE uppercase", "TRUE", true, false},
		{"false lowercase", "false", false, false},
		{"FALSE uppercase", "FALSE", false, false},
		{"1 as true", "1", true, false},
		{"0 as false", "0", false, false},
		{"t as true", "t", true, false},
		{"f as false", "f", false, false},
		{"invalid - text", "yes", false, true},
		{"invalid - number", "2", false, true},
		{"invalid - empty", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?value="+tt.value, nil)
			param := NewQueryParam(req, "value")

			result, err := param.Bool()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be a boolean")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParam_Time(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		format    string
		expectErr bool
	}{
		{"RFC3339", "2023-01-15T10:30:00Z", time.RFC3339, false},
		{"date only", "2023-01-15", "2006-01-02", false},
		{"custom format", "15/01/2023", "02/01/2006", false},
		{"invalid format", "2023-01-15", time.RFC3339, true},
		{"invalid date", "invalid", time.RFC3339, true},
		{"empty", "", time.RFC3339, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?value="+tt.value, nil)
			param := NewQueryParam(req, "value")

			result, err := param.Time(tt.format)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be a time")
			} else {
				require.NoError(t, err)
				expected, _ := time.Parse(tt.format, tt.value)
				assert.Equal(t, expected, result)
			}
		})
	}
}

func TestParam_Duration(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  time.Duration
		expectErr bool
	}{
		{"milliseconds", "300ms", 300 * time.Millisecond, false},
		{"seconds", "5s", 5 * time.Second, false},
		{"minutes", "10m", 10 * time.Minute, false},
		{"hours", "2h", 2 * time.Hour, false},
		{"negative", "-1.5h", -90 * time.Minute, false},
		{"combined", "2h45m", 2*time.Hour + 45*time.Minute, false},
		{"invalid - text", "two hours", 0, true},
		{"invalid - number only", "300", 0, true},
		{"invalid - empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			q := req.URL.Query()
			q.Set("value", tt.value)
			req.URL.RawQuery = q.Encode()
			param := NewQueryParam(req, "value")

			result, err := param.Duration()

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "must be a time duration")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParam_PathParamAndQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/123?id=456", nil)
	req.SetPathValue("id", "123")

	pathParam := NewPathParam(req, "id")
	queryParam := NewQueryParam(req, "id")

	assert.Equal(t, "123", pathParam.String(), "path param should be 123")
	assert.Equal(t, "456", queryParam.String(), "query param should be 456")
	assert.Equal(t, fromPath, pathParam.from)
	assert.Equal(t, fromQuery, queryParam.from)
}

func TestParam_MultipleQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=golang&limit=10&sort=asc", nil)

	qParam := NewQueryParam(req, "q")
	limitParam := NewQueryParam(req, "limit")
	sortParam := NewQueryParam(req, "sort")

	assert.Equal(t, "golang", qParam.String())
	assert.Equal(t, "10", limitParam.String())
	assert.Equal(t, "asc", sortParam.String())

	limit, err := limitParam.Int()
	require.NoError(t, err)
	assert.Equal(t, 10, limit)
}

func TestParam_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"spaces", "hello%20world", "hello world"},
		{"plus sign", "hello+world", "hello world"},
		{"ampersand", "a%26b", "a&b"},
		{"equals", "a%3Db", "a=b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test?value="+tt.value, nil)
			param := NewQueryParam(req, "value")

			assert.Equal(t, tt.expected, param.String())
		})
	}
}

func TestParam_OverflowNumbers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	q := req.URL.Query()
	q.Set("int", "9223372036854775808")
	q.Set("float", "1.8e+308")
	req.URL.RawQuery = q.Encode()

	intParam := NewQueryParam(req, "int")
	floatParam := NewQueryParam(req, "float")

	_, err := intParam.Int()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")

	_, err = floatParam.Float()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a float")
}
