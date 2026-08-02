package server

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/cache"
)

func TestCaptchaService_Generate(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	challenge, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}
	if challenge.CaptchaID == "" {
		t.Error("Generate() returned empty captchaID")
	}
	if challenge.ImageBase64 == "" {
		t.Error("Generate() returned empty imageBase64")
	}
	if challenge.ThumbBase64 == "" {
		t.Error("Generate() returned empty thumbBase64")
	}
	if challenge.BlockWidth <= 0 || challenge.BlockHeight <= 0 {
		t.Errorf("Generate() returned invalid tile size: %dx%d", challenge.BlockWidth, challenge.BlockHeight)
	}
}

func TestCaptchaService_VerifyCorrect(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	challenge, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Read stored answer
	key := "captcha:" + challenge.CaptchaID
	v, exist := store.Get(key)
	if !exist {
		t.Fatal("captcha not found in cache after Generate()")
	}
	ans := v.(*captchaAnswer)

	// Verify with correct value (within tolerance)
	if !svc.Verify(challenge.CaptchaID, ans.X, ans.Y) {
		t.Errorf("Verify() returned false for correct value (x=%d, y=%d)", ans.X, ans.Y)
	}

	// Should be deleted after verify
	if _, stillExists := store.Get(key); stillExists {
		t.Error("captcha still in cache after successful Verify()")
	}
}

func TestCaptchaService_VerifyWrong(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	challenge, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	// Verify with wrong value
	if svc.Verify(challenge.CaptchaID, -999, 0) {
		t.Error("Verify() returned true for obviously wrong value")
	}
}

func TestCaptchaService_VerifyNonexistent(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	if svc.Verify("nonexistent-id", 100, 0) {
		t.Error("Verify() returned true for nonexistent captcha ID")
	}
}

func TestCaptchaService_VerifyReplay(t *testing.T) {
	store := cache.NewCache(time.Minute, 10*time.Second)
	svc := newCaptchaService(store, time.Minute)

	challenge, err := svc.Generate()
	if err != nil {
		t.Fatalf("Generate() returned error: %v", err)
	}

	key := "captcha:" + challenge.CaptchaID
	v, exist := store.Get(key)
	if !exist {
		t.Fatal("captcha not found in cache after Generate()")
	}
	ans := v.(*captchaAnswer)

	// First verify should succeed
	if !svc.Verify(challenge.CaptchaID, ans.X, ans.Y) {
		t.Fatal("first Verify() failed for correct value")
	}

	// Second verify (replay) must fail
	if svc.Verify(challenge.CaptchaID, ans.X, ans.Y) {
		t.Error("Verify() returned true for replayed captcha")
	}
}
