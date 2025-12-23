package linkgen

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/parkan/go-hauk/config"
	"github.com/parkan/go-hauk/store"
)

const (
	alphaLower     = "0123456789abcdefghijklmnopqrstuvwxyz"
	alphaMixed     = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	alphaUpper     = "123456789ABCDEFGHIJKLMNPQRSTUVWXYZ"
	prefixLocdata  = "locdata-"
)

type Generator struct {
	store store.Store
	style config.LinkStyle
}

func New(s store.Store, style config.LinkStyle) *Generator {
	return &Generator{store: s, style: style}
}

func (g *Generator) Generate(ctx context.Context) (string, error) {
	for {
		s, err := g.generate()
		if err != nil {
			return "", err
		}
		exists, err := g.store.Exists(ctx, prefixLocdata+s)
		if err != nil {
			return "", err
		}
		if !exists {
			return s, nil
		}
	}
}

func (g *Generator) generate() (string, error) {
	switch g.style {
	case config.LinkUUIDv4:
		return uuidV4()
	case config.Link16Hex:
		return hexString(8)
	case config.Link16Lower:
		return randomString(alphaLower, 16)
	case config.Link16Mixed:
		return randomString(alphaMixed, 16)
	case config.Link16Upper:
		return randomString(alphaUpper, 16)
	case config.Link32Hex:
		return hexString(16)
	case config.Link32Lower:
		return randomString(alphaLower, 32)
	case config.Link32Mixed:
		return randomString(alphaMixed, 32)
	case config.Link32Upper:
		return randomString(alphaUpper, 32)
	case config.Link4Plus4Lower:
		return fourPlusFour(alphaLower)
	case config.Link4Plus4Mixed:
		return fourPlusFour(alphaMixed)
	case config.Link4Plus4Upper:
		fallthrough
	default:
		return fourPlusFour(alphaUpper)
	}
}

func uuidV4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32]), nil
}

func hexString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomString(alpha string, length int) (string, error) {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := randInt(len(alpha))
		if err != nil {
			return "", err
		}
		b[i] = alpha[idx]
	}
	return string(b), nil
}

func fourPlusFour(alpha string) (string, error) {
	s, err := randomString(alpha, 8)
	if err != nil {
		return "", err
	}
	return s[:4] + "-" + s[4:], nil
}

func randInt(max int) (int, error) {
	b := make([]byte, 1)
	for {
		if _, err := rand.Read(b); err != nil {
			return 0, err
		}
		if int(b[0]) < 256-(256%max) {
			return int(b[0]) % max, nil
		}
	}
}
