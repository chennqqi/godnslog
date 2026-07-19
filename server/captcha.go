package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/slide"

	"github.com/chennqqi/godnslog/cache"
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
	graph, shadow, mask := generateGraphShape()

	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: 300, Height: 220}),
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
	if err := png.Encode(masterBuf, data.GetMasterImage().Get()); err != nil {
		return "", "", "", fmt.Errorf("encode master image: %w", err)
	}
	masterBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(masterBuf.Bytes())

	// Encode thumb image (puzzle piece) to base64 PNG
	thumbBuf := new(bytes.Buffer)
	if err := png.Encode(thumbBuf, data.GetTileImage().Get()); err != nil {
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

// generateGraphShape creates a simple shape image for the puzzle piece and its shadow.
func generateGraphShape() (overlay, shadow, mask image.Image) {
	size := 60
	overlayImg := image.NewNRGBA(image.Rect(0, 0, size, size))
	shadowImg := image.NewNRGBA(image.Rect(0, 0, size, size))
	maskImg := image.NewNRGBA(image.Rect(0, 0, size, size))

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + 100))
	fg := color.NRGBA{
		R: uint8(rng.Intn(100) + 100),
		G: uint8(rng.Intn(100) + 100),
		B: uint8(rng.Intn(100) + 100),
		A: 255,
	}
	bg := color.NRGBA{A: 0} // transparent

	// Draw overlay (solid rounded rectangle)
	draw.Draw(overlayImg, overlayImg.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	drawRoundedRect(overlayImg, 5, 5, size-10, size-10, 10, fg)

	// Draw shadow (darker, semi-transparent, offset by 2px down-right)
	shadowColor := color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	draw.Draw(shadowImg, shadowImg.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	drawRoundedRect(shadowImg, 5+2, 5+2, size-10+2, size-10+2, 10, shadowColor)

	// Draw mask (white shape on transparent background — defines the cutout region)
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(maskImg, maskImg.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
	drawRoundedRect(maskImg, 5, 5, size-10, size-10, 10, white)

	return overlayImg, shadowImg, maskImg
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
