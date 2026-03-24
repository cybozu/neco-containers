package authenticator

import "strings"

type Mode string

const (
	ModeX509 Mode = "x509"
	ModeJWT  Mode = "jwt"
)

func ParseMode(s string) Mode {
	switch strings.ToLower(s) {
	case string(ModeX509):
		return ModeX509
	case string(ModeJWT):
		return ModeJWT
	default:
		return ""
	}
}
