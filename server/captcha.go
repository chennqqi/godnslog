package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/chennqqi/godnslog/cache"
)

const (
	defaultCaptchaExpire = 2 * time.Minute
	captchaTolerance     = 5
	captchaImageWidth    = 300
	captchaImageHeight   = 160
)

type captchaAnswer struct {
	X int
	Y int
}

// CaptchaChallenge carries everything the frontend needs to render the slide
// captcha at 1:1 scale: the images plus the puzzle tile's display position and
// size in image coordinates. The answer (hole position X/Y) is never exposed.
type CaptchaChallenge struct {
	CaptchaID   string
	ImageBase64 string
	ThumbBase64 string
	BlockDX     int
	BlockDY     int
	BlockWidth  int
	BlockHeight int
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
	graph, shadow, mask := generateGraphShape()

	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: captchaImageWidth, Height: captchaImageHeight}),
	)
	builder.SetResources(
		slide.WithBackgrounds([]image.Image{bg}),
		slide.WithGraphImages([]*slide.GraphImage{
			{
				OverlayImage: graph,
				ShadowImage:  shadow,
				MaskImage:    mask,
			},
		}),
	)
	capt := builder.Make()

	return &captchaService{
		store:  store,
		capt:   capt,
		expire: expire,
	}
}

// Generate creates a new slide captcha challenge.
func (s *captchaService) Generate() (*CaptchaChallenge, error) {
	data, err := s.capt.Generate()
	if err != nil {
		return nil, fmt.Errorf("captcha generate: %w", err)
	}

	block := data.GetData()
	if block == nil {
		return nil, fmt.Errorf("captcha data is nil")
	}

	masterBuf := new(bytes.Buffer)
	if err := png.Encode(masterBuf, data.GetMasterImage().Get()); err != nil {
		return nil, fmt.Errorf("encode master image: %w", err)
	}
	masterBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(masterBuf.Bytes())

	thumbBuf := new(bytes.Buffer)
	if err := png.Encode(thumbBuf, data.GetTileImage().Get()); err != nil {
		return nil, fmt.Errorf("encode thumb image: %w", err)
	}
	thumbBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(thumbBuf.Bytes())

	challenge := &CaptchaChallenge{
		CaptchaID:   uuid.New().String(),
		ImageBase64: masterBase64,
		ThumbBase64: thumbBase64,
		BlockDX:     block.DX,
		BlockDY:     block.DY,
		BlockWidth:  block.Width,
		BlockHeight: block.Height,
	}
	answer := &captchaAnswer{X: block.X, Y: block.Y}
	s.store.Set("captcha:"+challenge.CaptchaID, answer, s.expire)

	return challenge, nil
}

// Verify checks the user-provided captcha position using go-captcha's Validate.
// Always deletes the challenge from cache after verification (one-time use).
func (s *captchaService) Verify(captchaID string, x, y int) bool {
	key := "captcha:" + captchaID
	v, exist := s.store.Get(key)
	if !exist {
		return false
	}

	ans, ok := v.(*captchaAnswer)
	if !ok {
		return false
	}
	s.store.Delete(key)

	return slide.Validate(ans.X, ans.Y, x, y, captchaTolerance)
}

// --- image generation helpers ---

func generateBackground() image.Image {
	width, height := captchaImageWidth, captchaImageHeight
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Soft pastel palette with subtle color shifts
	baseH := 30.0 + rng.Float64()*60 // warm hue range
	baseL := 0.75 + rng.Float64()*0.15
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			h := baseH + float64(x)*0.08 + float64(y)*0.02 + rng.Float64()*3
			s := 0.12 + rng.Float64()*0.08
			l := baseL + float64(y)*0.0003 + rng.Float64()*0.03
			r, g, b := hslToRGB(h, s, l)
			img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Organic curved shapes (like watercolor strokes)
	for i := 0; i < 3; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		cr := rng.Intn(60) + 30
		rc := color.NRGBA{
			R: uint8(rng.Intn(60) + 180),
			G: uint8(rng.Intn(60) + 180),
			B: uint8(rng.Intn(60) + 180),
			A: uint8(rng.Intn(30) + 15),
		}
		for y := cy - cr; y <= cy+cr; y++ {
			for x := cx - cr; x <= cx+cr; x++ {
				if x < 0 || x >= width || y < 0 || y >= height {
					continue
				}
				dx, dy := float64(x-cx), float64(y-cy)
				d := dx*dx/(float64(cr)*float64(cr)) + dy*dy/(float64(cr*cr/2))
				if d < 1.0 {
					alpha := uint8(float64(rc.A) * (1.0 - d))
					existing := img.NRGBAAt(x, y)
					r := uint8((int(existing.R)*int(255-alpha) + int(rc.R)*int(alpha)) / 255)
					g := uint8((int(existing.G)*int(255-alpha) + int(rc.G)*int(alpha)) / 255)
					b := uint8((int(existing.B)*int(255-alpha) + int(rc.B)*int(alpha)) / 255)
					img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}

	// Subtle wavy lines
	for i := 0; i < 2; i++ {
		startY := rng.Intn(height)
		lc := color.NRGBA{
			R: uint8(rng.Intn(40) + 200),
			G: uint8(rng.Intn(40) + 200),
			B: uint8(rng.Intn(40) + 200),
			A: uint8(rng.Intn(30) + 15),
		}
		amplitude := float64(rng.Intn(15) + 5)
		freq := rng.Float64()*0.03 + 0.02
		phase := rng.Float64() * 10
		for x := 0; x < width; x++ {
			y := startY + int(amplitude*math.Sin(float64(x)*freq+phase))
			if y >= 0 && y < height {
				for dy := -2; dy <= 2; dy++ {
					if y+dy >= 0 && y+dy < height {
						existing := img.NRGBAAt(x, y+dy)
						r := uint8((int(existing.R)*int(255-lc.A) + int(lc.R)*int(lc.A)) / 255)
						g := uint8((int(existing.G)*int(255-lc.A) + int(lc.G)*int(lc.A)) / 255)
						b := uint8((int(existing.B)*int(255-lc.A) + int(lc.B)*int(lc.A)) / 255)
						img.Set(x, y+dy, color.NRGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}

	return img
}

func generateGraphShape() (overlay, shadow, mask image.Image) {
	size := 60
	overlayImg := image.NewNRGBA(image.Rect(0, 0, size, size))
	shadowImg := image.NewNRGBA(image.Rect(0, 0, size, size))
	maskImg := image.NewNRGBA(image.Rect(0, 0, size, size))
	transparent := color.NRGBA{A: 0}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + 100))
	r := uint8(rng.Intn(40) + 200)
	g := uint8(rng.Intn(40) + 200)
	b := uint8(rng.Intn(40) + 200)
	fg := color.NRGBA{R: r, G: g, B: b, A: 230}

	pad := 3
	tabH := 8
	total := size - pad*2

	draw.Draw(overlayImg, overlayImg.Bounds(), &image.Uniform{transparent}, image.Point{}, draw.Src)
	drawPuzzleShape(overlayImg, pad, pad+tabH, total, total-tabH, 5, fg, tabH, true)

	// The shadow must be drawn at the SAME position as the overlay so that when
	// the user aligns the puzzle piece with the hole, their left edges coincide.
	sh := color.NRGBA{R: 0, G: 0, B: 0, A: 70}
	draw.Draw(shadowImg, shadowImg.Bounds(), &image.Uniform{transparent}, image.Point{}, draw.Src)
	drawPuzzleShape(shadowImg, pad, pad+tabH, total, total-tabH, 5, sh, tabH, true)

	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(maskImg, maskImg.Bounds(), &image.Uniform{transparent}, image.Point{}, draw.Src)
	drawPuzzleShape(maskImg, pad, pad+tabH, total, total-tabH, 5, white, tabH, true)

	return overlayImg, shadowImg, maskImg
}

func drawPuzzleShape(img *image.NRGBA, x1, y1, w, h, r int, c color.Color, tabH int, top bool) {
	x2 := x1 + w
	y2 := y1 + h
	drawRoundedRect(img, x1, y1, x2, y2, r, c)
	if top {
		cx := (x1 + x2) / 2
		tw := max(8, w/3)
		for y := y1 - tabH; y < y1; y++ {
			for x := cx - tw/2; x < cx+tw/2; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, c)
				}
			}
		}
		// Round the tab top
		tabTop := y1 - tabH
		for x := cx - tw/2 + 2; x < cx+tw/2-1; x++ {
			if tabTop >= 0 && tabTop < img.Bounds().Dy() && x >= 0 && x < img.Bounds().Dx() {
				img.Set(x, tabTop, c)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(h, 360) / 360
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := l - c/2
	var rp, gp, bp float64
	switch int(h * 6) {
	case 0:
		rp, gp, bp = c, x, 0
	case 1:
		rp, gp, bp = x, c, 0
	case 2:
		rp, gp, bp = 0, c, x
	case 3:
		rp, gp, bp = 0, x, c
	case 4:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}
	return uint8((rp + m) * 255), uint8((gp + m) * 255), uint8((bp + m) * 255)
}

func drawRoundedRect(img *image.NRGBA, x1, y1, x2, y2, r int, c color.Color) {
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
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

func drawCircle(img *image.NRGBA, cx, cy, radius int, c color.Color) {}

func drawRandomCircles(img *image.NRGBA, rng *rand.Rand) {}
