//go:build grudemo

package examples

// PublicDemoMode is true when building with -tags grudemo (public curated demo app).
func PublicDemoMode() bool { return true }
