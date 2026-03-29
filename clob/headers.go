package clob

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

func createL1Headers(ctx context.Context, signer ClobSigner, chainID Chain, nonce *int64, timestamp *int64) (map[string]string, error) {
	effectiveTS := time.Now().Unix()
	if timestamp != nil {
		effectiveTS = *timestamp
	}
	n := int64(0)
	if nonce != nil {
		n = *nonce
	}
	sig, err := signer.SignClobAuth(ctx, chainID, effectiveTS, n)
	if err != nil {
		return nil, err
	}
	address, err := signer.Address(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"POLY_ADDRESS":   address,
		"POLY_SIGNATURE": sig,
		"POLY_TIMESTAMP": fmt.Sprintf("%d", effectiveTS),
		"POLY_NONCE":     fmt.Sprintf("%d", n),
	}, nil
}

func createL2Headers(ctx context.Context, signer ClobSigner, creds ApiKeyCreds, args L2HeaderArgs, timestamp *int64) (map[string]string, error) {
	ts := time.Now().Unix()
	if timestamp != nil {
		ts = *timestamp
	}
	address, err := signer.Address(ctx)
	if err != nil {
		return nil, err
	}
	sig, err := buildPolyHMACSignature(creds.Secret, ts, args.Method, args.RequestPath, args.Body)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"POLY_ADDRESS":    address,
		"POLY_SIGNATURE":  sig,
		"POLY_TIMESTAMP":  fmt.Sprintf("%d", ts),
		"POLY_API_KEY":    creds.Key,
		"POLY_PASSPHRASE": creds.Passphrase,
	}, nil
}

func buildPolyHMACSignature(secret string, timestamp int64, method string, requestPath string, body string) (string, error) {
	message := fmt.Sprintf("%d%s%s", timestamp, method, requestPath)
	if body != "" {
		message += body
	}
	s := strings.ReplaceAll(strings.ReplaceAll(secret, "-", "+"), "_", "/")
	for len(s)%4 != 0 {
		s += "="
	}
	key, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(message))
	sig := base64.StdEncoding.EncodeToString(h.Sum(nil))
	sig = strings.ReplaceAll(sig, "+", "-")
	sig = strings.ReplaceAll(sig, "/", "_")
	return sig, nil
}
