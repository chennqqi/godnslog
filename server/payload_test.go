package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestIndex tests the Index handler returns expected HTML.
func TestIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	h := &WebServer{}
	h.Index(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "d4f800167a6e317f35454ed9024ebd420")
}

// TestStatus tests the Status handler returns empty 200.
func TestStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/status", nil)

	h := &WebServer{}
	h.Status(c)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Body.String())
}

// TestPhpRFI tests the phpRFI handler returns PHP code.
func TestPhpRFI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/php", nil)

	h := &WebServer{}
	h.phpRFI(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "md5")
}

// TestXss tests the xss handler returns XSS payload.
func TestXss(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/xss", nil)

	h := &WebServer{}
	h.xss(c)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "prompt")
}

// TestBoolPtr tests boolPtr helper.
func TestBoolPtr(t *testing.T) {
	b := boolPtr(true)
	assert.NotNil(t, b)
	assert.True(t, *b)

	b2 := boolPtr(false)
	assert.False(t, *b2)
}
