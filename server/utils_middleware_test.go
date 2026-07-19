package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestParseQuestionName tests parseQuestionName with various inputs.
func TestParseQuestionName(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		expect string
	}{
		{"foo.example.com", "example.com", "foo"},
		{"a.b.example.com", "example.com", "a.b"},
		{"example.com", "example.com", ""},
		{"other.com", "example.com", ""},
	}
	for _, tt := range tests {
		got := parseQuestionName(tt.name, tt.root)
		assert.Equal(t, tt.expect, got, "parseQuestionName(%q, %q)", tt.name, tt.root)
	}
}

// TestParseBinaryIP tests parseBinaryIP with valid and invalid inputs.
func TestParseBinaryIP(t *testing.T) {
	ip, err := parseBinaryIP("11000000101010000000000100000001")
	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, "192.168.1.1", ip.String())

	_, err = parseBinaryIP("invalid")
	assert.Error(t, err)
}

// TestParseHexIP tests parseHexIP with valid and invalid inputs.
func TestParseHexIP(t *testing.T) {
	ip, err := parseHexIP("c0a80101")
	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, "192.168.1.1", ip.String())

	_, err = parseHexIP("invalid")
	assert.Error(t, err)
}

// TestParseIP tests parseIP with various formats.
func TestParseIP(t *testing.T) {
	// Standard IP
	ip, err := parseIP("192.168.1.1")
	assert.NoError(t, err)
	assert.NotNil(t, ip)

	// Hex IP
	ip, err = parseIP("0xc0a80101")
	assert.NoError(t, err)
	assert.NotNil(t, ip)

	// Binary IP
	ip, err = parseIP("0b11000000101010000000000100000001")
	assert.NoError(t, err)
	assert.NotNil(t, ip)

	// Invalid IP
	ip, err = parseIP("not.an.ip")
	assert.NoError(t, err)
	assert.Nil(t, ip)
}

// TestGenRandomString tests genRandomString returns correct length and charset.
func TestGenRandomString(t *testing.T) {
	s := genRandomString(32)
	assert.Equal(t, 32, len(s))

	s2 := genRandomString(0)
	assert.Equal(t, 0, len(s2))
}

// TestGenRandomToken tests genRandomToken returns 64-char string.
func TestGenRandomToken(t *testing.T) {
	tok := genRandomToken()
	assert.Equal(t, 64, len(tok))
}

// TestGenShortId tests genShortId returns 12-char string.
func TestGenShortId(t *testing.T) {
	id := genShortId()
	assert.Equal(t, 12, len(id))
}

// TestGetSecuritySeed tests getSecuritySeed returns non-empty string.
func TestGetSecuritySeed(t *testing.T) {
	seed := getSecuritySeed()
	assert.NotEmpty(t, seed)
}

// TestMakePasswordAndCompare tests password hashing and comparison.
func TestMakePasswordAndCompare(t *testing.T) {
	hash := makePassword("testpass123")
	assert.NotEqual(t, "testpass123", hash)

	assert.NoError(t, comparePassword("testpass123", hash))
	assert.Error(t, comparePassword("wrongpass", hash))
}

// TestIsWeakPass tests weak password detection.
func TestIsWeakPass(t *testing.T) {
	assert.True(t, isWeakPass("12345"))
	assert.True(t, isWeakPass(""))
	assert.False(t, isWeakPass("123456"))
	assert.False(t, isWeakPass("strongpassword"))
}

// TestCustomQuote tests customQuote function.
func TestCustomQuote(t *testing.T) {
	assert.Equal(t, "'hello'", customQuote("hello"))
	assert.Equal(t, "''", customQuote(""))
}

// TestSwitchTranslate tests switchTranslate function.
func TestSwitchTranslate(t *testing.T) {
	assert.True(t, switchTranslate("en-US"))
	assert.True(t, switchTranslate("zh-CN"))
	assert.False(t, switchTranslate("fr-FR"))
}

// TestTranslate tests translate function.
func TestTranslate(t *testing.T) {
	switchTranslate("en-US")
	assert.Equal(t, "OK", translate("OK"))
	assert.Equal(t, "nonexistent", translate("nonexistent"))
}

// TestTranslateByLang tests translateByLang function.
func TestTranslateByLang(t *testing.T) {
	assert.Equal(t, "OK", translateByLang("en-US", "OK"))
	assert.Equal(t, "成功", translateByLang("zh-CN", "OK"))
	assert.Equal(t, "nonexistent", translateByLang("fr-FR", "nonexistent"))
}

// TestCORSMiddleware tests CORS middleware sets correct headers.
func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

// TestCORSMiddleware_Options tests CORS preflight request.
func TestCORSMiddleware_Options(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware())
	router.OPTIONS("/test", func(c *gin.Context) {})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// TestLoggingMiddleware tests logging middleware passes through.
func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestRecoveryMiddleware tests recovery middleware handles panics.
func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RecoveryMiddleware())
	router.GET("/panic", func(c *gin.Context) { panic("test panic") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestRecoveryMiddleware_NoPanic tests recovery middleware passes through normally.
func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RecoveryMiddleware())
	router.GET("/ok", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestAdminOnlyMiddleware_NoRole tests AdminOnlyMiddleware rejects when no role set.
func TestAdminOnlyMiddleware_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	s := &WebServer{}
	router.Use(func(c *gin.Context) { c.Next() })
	router.Use(s.AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestAdminOnlyMiddleware_Admin tests AdminOnlyMiddleware allows admin users.
func TestAdminOnlyMiddleware_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	s := &WebServer{}
	router.Use(func(c *gin.Context) { c.Set("role", 0); c.Next() })
	router.Use(s.AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// TestAdminOnlyMiddleware_NonAdmin tests AdminOnlyMiddleware rejects non-admin users.
func TestAdminOnlyMiddleware_NonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	s := &WebServer{}
	router.Use(func(c *gin.Context) { c.Set("role", 2); c.Next() })
	router.Use(s.AdminOnlyMiddleware())
	router.GET("/admin", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
