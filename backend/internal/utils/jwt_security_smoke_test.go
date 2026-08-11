package utils

import (
	"testing"
	"time"

	"schedule-system-v2/backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTRoundtrip 正常 HS256 token 签发/解析应成功
func TestJWTRoundtrip(t *testing.T) {
	token, err := GenerateToken(1, "20230001", "admin", "办公室", "dept_admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 1 || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

// TestRejectOtherAlgorithms 算法混淆防护：HS384/HS512/none 签发的 token 必须被拒绝
func TestRejectOtherAlgorithms(t *testing.T) {
	cfg := config.GetConfig()
	claims := Claims{
		UserID:    1,
		StudentID: "20230001",
		Role:      "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// HS512 混淆
	hs512 := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	s512, err := hs512.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		t.Fatalf("sign hs512 failed: %v", err)
	}
	if _, err := ParseToken(s512); err == nil {
		t.Fatal("HS512 token should be rejected but was accepted")
	}

	// none 算法
	noneTok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	sNone, err := noneTok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none failed: %v", err)
	}
	if _, err := ParseToken(sNone); err == nil {
		t.Fatal("none-alg token should be rejected but was accepted")
	}

	// 错误密钥
	bad := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	sBad, _ := bad.SignedString([]byte("wrong-secret"))
	if _, err := ParseToken(sBad); err == nil {
		t.Fatal("wrong-secret token should be rejected but was accepted")
	}
}
