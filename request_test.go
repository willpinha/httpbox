package httpbox

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
	Age   int    `json:"age" xml:"age"`
}

type validatingStruct struct {
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
	Age   int    `json:"age" xml:"age"`
}

func (v *validatingStruct) Validate() error {
	if v.Name == "" {
		return errors.New("name is required")
	}
	if v.Email == "" {
		return errors.New("email is required")
	}
	// Mutate age if it's negative
	if v.Age < 0 {
		v.Age = 0
	}
	return nil
}

func TestReadJSON_Success(t *testing.T) {
	jsonData := `{"name":"John","email":"john@example.com","age":30}`
	reader := strings.NewReader(jsonData)

	result, err := ReadJSON[testStruct](reader)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
	}{
		{"malformed JSON", `{"name":"John","email":}`},
		{"invalid syntax", `{name:"John"}`},
		{"incomplete", `{"name":"John"`},
		{"not JSON", `this is not JSON`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.jsonData)

			result, err := ReadJSON[testStruct](reader)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid JSON body")
			assert.Empty(t, result.Name)
		})
	}
}

func TestReadJSON_WithValidation_Success(t *testing.T) {
	jsonData := `{"name":"John","email":"john@example.com","age":30}`
	reader := strings.NewReader(jsonData)

	result, err := ReadJSON[*validatingStruct](reader)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadJSON_WithValidation_Failure(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectedErr string
	}{
		{
			name:        "missing name",
			jsonData:    `{"email":"john@example.com"}`,
			expectedErr: "name is required",
		},
		{
			name:        "missing email",
			jsonData:    `{"name":"John"}`,
			expectedErr: "email is required",
		},
		{
			name:        "both missing",
			jsonData:    `{}`,
			expectedErr: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.jsonData)

			_, err := ReadJSON[*validatingStruct](reader)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestReadJSON_WithValidation_AgeMutation(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectedAge int
	}{
		{
			name:        "negative age",
			jsonData:    `{"name":"John","email":"john@example.com","age":-5}`,
			expectedAge: 0,
		},
		{
			name:        "zero age",
			jsonData:    `{"name":"John","email":"john@example.com","age":0}`,
			expectedAge: 0,
		},
		{
			name:        "positive age",
			jsonData:    `{"name":"John","email":"john@example.com","age":25}`,
			expectedAge: 25,
		},
		{
			name:        "large negative age",
			jsonData:    `{"name":"John","email":"john@example.com","age":-100}`,
			expectedAge: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.jsonData)

			result, err := ReadJSON[*validatingStruct](reader)

			require.NoError(t, err)
			assert.Equal(t, "John", result.Name)
			assert.Equal(t, "john@example.com", result.Email)
			assert.Equal(t, tt.expectedAge, result.Age)
		})
	}
}

func TestReadXML_Success(t *testing.T) {
	xmlData := `<testStruct><name>John</name><email>john@example.com</email><age>30</age></testStruct>`
	reader := strings.NewReader(xmlData)

	result, err := ReadXML[testStruct](reader)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadXML_InvalidXML(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
	}{
		{"invalid syntax", `<testStruct><name>John</testStruct>`},
		{"not XML", `this is not XML`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xmlData)

			result, err := ReadXML[testStruct](reader)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid XML body")
			assert.Empty(t, result.Name)
		})
	}
}

func TestReadXML_EmptyBody(t *testing.T) {
	reader := strings.NewReader("")

	result, err := ReadXML[testStruct](reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid XML body")
	assert.Empty(t, result.Name)
}

func TestReadXML_WithValidation_Success(t *testing.T) {
	xmlData := `<validatingStruct><name>John</name><email>john@example.com</email><age>30</age></validatingStruct>`
	reader := strings.NewReader(xmlData)

	result, err := ReadXML[*validatingStruct](reader)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadXML_WithValidation_Failure(t *testing.T) {
	xmlData := `<validatingStruct><email>john@example.com</email></validatingStruct>`
	reader := strings.NewReader(xmlData)

	_, err := ReadXML[*validatingStruct](reader)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestReadXML_WithValidation_AgeMutation(t *testing.T) {
	tests := []struct {
		name        string
		xmlData     string
		expectedAge int
	}{
		{
			name:        "negative age",
			xmlData:     `<validatingStruct><name>John</name><email>john@example.com</email><age>-5</age></validatingStruct>`,
			expectedAge: 0,
		},
		{
			name:        "zero age",
			xmlData:     `<validatingStruct><name>John</name><email>john@example.com</email><age>0</age></validatingStruct>`,
			expectedAge: 0,
		},
		{
			name:        "positive age",
			xmlData:     `<validatingStruct><name>John</name><email>john@example.com</email><age>25</age></validatingStruct>`,
			expectedAge: 25,
		},
		{
			name:        "large negative age",
			xmlData:     `<validatingStruct><name>John</name><email>john@example.com</email><age>-100</age></validatingStruct>`,
			expectedAge: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.xmlData)

			result, err := ReadXML[*validatingStruct](reader)

			require.NoError(t, err)
			assert.Equal(t, "John", result.Name)
			assert.Equal(t, "john@example.com", result.Email)
			assert.Equal(t, tt.expectedAge, result.Age)
		})
	}
}

func TestReadBytes_Success(t *testing.T) {
	data := []byte("Hello, World!")
	reader := bytes.NewReader(data)

	result, err := ReadBytes(reader)

	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestReadBytes_EmptyBody(t *testing.T) {
	reader := bytes.NewReader([]byte{})

	result, err := ReadBytes(reader)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestVerifyValidator_WithValidator(t *testing.T) {
	valid := validatingStruct{Name: "John", Email: "john@example.com", Age: 30}
	err := verifyValidator(&valid)
	assert.NoError(t, err)
	assert.Equal(t, 30, valid.Age)

	invalid := validatingStruct{}
	err = verifyValidator(&invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	// Test age mutation
	negativeAge := validatingStruct{Name: "John", Email: "john@example.com", Age: -5}
	err = verifyValidator(&negativeAge)
	assert.NoError(t, err)
	assert.Equal(t, 0, negativeAge.Age, "negative age should be mutated to 0")
}

func TestVerifyValidator_WithoutValidator(t *testing.T) {
	nonValidator := testStruct{Name: "John"}
	err := verifyValidator(nonValidator)
	assert.NoError(t, err)
}

func TestReadJSON_FromRequest(t *testing.T) {
	jsonData := `{"name":"John","email":"john@example.com","age":30}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonData))

	result, err := ReadJSON[testStruct](req.Body)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadXML_FromRequest(t *testing.T) {
	xmlData := `<testStruct><name>John</name><email>john@example.com</email><age>30</age></testStruct>`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(xmlData))

	result, err := ReadXML[testStruct](req.Body)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadBytes_FromRequest(t *testing.T) {
	data := []byte("test data")
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(data))

	result, err := ReadBytes(req.Body)

	require.NoError(t, err)
	assert.Equal(t, data, result)
}

// errorReader is a helper type that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestReadJSON_ErrorDetails(t *testing.T) {
	jsonData := `invalid json`
	reader := strings.NewReader(jsonData)

	_, err := ReadJSON[testStruct](reader)

	require.Error(t, err)
	httpErr, ok := err.(*Error)
	require.True(t, ok)
	assert.NotNil(t, httpErr.Details)
}

func TestReadXML_ErrorDetails(t *testing.T) {
	xmlData := `invalid xml`
	reader := strings.NewReader(xmlData)

	_, err := ReadXML[testStruct](reader)

	require.Error(t, err)
	httpErr, ok := err.(*Error)
	require.True(t, ok)
	assert.NotNil(t, httpErr.Details)
}

func TestReadBytes_ErrorDetails(t *testing.T) {
	reader := &errorReader{}

	_, err := ReadBytes(reader)

	require.Error(t, err)
	httpErr, ok := err.(*Error)
	require.True(t, ok)
	assert.NotNil(t, httpErr.Details)
}

func TestReadJSON_PointerTypes(t *testing.T) {
	jsonData := `{"name":"John","email":"john@example.com","age":30}`
	reader := strings.NewReader(jsonData)

	result, err := ReadJSON[*testStruct](reader)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}

func TestReadJSON_WithExtraFields(t *testing.T) {
	jsonData := `{"name":"John","email":"john@example.com","age":30,"extra":"ignored"}`
	reader := strings.NewReader(jsonData)

	result, err := ReadJSON[testStruct](reader)

	require.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
	assert.Equal(t, 30, result.Age)
}
