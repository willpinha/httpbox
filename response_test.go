package httpbox

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponse struct {
	Message string `json:"message" xml:"message"`
	Code    int    `json:"code" xml:"code"`
}

func TestWriteJSON_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := testResponse{
		Message: "success",
		Code:    200,
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var result testResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, 200, result.Code)
}

func TestWriteJSON_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			data := map[string]string{"status": "test"}

			err := WriteJSON(rec, tt.statusCode, data)

			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}

func TestWriteJSON_DifferentDataTypes(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{"map", map[string]any{"key": "value", "number": 42}},
		{"slice", []string{"apple", "banana", "cherry"}},
		{"struct", testResponse{Message: "test", Code: 1}},
		{"primitive", 42},
		{"string", "hello"},
		{"bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			err := WriteJSON(rec, http.StatusOK, tt.data)

			require.NoError(t, err)
			assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
			assert.NotEmpty(t, rec.Body.String())
		})
	}
}

func TestWriteJSON_Nil(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteJSON(rec, http.StatusOK, nil)

	require.NoError(t, err)
	assert.Equal(t, "null\n", rec.Body.String())
}

func TestWriteJSON_NestedStructure(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	rec := httptest.NewRecorder()
	data := Person{
		Name: "John",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "John")
	assert.Contains(t, rec.Body.String(), "123 Main St")
	assert.Contains(t, rec.Body.String(), "NYC")
}

func TestWriteJSON_NonSerializable(t *testing.T) {
	rec := httptest.NewRecorder()
	// Channels cannot be serialized to JSON
	data := make(chan int)

	err := WriteJSON(rec, http.StatusOK, data)

	require.Error(t, err)
}

func TestWriteXML_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := testResponse{
		Message: "success",
		Code:    200,
	}

	err := WriteXML(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/xml")

	var result testResponse
	err = xml.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, 200, result.Code)
}

func TestWriteXML_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			data := testResponse{Message: "test", Code: 1}

			err := WriteXML(rec, tt.statusCode, data)

			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}

func TestWriteXML_NestedStructure(t *testing.T) {
	type Address struct {
		Street string `xml:"street"`
		City   string `xml:"city"`
	}
	type Person struct {
		XMLName xml.Name `xml:"person"`
		Name    string   `xml:"name"`
		Address Address  `xml:"address"`
	}

	rec := httptest.NewRecorder()
	data := Person{
		Name: "John",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	err := WriteXML(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "John")
	assert.Contains(t, rec.Body.String(), "123 Main St")
	assert.Contains(t, rec.Body.String(), "NYC")
}

func TestWriteBytes_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := []byte("Hello, World!")

	err := WriteBytes(rec, http.StatusOK, "text/plain", data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
	assert.Equal(t, "Hello, World!", rec.Body.String())
}

func TestWriteBytes_DifferentContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		data        []byte
	}{
		{"plain text", "text/plain", []byte("plain text")},
		{"html", "text/html", []byte("<html><body>Hello</body></html>")},
		{"binary", "application/octet-stream", []byte{0x00, 0x01, 0xFF}},
		{"csv", "text/csv", []byte("name,email\nJohn,john@example.com")},
		{"custom", "application/vnd.custom+json", []byte("{}")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			err := WriteBytes(rec, http.StatusOK, tt.contentType, tt.data)

			require.NoError(t, err)
			assert.Equal(t, tt.contentType, rec.Header().Get("Content-Type"))
			assert.Equal(t, tt.data, rec.Body.Bytes())
		})
	}
}

func TestWriteBytes_EmptyData(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteBytes(rec, http.StatusOK, "text/plain", []byte{})

	require.NoError(t, err)
	assert.Empty(t, rec.Body.String())
}

func TestWriteBytes_LargeData(t *testing.T) {
	rec := httptest.NewRecorder()
	data := bytes.Repeat([]byte("A"), 10000)

	err := WriteBytes(rec, http.StatusOK, "text/plain", data)

	require.NoError(t, err)
	assert.Equal(t, len(data), rec.Body.Len())
}

func TestWriteFromReader_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := "Hello from reader!"
	reader := strings.NewReader(data)

	err := WriteFromReader(rec, reader, http.StatusOK, "text/plain")

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
	assert.Equal(t, data, rec.Body.String())
}

func TestWriteFromReader_EmptyReader(t *testing.T) {
	rec := httptest.NewRecorder()
	reader := strings.NewReader("")

	err := WriteFromReader(rec, reader, http.StatusOK, "text/plain")

	require.NoError(t, err)
	assert.Empty(t, rec.Body.String())
}

func TestWriteFromReader_LargeReader(t *testing.T) {
	rec := httptest.NewRecorder()
	data := strings.Repeat("A", 100000)
	reader := strings.NewReader(data)

	err := WriteFromReader(rec, reader, http.StatusOK, "text/plain")

	require.NoError(t, err)
	assert.Equal(t, len(data), rec.Body.Len())
}

func TestWriteFromReader_BinaryData(t *testing.T) {
	rec := httptest.NewRecorder()
	data := []byte{0x00, 0x01, 0xFF, 0xFE, 0xAB, 0xCD}
	reader := bytes.NewReader(data)

	err := WriteFromReader(rec, reader, http.StatusOK, "application/octet-stream")

	require.NoError(t, err)
	assert.Equal(t, data, rec.Body.Bytes())
}

func TestWriteFromReader_ErrorReader(t *testing.T) {
	rec := httptest.NewRecorder()
	reader := &errorReader{}

	err := WriteFromReader(rec, reader, http.StatusOK, "text/plain")

	require.Error(t, err)
}

func TestWriteJSON_WithPointer(t *testing.T) {
	rec := httptest.NewRecorder()
	data := &testResponse{
		Message: "success",
		Code:    200,
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "success")
}

func TestWriteXML_WithPointer(t *testing.T) {
	rec := httptest.NewRecorder()
	data := &testResponse{
		Message: "success",
		Code:    200,
	}

	err := WriteXML(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "success")
}

func TestWriteJSON_ContentTypeSet(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteJSON(rec, http.StatusOK, map[string]string{"key": "value"})

	require.NoError(t, err)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestWriteXML_ContentTypeSet(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteXML(rec, http.StatusOK, testResponse{Message: "test", Code: 1})

	require.NoError(t, err)
	assert.Equal(t, "application/xml", rec.Header().Get("Content-Type"))
}

func TestWriteBytes_ContentTypeSet(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteBytes(rec, http.StatusOK, "image/png", []byte{})

	require.NoError(t, err)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
}

func TestWrite_StatusCodeSet(t *testing.T) {
	tests := []struct {
		name       string
		writeFunc  func(*httptest.ResponseRecorder) error
		statusCode int
	}{
		{
			name: "WriteJSON",
			writeFunc: func(rec *httptest.ResponseRecorder) error {
				return WriteJSON(rec, http.StatusCreated, map[string]string{})
			},
			statusCode: http.StatusCreated,
		},
		{
			name: "WriteXML",
			writeFunc: func(rec *httptest.ResponseRecorder) error {
				return WriteXML(rec, http.StatusAccepted, testResponse{})
			},
			statusCode: http.StatusAccepted,
		},
		{
			name: "WriteBytes",
			writeFunc: func(rec *httptest.ResponseRecorder) error {
				return WriteBytes(rec, http.StatusNoContent, "text/plain", []byte{})
			},
			statusCode: http.StatusNoContent,
		},
		{
			name: "WriteFromReader",
			writeFunc: func(rec *httptest.ResponseRecorder) error {
				return WriteFromReader(rec, strings.NewReader(""), http.StatusPartialContent, "text/plain")
			},
			statusCode: http.StatusPartialContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			err := tt.writeFunc(rec)

			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}

func TestWriteJSON_Array(t *testing.T) {
	rec := httptest.NewRecorder()
	data := []testResponse{
		{Message: "first", Code: 1},
		{Message: "second", Code: 2},
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "first")
	assert.Contains(t, rec.Body.String(), "second")

	var result []testResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestWriteBytes_Nil(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteBytes(rec, http.StatusOK, "text/plain", nil)

	require.NoError(t, err)
	assert.Empty(t, rec.Body.String())
}

func TestWriteJSON_RealWorldExample(t *testing.T) {
	type UserResponse struct {
		ID       int      `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
		Active   bool     `json:"active"`
	}

	rec := httptest.NewRecorder()
	data := UserResponse{
		ID:       123,
		Username: "johndoe",
		Email:    "john@example.com",
		Roles:    []string{"admin", "user"},
		Active:   true,
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)

	var result UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestWriteFromReader_MultipleChunks(t *testing.T) {
	rec := httptest.NewRecorder()
	// Create a reader with data larger than typical buffer size
	data := strings.Repeat("ABCDEFGHIJ", 10000)
	reader := strings.NewReader(data)

	err := WriteFromReader(rec, reader, http.StatusOK, "text/plain")

	require.NoError(t, err)
	assert.Equal(t, data, rec.Body.String())
}

func TestWriteJSON_SpecialCharacters(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{
		"message": "Hello \"World\" with special chars: <>&",
		"unicode": "Hello 世界 🌍",
	}

	err := WriteJSON(rec, http.StatusOK, data)

	require.NoError(t, err)

	var result map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestWriteXML_SpecialCharacters(t *testing.T) {
	type SpecialChars struct {
		XMLName xml.Name `xml:"data"`
		Message string   `xml:"message"`
	}

	rec := httptest.NewRecorder()
	data := SpecialChars{
		Message: "Hello <>&\" world",
	}

	err := WriteXML(rec, http.StatusOK, data)

	require.NoError(t, err)
	// XML encoder should properly escape special characters
	assert.Contains(t, rec.Body.String(), "&lt;")
	assert.Contains(t, rec.Body.String(), "&gt;")
	assert.Contains(t, rec.Body.String(), "&amp;")
}

func TestWriteFromReader_WithBytesReader(t *testing.T) {
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	data := []byte("test data")

	// Using WriteBytes
	err1 := WriteBytes(rec1, http.StatusOK, "text/plain", data)
	require.NoError(t, err1)

	// Using WriteFromReader directly
	err2 := WriteFromReader(rec2, bytes.NewReader(data), http.StatusOK, "text/plain")
	require.NoError(t, err2)

	// Both should produce the same result
	assert.Equal(t, rec1.Body.String(), rec2.Body.String())
	assert.Equal(t, rec1.Code, rec2.Code)
	assert.Equal(t, rec1.Header().Get("Content-Type"), rec2.Header().Get("Content-Type"))
}

func TestWrite_ResponseAlreadyWritten(t *testing.T) {
	rec := httptest.NewRecorder()

	// First write
	err := WriteJSON(rec, http.StatusOK, map[string]string{"first": "write"})
	require.NoError(t, err)

	// Second write - should still work but append to body
	err = WriteJSON(rec, http.StatusCreated, map[string]string{"second": "write"})
	require.NoError(t, err)

	// Body should contain both writes
	body := rec.Body.String()
	assert.Contains(t, body, "first")
	assert.Contains(t, body, "second")
}
