package ui

import "strings"

// fxOptOut is true when an env explicitly disables an fx kernel (testing / A-B).
// Unset env means fx stays on when emit-C is linked (native CGO build).
func fxOptOut(primary, legacy string) bool {
	v := strings.TrimSpace(EnvOr(primary, legacy))
	return v == "0" || strings.EqualFold(v, "false") || strings.EqualFold(v, "off")
}
