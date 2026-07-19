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
