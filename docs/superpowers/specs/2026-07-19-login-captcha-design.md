# Login Captcha: Slide Behavior Verification

## Overview

为 GODNSLOG 登录流程添加滑块行为验证码，防止暴力破解和自动化登录攻击。

## Background

- 项目使用 Go + Gin 后端和 Next.js (React/TypeScript) 前端
- 现有登录为 `POST /api/auth/login`，仅有用户名密码验证
- 需要完全离线部署，无第三方 SaaS 依赖
- 系统支持 `en-US` / `zh-CN` 双语切换

## Chosen Solution

**wenlng/go-captcha v2** — 滑块模式 (Slide)

选型理由：
- 纯 Go，无 CGO，完全离线可用
- 行为验证码强度"强"，有效防御自动化攻击
- 滑块操作与语言无关，中英文环境一致体验
- 社区成熟 (5k+ stars)，文档完善
- challenge 可通过现有 cache/Redis 共享，支持集群部署

## Architecture

### Data Flow

```
Login Page Load
  → GET /api/auth/captcha
  ← { captcha_id, image_base64, thumb_base64 }

User drags slider to align puzzle piece
User clicks Login
  → POST /api/auth/login { username, password, captcha_id, captcha_value }

Server:
  1. Retrieve expected answer from cache by captcha_id
  2. Compare |captcha_value - expected_x| ≤ tolerance (3px)
  3. Delete captcha_id from cache (one-time use)
  4. If valid → verify password → issue JWT
  5. If invalid → return error, frontend refreshes captcha
```

**Validation order:** captcha → password. Captcha is verified before password
check to prevent timing-based information leakage.

### Backend

#### New file: `server/captcha.go`

- Initialize `wenlng/go-captcha/v2` slide builder
- `GenerateCaptcha()` → creates challenge, stores answer in cache, returns
  `(captchaID, imageBase64, thumbBase64, error)`
- `VerifyCaptcha(captchaID string, value int) bool` → validates answer with
  tolerance, always deletes from cache afterwards
- Background images generated at runtime (gradient/noise patterns) — no bundled
  asset files needed

#### Modified API: `POST /api/auth/login` (`server/webui.go:321`)

`LoginRequest` gains two fields:

```go
type LoginRequest struct {
    Email        string `json:"email"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    CaptchaID   string `json:"captcha_id"`    // new
    CaptchaValue int    `json:"captcha_value"` // new: slide X offset
}
```

Login handler adds captcha verification step before password check.

#### New endpoint: `GET /api/auth/captcha` (`server/webui.go`)

Returns a fresh captcha challenge. No auth required.

```json
{
  "code": 0,
  "data": {
    "captcha_id": "uuid-v4",
    "image_base64": "data:image/png;base64,...",
    "thumb_base64": "data:image/png;base64,..."
  }
}
```

Error response (server error):

```json
{
  "code": 5,
  "message": "Failed to generate captcha"
}
```

#### Cache management

- **Key:** `captcha:{captcha_id}`, **Value:** `{ X: int, Y: int }`
- **TTL:** 2 minutes (configured via `CaptchaExpire` in WebServerConfig)
- **Deletion:** on verify success or failure (one-time use)
- **Cluster:** automatically shared when Redis cache is configured

#### Configuration (`server/webserver.go`)

```go
type WebServerConfig struct {
    // ... existing fields
    CaptchaEnabled bool   `json:"captcha_enabled"` // default true
    CaptchaExpire  time.Duration                   // default 2min
}
```

Command-line flags: `-captcha-enabled=true`, `-captcha-expire=2m`

### Frontend

#### New file: `frontend-next/src/features/auth/components/captcha-slide.tsx`

A React component that:

1. On mount → calls `GET /api/auth/captcha` to load challenge
2. Renders the background image (with cutout) in a container
3. Renders the puzzle piece as a draggable overlay
4. **Mouse drag:** tracks `mousedown` / `mousemove` / `mouseup` to slide piece
5. **Touch drag:** tracks `touchstart` / `touchmove` / `touchend` for mobile
6. On drag end → records the X offset as `captcha_value`
7. Shows a refresh button to re-fetch challenge
8. Displays error state if captcha loading fails
9. Exposes `captchaValue` and `captchaID` to parent form via callback props

Component props:

```typescript
interface SlideCaptchaProps {
  onReady: (captchaId: string, captchaValue: number) => void
  onRefresh: () => void
  invalid: boolean      // true = last verify failed, reset captcha
  disabled: boolean     // true = form submitting, lock interaction
}
```

#### Modified: `frontend-next/src/app/login/page.tsx`

- Add `captchaID` and `captchaValue` to login form state
- Render `<SlideCaptcha>` between password field and submit button
- Pass `captcha_id` and `captcha_value` in login request payload
- On captcha error → reset captcha component, show error message

#### Modified: `frontend-next/src/lib/api-client.ts`

Add captcha API:

```typescript
export const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  logout: () => api.post('/auth/logout'),
  info: () => api.get('/auth/info'),
  captcha: () => api.get<CaptchaResponse>('/auth/captcha'),  // new
}
```

#### Modified: `frontend-next/src/types/index.ts`

```typescript
export interface LoginRequest {
  username: string
  password: string
  captcha_id: string    // new
  captcha_value: number // new
}

export interface CaptchaResponse {           // new
  captcha_id: string
  image_base64: string
  thumb_base64: string
}
```

#### i18n: `frontend-next/src/lib/i18n-context.tsx`

Add captcha-related translation keys:

| Key | en-US | zh-CN |
|-----|-------|-------|
| `login.captcha.slide_hint` | "Slide to verify" | "拖动滑块完成验证" |
| `login.captcha.refresh` | "Refresh" | "刷新" |
| `login.captcha.error` | "Verification failed, please try again" | "验证失败，请重试" |
| `login.captcha.loading` | "Loading captcha..." | "加载验证码中..." |
| `login.captcha.invalid` | "Please complete the captcha" | "请完成验证码" |

### Error Handling

| Scenario | HTTP Status | Code | Frontend Behavior |
|----------|-------------|------|-------------------|
| Captcha expired | 400 | 2 (BadData) | Refresh captcha, show generic error |
| Captcha value wrong | 400 | 2 (BadData) | Refresh captcha, show generic error |
| Captcha ID missing | 400 | 2 (BadData) | Show validation error |
| Captcha replayed | 400 | 2 (BadData) | Refresh captcha, show generic error |
| Captcha generation failed | 502 | 5 (ServerInternal) | Show generic error, retry button |
| Backend disabled captcha | 404 | 6 (NoData) JSON response | Hide captcha component |

All captcha errors return the same generic message to the client, avoiding
information leakage about the verification internals.

### Configurability

Captcha can be toggled on/off via server flag `-captcha-enabled`. When disabled:
- `GET /api/auth/captcha` returns 404
- Login endpoint skips captcha verification
- Frontend hides the captcha component

## Security Considerations

- Challenge generated with cryptographically secure random source
- One-time use: deleted from cache after verify success or failure
- 2-minute default TTL prevents replay window
- Answer is X coordinate only, verified with ±3px tolerance
- No answer data returned to client (client only sees images)
- Failed verification resets challenge (new image, new answer)
- Error messages do not distinguish between "wrong captcha" and "expired captcha"
- Same rate limiting applies whether captcha is enabled or not

## Out of Scope

- Audio/accessibility alternatives (can be added later with click mode)
- Click/text mode captcha (slide is language-independent, preferred)
- Challenge storage in database (cache TTL is sufficient)
- Captcha admin UI (config is CLI flag only for now)

## Testing

- Backend: unit tests for captcha generate/verify logic in `server/captcha_test.go`
- Frontend: component render test, drag interaction simulation
- Integration: login flow with correct/incorrect/expired captcha values
