# Login Captcha Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add slide behavior captcha (wenlng/go-captcha v2) to the login flow for brute-force protection.

**Architecture:** Backend Go captcha service generates slide challenges, stores answers in cache. Frontend React component renders draggable puzzle piece. Captcha verified before password check.

**Tech Stack:** Go 1.25, Gin, wenlng/go-captcha v2, Next.js (React/TypeScript), Tailwind CSS

## Global Constraints

- Captcha must work fully offline — no external SaaS or CDN
- Background images generated at runtime (no bundled asset files)
- All captcha errors return generic message to client (no info leakage)
- Challenge TTL: 2 minutes, one-time use (deleted after verify success or failure)
- Tolerance for slide X position: ±3px
- System supports en-US / zh-CN — captcha UI labels use existing i18n context
- The frontend codebase is in `frontend-next/` (not the old `frontend/` directory)
- Existing cache.Cache (patrickmn/go-cache) used for captcha state storage
- Cluster: captcha state automatically shared when Redis is configured

---
### Task 1: Backend Captcha Service (`server/captcha.go`)

**Files:**
- Create: `server/captcha.go`
- Create: `server/captcha_test.go`

**Add dependency:**
```bash
go get github.com/wenlng/go-captcha/v2
```

**Interfaces:**
- Consumes: `cache.Cache` (existing), `time.Duration` (expire from config)
- Produces: `type captchaService struct`, `func newCaptchaService(store *cache.Cache, expire time.Duration) *captchaService`
  - `func (s *captchaService) Generate() (captchaID, imageBase64, thumbBase64 string, err error)`
  - `func (s *captchaService) Verify(captchaID string, value int) bool`

- [ ] **Step 1: Add go-captcha dependency**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go get github.com/wenlng/go-captcha/v2
```

- [ ] **Step 2: Create `server/captcha.go` with captcha service, resource generator, and slide builder**

```go
package server

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"bytes"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/wenlng/go-captcha/v2"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/chennqqi/godnslog/cache"
	"github.com/sirupsen/logrus"
)

const (
	defaultCaptchaExpire = 2 * time.Minute
	captchaTolerance     = 3 // pixels
)

type captchaAnswer struct {
	X int `json:"x"`
}

type captchaService struct {
	store  *cache.Cache
	capt   slide.Captcha
	expire time.Duration
}

func newCaptchaService(store *cache.Cache, expire time.Duration) *captchaService {
	if expire <= 0 {
		expire = defaultCaptchaExpire
	}

	bg := generateBackground()
	graph := generateGraphShape()

	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: 300, Height: 220}),
	)
	builder.SetResources(
		slide.WithBackgrounds([]image.Image{bg}),
		slide.WithGraphImages([]image.Image{graph}),
	)
	capt := builder.Make()

	return &captchaService{
		store:  store,
		capt:   capt,
		expire: expire,
	}
}

// Generate creates a new slide captcha challenge.
// Returns captchaID, base64-encoded main image, base64-encoded thumb image, and error.
func (s *captchaService) Generate() (captchaID, imageBase64, thumbBase64 string, err error) {
	data, err := s.capt.Generate()
	if err != nil {
		return "", "", "", fmt.Errorf("captcha generate: %w", err)
	}

	block := data.GetData()
	if block == nil {
		return "", "", "", fmt.Errorf("captcha data is nil")
	}

	// Encode master image (background with cutout) to base64 PNG
	masterBuf := new(bytes.Buffer)
	if _, err := png.Encode(masterBuf, data.GetMasterImage()); err != nil {
		return "", "", "", fmt.Errorf("encode master image: %w", err)
	}
	masterBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(masterBuf.Bytes())

	// Encode thumb image (puzzle piece) to base64 PNG
	thumbBuf := new(bytes.Buffer)
	if _, err := png.Encode(thumbBuf, data.GetThumbImage()); err != nil {
		return "", "", "", fmt.Errorf("encode thumb image: %w", err)
	}
	thumbBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(thumbBuf.Bytes())

	// Store answer
	captchaID = uuid.New().String()
	answer := &captchaAnswer{X: block.X}
	s.store.Set("captcha:"+captchaID, answer, s.expire)

	return captchaID, masterBase64, thumbBase64, nil
}

// Verify checks the user-provided captcha value against the stored answer.
// Always deletes the challenge from cache after verification (one-time use).
func (s *captchaService) Verify(captchaID string, value int) bool {
	key := "captcha:" + captchaID
	v, exist := s.store.Get(key)
	if !exist {
		return false
	}
	s.store.Delete(key)

	ans, ok := v.(*captchaAnswer)
	if !ok {
		return false
	}

	diff := value - ans.X
	if diff < 0 {
		diff = -diff
	}
	return diff <= captchaTolerance
}

// generateBackground creates a simple gradient background image at runtime.
func generateBackground() image.Image {
	width, height := 300, 220
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Seed with current time for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseR := uint8(rng.Intn(60) + 40)
	baseG := uint8(rng.Intn(60) + 40)
	baseB := uint8(rng.Intn(60) + 100)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Vertical gradient + slight noise
			noise := uint8(rng.Intn(30))
			r := uint8(float64(baseR)*(1-float64(y)/float64(height))) + noise
			g := uint8(float64(baseG)*(1-float64(y)/float64(height))) + noise
			b := uint8(float64(baseB)*(1-float64(y)/float64(height))) + noise
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Draw some random circles for visual complexity
	drawRandomCircles(img, rng)
	return img
}

// drawRandomCircles adds decorative circles to the background.
func drawRandomCircles(img *image.NRGBA, rng *rand.Rand) {
	bounds := img.Bounds()
	for i := 0; i < 8; i++ {
		cx := rng.Intn(bounds.Dx())
		cy := rng.Intn(bounds.Dy())
		radius := rng.Intn(30) + 10
		c := color.NRGBA{
			R: uint8(rng.Intn(80) + 40),
			G: uint8(rng.Intn(80) + 40),
			B: uint8(rng.Intn(80) + 40),
			A: 80,
		}
		drawCircle(img, cx, cy, radius, c)
	}
}

// drawCircle draws a filled circle on the image.
func drawCircle(img *image.NRGBA, cx, cy, radius int, c color.Color) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue
			}
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, c)
			}
		}
	}
}

// generateGraphShape creates a simple shape image for the puzzle piece.
func generateGraphShape() image.Image {
	size := 60
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + 100))
	fg := color.NRGBA{
		R: uint8(rng.Intn(100) + 100),
		G: uint8(rng.Intn(100) + 100),
		B: uint8(rng.Intn(100) + 100),
		A: 255,
	}
	bg := color.NRGBA{A: 0} // transparent

	// Draw a rounded rectangle shape
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	drawRoundedRect(img, 5, 5, size-10, size-10, 10, fg)

	return img
}

// drawRoundedRect draws a filled rounded rectangle.
func drawRoundedRect(img *image.NRGBA, x1, y1, x2, y2, r int, c color.Color) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			// Corner rounding logic
			inCorner := false
			if x < x1+r && y < y1+r {
				inCorner = (x-x1-r)*(x-x1-r)+(y-y1-r)*(y-y1-r) > r*r
			} else if x > x2-r && y < y1+r {
				inCorner = (x-x2+r)*(x-x2+r)+(y-y1-r)*(y-y1-r) > r*r
			} else if x < x1+r && y > y2-r {
				inCorner = (x-x1-r)*(x-x1-r)+(y-y2+r)*(y-y2+r) > r*r
			} else if x > x2-r && y > y2-r {
				inCorner = (x-x2+r)*(x-x2+r)+(y-y2+r)*(y-y2+r) > r*r
			}
			if !inCorner {
				img.Set(x, y, c)
			}
		}
	}
}
```

- [ ] **Step 3: Write unit tests in `server/captcha_test.go`**

```go
package server

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/cache"
)

func TestCaptchaService_Generate(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	captchaID, imageBase64, thumbBase64, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if captchaID == "" {
		t.Error("Generate() returned empty captchaID")
	}
	if imageBase64 == "" {
		t.Error("Generate() returned empty imageBase64")
	}
	if thumbBase64 == "" {
		t.Error("Generate() returned empty thumbBase64")
	}
}

func TestCaptchaService_VerifyCorrect(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	captchaID, _, _, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Read stored answer
	key := "captcha:" + captchaID
	v, exist := store.Get(key)
	if !exist {
		t.Fatal("captcha not found in cache after Generate()")
	}
	ans := v.(*captchaAnswer)

	// Verify with correct value (within tolerance)
	if !svc.Verify(captchaID, ans.X) {
		t.Errorf("Verify() returned false for correct value (x=%d)", ans.X)
	}

	// Should be deleted after verify
	if _, stillExists := store.Get(key); stillExists {
		t.Error("captcha still in cache after successful Verify()")
	}
}

func TestCaptchaService_VerifyWrong(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	captchaID, _, _, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Verify with wrong value
	if svc.Verify(captchaID, -999) {
		t.Error("Verify() returned true for obviously wrong value")
	}
}

func TestCaptchaService_VerifyNonexistent(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	if svc.Verify("nonexistent-id", 100) {
		t.Error("Verify() returned true for nonexistent captcha ID")
	}
}

func TestCaptchaService_VerifyReplay(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	captchaID, _, _, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	key := "captcha:" + captchaID
	v, exist := store.Get(key)
	if !exist {
		t.Fatal("captcha not found in cache after Generate()")
	}
	ans := v.(*captchaAnswer)

	// First verify should succeed
	if !svc.Verify(captchaID, ans.X) {
		t.Fatal("first Verify() failed for correct value")
	}

	// Second verify (replay) must fail
	if svc.Verify(captchaID, ans.X) {
		t.Error("Verify() returned true for replayed captcha")
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go test ./server/ -run TestCaptchaService -v
```
Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add server/captcha.go server/captcha_test.go go.mod go.sum
git commit -m "feat: add captcha service for slide verification"
```

---
### Task 2: Backend API Integration (endpoint + login + config)

**Files:**
- Modify: `models/api.go`
- Modify: `server/webui.go`
- Modify: `server/webserver.go`
- Modify: `servecmd.go`
- Modify: `main.go`

**Interfaces:**
- Consumes: `captchaService` from Task 1, `WebServerConfig.CaptchaEnabled`, `WebServerConfig.CaptchaExpire`
- Produces: `GET /api/auth/captcha` endpoint, modified `POST /api/auth/login` with captcha verification

- [ ] **Step 1: Add captcha fields to LoginRequest in `models/api.go:29`**

```go
type LoginRequest struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CaptchaID   string `json:"captcha_id"`    // new
	CaptchaValue int    `json:"captcha_value"` // new
}
```

- [ ] **Step 2: Add captcha config fields to WebServerConfig in `server/webserver.go:37-62`**

Inside `WebServerConfig struct`, add before the closing `}`:

```go
	CaptchaEnabled bool          `json:"captcha_enabled"` // default true
	CaptchaExpire  time.Duration // default 2min
```

- [ ] **Step 3: Initialize captcha service in NewWebServer in `server/webserver.go:90`**

After the existing GeoIP fingerprinter block (around line 128), add:

```go
	// Initialize captcha service
	if cfg.CaptchaEnabled {
		captchaExpire := cfg.CaptchaExpire
		if captchaExpire <= 0 {
			captchaExpire = 2 * time.Minute
		}
		app.captchaSvc = newCaptchaService(app.store, captchaExpire)
		logrus.Infof("[webserver.go::NewWebServer] captcha service initialized (expire=%v)", captchaExpire)
	} else {
		logrus.Info("[webserver.go::NewWebServer] captcha disabled")
	}
```

- [ ] **Step 4: Add captchaSvc field to WebServer struct in `server/webserver.go:64-88`**

Inside `WebServer struct`, after `fingerprinter` field:

```go
		captchaSvc  *captchaService
```

- [ ] **Step 5: Add captcha endpoint and modify login handler in `server/webui.go`**

First, add the captcha route alongside existing auth routes in the `Run()` method
(at `server/webserver.go:316`, in the `auth` group):

```go
	auth.POST("/login", self.userLogin)
	auth.POST("/logout", self.authHandler, self.userLogout)
	auth.GET("/info", self.authHandler, self.userInfo)
	auth.GET("/nav", self.authHandler, self.userNav)
	auth.GET("/captcha", self.getCaptcha) // new
```

Then, add the captcha endpoint handler in `server/webui.go` (before `userLogin`):

```go
func (self *WebServer) getCaptcha(c *gin.Context) {
	if self.captchaSvc == nil {
		self.resp(c, 404, &CR{
			Code:    CodeNoData,
			Message: "captcha disabled",
		})
		return
	}

	captchaID, imageBase64, thumbBase64, err := self.captchaSvc.Generate()
	if err != nil {
		logrus.Errorf("[webui.go::getCaptcha] Generate: %v", err)
		self.resp(c, 502, &CR{
			Code:    CodeServerInternal,
			Message: "Failed to generate captcha",
		})
		return
	}

	self.resp(c, 200, &CR{
		Code: CodeOK,
		Data: map[string]interface{}{
			"captcha_id":   captchaID,
			"image_base64": imageBase64,
			"thumb_base64": thumbBase64,
		},
	})
}
```

- [ ] **Step 6: Modify `userLogin` in `server/webui.go:321` to verify captcha before password**

Replace the current `userLogin` function body with one that checks captcha first. The new flow:

```go
func (self *WebServer) userLogin(c *gin.Context) {
	T := getTranslateFunc(c)

	var req LoginRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.Infof("[webui.go::userLogin] bad input param")
		self.resp(c, 400, &CR{
			Code:    CodeBadData,
			Message: T("bad input"),
		})
		return
	}

	// Captcha verification (if enabled)
	if self.captchaSvc != nil {
		if req.CaptchaID == "" {
			self.resp(c, 400, &CR{
				Code:    CodeBadData,
				Message: T("bad request"),
			})
			return
		}
		if !self.captchaSvc.Verify(req.CaptchaID, req.CaptchaValue) {
			self.resp(c, 400, &CR{
				Code:    CodeBadData,
				Message: T("bad request"),
			})
			return
		}
	}

	// Original login logic continues below
	session := self.orm.NewSession()
	defer session.Close()
	// ... rest of existing login handler unchanged from line 334 onward
}
```

Note: the existing login handler from `session := self.orm.NewSession()` through the JWT issuance stays exactly the same.

- [ ] **Step 7: Add CLI flags in `servecmd.go:45-63`**

Inside the `SetFlags` method, add captcha flags alongside existing ones:

```go
	f.BoolVar(&p.captchaEnabled, "captcha-enabled", true, "enable captcha verification on login, option")
	f.DurationVar(&p.captchaExpire, "captcha-expire", 2*time.Minute, "captcha challenge TTL, option")
```

- [ ] **Step 8: Add captcha fields to `servePwCmd` struct in `servecmd.go:20-35`**

```go
type servePwCmd struct {
	// ... existing fields
	captchaEnabled bool
	captchaExpire  time.Duration
}
```

- [ ] **Step 9: Pass captcha config to WebServerConfig in `servecmd.go:84-101`**

Add to the WebServerConfig literal, after `GeoIPLicenseKey`:

```go
	CaptchaEnabled:              p.captchaEnabled,
	CaptchaExpire:               p.captchaExpire,
```

- [ ] **Step 10: Add default constants in `main.go` (near existing defaults)**

```go
	DefaultCaptchaExpire = 2 * time.Minute
```

- [ ] **Step 11: Build and verify**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go build ./...
```
Expected: Build succeeds with no errors.

- [ ] **Step 12: Run existing tests to ensure no regression**

```bash
cd /data/dev/github.com/chennqqi/godnslog && go test ./server/ -v -count=1 2>&1 | head -80
```
Expected: All existing tests pass.

- [ ] **Step 13: Commit**

```bash
git add models/api.go server/webui.go server/webserver.go servecmd.go main.go
git commit -m "feat: integrate captcha into login API endpoint"
```

---
### Task 3: Frontend Captcha Infrastructure (Types + API + i18n)

**Files:**
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/lib/api-client.ts`
- Modify: `frontend-next/src/lib/i18n-context.tsx`

**Interfaces:**
- Consumes: `CaptchaResponse` type, `authApi.captcha()` method, captcha i18n keys
- Produces: Used by Task 4 (captcha-slide component) and Task 5 (login page)

- [ ] **Step 1: Add CaptchaResponse type and update LoginRequest in `frontend-next/src/types/index.ts`**

At `LoginRequest` interface (line 40), add captcha fields:

```typescript
export interface LoginRequest {
  username: string
  password: string
  captcha_id: string
  captcha_value: number
}
```

After `LoginResponse` interface (line 48), add new type:

```typescript
export interface CaptchaResponse {
  captcha_id: string
  image_base64: string
  thumb_base64: string
}
```

- [ ] **Step 2: Add captcha API method in `frontend-next/src/lib/api-client.ts`**

Add to the existing `authApi` object (around line 86):

```typescript
export const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  info: () => api.get('/auth/info'),
  captcha: () => api.get<CaptchaResponse>('/auth/captcha'),
}
```

- [ ] **Step 3: Add captcha i18n keys in `frontend-next/src/lib/i18n-context.tsx`**

In the `en-US` section (after line 28, around existing login keys), add:

```typescript
'login.captcha.slide_hint': 'Slide to verify',
'login.captcha.refresh': 'Refresh',
'login.captcha.error': 'Verification failed, please try again',
'login.captcha.loading': 'Loading captcha...',
'login.captcha.invalid': 'Please complete the captcha',
```

In the `zh-CN` section (after line 498, around existing login keys), add:

```typescript
'login.captcha.slide_hint': '拖动滑块完成验证',
'login.captcha.refresh': '刷新',
'login.captcha.error': '验证失败，请重试',
'login.captcha.loading': '加载验证码中...',
'login.captcha.invalid': '请完成验证码',
```

- [ ] **Step 4: Verify frontend builds**

```bash
cd /data/dev/github.com/chennqqi/godnslog/frontend-next && npx tsc --noEmit 2>&1 | head -30
```
Expected: TypeScript compiles cleanly (or shows only pre-existing errors unrelated to captcha).

- [ ] **Step 5: Commit**

```bash
git add frontend-next/src/types/index.ts frontend-next/src/lib/api-client.ts frontend-next/src/lib/i18n-context.tsx
git commit -m "feat: add frontend captcha types, API client, and i18n keys"
```

---
### Task 4: Frontend Slide Captcha Component

**Files:**
- Create: `frontend-next/src/features/auth/components/captcha-slide.tsx`

**Interfaces:**
- Consumes: `authApi.captcha()` from Task 3 (via `useI18n` for labels, `useEffect`/`useState` for lifecycle)
- Produces: `<SlideCaptcha>` component consumed by Task 5 login page

- [ ] **Step 1: Find the correct Tailwind color palette used in the login page**

Read current login page to check exact Tailwind class names used:

- [ ] **Step 2: Create `frontend-next/src/features/auth/components/captcha-slide.tsx`**

```tsx
'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { useI18n } from '@/lib/i18n-context'
import { authApi } from '@/lib/api-client'

export interface SlideCaptchaProps {
  onReady: (captchaId: string, captchaValue: number) => void
  onRefresh: () => void
  invalid: boolean
  disabled: boolean
}

export function SlideCaptcha({ onReady, onRefresh, invalid, disabled }: SlideCaptchaProps) {
  const { t } = useI18n()
  const [captchaId, setCaptchaId] = useState('')
  const [imageBase64, setImageBase64] = useState('')
  const [thumbBase64, setThumbBase64] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [dragging, setDragging] = useState(false)
  const [position, setPosition] = useState(0)
  const [completed, setCompleted] = useState(false)

  const containerRef = useRef<HTMLDivElement>(null)
  const thumbRef = useRef<HTMLDivElement>(null)
  const startXRef = useRef(0)
  const startPosRef = useRef(0)

  const loadCaptcha = useCallback(async () => {
    setLoading(true)
    setError('')
    setPosition(0)
    setCompleted(false)
    setCaptchaId('')

    try {
      const response = await authApi.captcha()
      if (response.code === 0 && response.data) {
        setCaptchaId(response.data.captcha_id)
        setImageBase64(response.data.image_base64)
        setThumbBase64(response.data.thumb_base64)
      } else {
        setError(response.message || t('login.captcha.error'))
      }
    } catch {
      setError(t('login.captcha.error'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadCaptcha()
  }, [loadCaptcha])

  useEffect(() => {
    if (invalid) {
      loadCaptcha()
    }
  }, [invalid, loadCaptcha])

  const handleRefresh = () => {
    loadCaptcha()
    onRefresh()
  }

  // Mouse drag handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    if (disabled || completed) return
    setDragging(true)
    startXRef.current = e.clientX
    startPosRef.current = position
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!dragging || disabled) return
    const container = containerRef.current
    if (!container) return

    const delta = e.clientX - startXRef.current
    const maxWidth = container.offsetWidth - 48 // thumb size
    const newPos = Math.max(0, Math.min(maxWidth, startPosRef.current + delta))
    setPosition(newPos)
  }

  const handleMouseUp = () => {
    if (!dragging) return
    setDragging(false)
    if (captchaId) {
      setCompleted(true)
      onReady(captchaId, position)
    }
  }

  // Touch drag handlers
  const handleTouchStart = (e: React.TouchEvent) => {
    if (disabled || completed) return
    setDragging(true)
    startXRef.current = e.touches[0].clientX
    startPosRef.current = position
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!dragging || disabled) return
    const container = containerRef.current
    if (!container) return

    const delta = e.touches[0].clientX - startXRef.current
    const maxWidth = container.offsetWidth - 48
    const newPos = Math.max(0, Math.min(maxWidth, startPosRef.current + delta))
    setPosition(newPos)
  }

  const handleTouchEnd = () => {
    if (!dragging) return
    setDragging(false)
    if (captchaId) {
      setCompleted(true)
      onReady(captchaId, position)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-24 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
        <span className="text-sm text-gray-400">{t('login.captcha.loading')}</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-center">
        <p className="text-sm text-red-600 dark:text-red-400 mb-2">{error}</p>
        <button
          type="button"
          onClick={handleRefresh}
          className="text-sm text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300 font-medium"
        >
          {t('login.captcha.refresh')}
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {/* Captcha image container */}
      <div
        ref={containerRef}
        className="relative w-full overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700 select-none"
        style={{ maxWidth: 300 }}
      >
        {/* Background image with cutout */}
        <img
          src={imageBase64}
          alt="captcha background"
          className="block w-full h-auto"
          draggable={false}
        />

        {/* Draggable puzzle piece overlay */}
        <div
          ref={thumbRef}
          className="absolute top-0 cursor-grab active:cursor-grabbing"
          style={{
            left: `${position}px`,
            top: 0,
            opacity: 0,
            pointerEvents: 'none',
          }}
        >
          <img
            src={thumbBase64}
            alt="puzzle piece"
            className="block"
            draggable={false}
          />
        </div>
      </div>

      {/* Slider track */}
      <div
        className="relative h-12 bg-gray-100 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
      >
        {/* Track background */}
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-sm text-gray-400 dark:text-gray-500 select-none">
            {completed ? '' : t('login.captcha.slide_hint')}
          </span>
        </div>

        {/* Slider thumb */}
        <div
          className="absolute top-0 left-0 h-full flex items-center justify-center bg-white dark:bg-gray-700 border-r border-gray-200 dark:border-gray-600 rounded-l-lg transition-shadow"
          style={{
            width: 48,
            transform: `translateX(${position}px)`,
          }}
        >
          {completed ? (
            <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          ) : (
            <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          )}
        </div>
      </div>

      {/* Refresh button */}
      <div className="flex justify-end">
        <button
          type="button"
          onClick={handleRefresh}
          disabled={disabled}
          className="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {t('login.captcha.refresh')}
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Verify frontend builds**

```bash
cd /data/dev/github.com/chennqqi/godnslog/frontend-next && npx tsc --noEmit 2>&1 | head -30
```
Expected: No errors related to the new component.

- [ ] **Step 4: Commit**

```bash
git add frontend-next/src/features/auth/components/captcha-slide.tsx
git commit -m "feat: add slide captcha React component"
```

---
### Task 5: Frontend Login Page Integration

**Files:**
- Modify: `frontend-next/src/app/login/page.tsx`

**Interfaces:**
- Consumes: `<SlideCaptcha>` component from Task 4, i18n keys from Task 3
- Produces: Login form with captcha challenge

- [ ] **Step 1: Modify login page to integrate captcha `frontend-next/src/app/login/page.tsx`**

Import the captcha component at the top (after existing imports):

```tsx
import { SlideCaptcha } from '@/features/auth/components/captcha-slide'
```

Add captcha state variables after the existing form state declarations (after line 51, around `const [error, setError] = useState('')`):

```tsx
const [captchaID, setCaptchaID] = useState('')
const [captchaValue, setCaptchaValue] = useState(0)
const [captchaInvalid, setCaptchaInvalid] = useState(false)
```

Modify the `onSubmit` handler to include captcha fields. Replace the existing handler (around line 70-96):

```tsx
const onSubmit = async (data: LoginFormValues) => {
  if (!captchaID) {
    setError(t('login.captcha.invalid'))
    return
  }

  setLoading(true)
  setError('')

  try {
    const response = await authApi.login({
      ...(data as LoginRequest),
      captcha_id: captchaID,
      captcha_value: captchaValue,
    })
    if (response.code === 0 && response.data) {
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      setToken(response.data.token)
      setUser({
        id: String(response.data.user.id),
        name: response.data.user.username,
        email: response.data.user.email,
        role: String(response.data.user.role),
      })
      router.push('/')
    } else {
      setCaptchaInvalid(true)
      setError(response.message || t('login.error'))
    }
  } catch (err: unknown) {
    const error = err as { response?: { data?: { message?: string } }, message?: string }
    setCaptchaInvalid(true)
    setError(error.response?.data?.message || error.message || t('login.error'))
  } finally {
    setLoading(false)
  }
}
```

Add the captcha component to the form, between the password field and submit button
(after the password input section, around line 241):

```tsx
            {/* Password */}
            ...
            </div>

            {/* Captcha */}
            <SlideCaptcha
              onReady={(id, value) => {
                setCaptchaID(id)
                setCaptchaValue(value)
              }}
              onRefresh={() => {
                setCaptchaID('')
                setCaptchaValue(0)
              }}
              invalid={captchaInvalid}
              disabled={loading}
            />

            {/* Submit */}
            <button
```

- [ ] **Step 2: Verify frontend builds**

```bash
cd /data/dev/github.com/chennqqi/godnslog/frontend-next && npx tsc --noEmit 2>&1 | head -30
```
Expected: No TypeScript errors.

- [ ] **Step 3: Commit**

```bash
git add frontend-next/src/app/login/page.tsx
git commit -m "feat: integrate slide captcha into login page"
```

---
## Self-Review

### Spec Coverage

| Spec Requirement | Task | Status |
|---|---|---|
| `server/captcha.go` — captcha service with Generate/Verify | Task 1 | Covered |
| `POST /api/auth/login` — captcha fields + verification before password | Task 2 | Covered |
| `GET /api/auth/captcha` — new endpoint | Task 2 | Covered |
| Cache management (captcha:{id}, 2min TTL, one-time use) | Task 1 | Covered |
| WebServerConfig — CaptchaEnabled, CaptchaExpire | Task 2 | Covered |
| CLI flags — -captcha-enabled, -captcha-expire | Task 2 | Covered |
| `frontend-next/src/features/auth/components/captcha-slide.tsx` | Task 4 | Covered |
| `frontend-next/src/app/login/page.tsx` — captcha integration | Task 5 | Covered |
| `frontend-next/src/lib/api-client.ts` — authApi.captcha() | Task 3 | Covered |
| `frontend-next/src/types/index.ts` — CaptchaResponse, LoginRequest | Task 3 | Covered |
| `frontend-next/src/lib/i18n-context.tsx` — 5 captcha translation keys | Task 3 | Covered |
| Captcha disabled mode (server returns 404, frontend hides component) | Task 2, 4 | Covered |
| Error handling — all captcha errors return generic message | Task 2 | Covered |
| Slide verification tolerance ±3px | Task 1 | Covered |

### Placeholder Scan
No placeholders found. All code blocks contain complete, compilable code.

### Type Consistency
- `captchaAnswer.X` (int) in Task 1 ← → `CaptchaValue int` in Task 2 LoginRequest
- `captchaService.Generate()` returns `(captchaID, imageBase64, thumbBase64 string, error)` in Task 1 ← → `CaptchaResponse { captcha_id, image_base64, thumb_base64 }` in Task 3
- `SlideCaptchaProps.onReady(captchaId, captchaValue)` in Task 4 ← → `setCaptchaID(id); setCaptchaValue(value)` in Task 5
- All field mappings are consistent end-to-end.

### No gaps found.
