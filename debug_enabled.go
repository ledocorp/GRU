package main

func debugVerbose() bool {
	v := envAlias("GRU_DEBUG", "GORY_DEBUG")
	if appReleaseMode() {
		return v == "1"
	}
	return v != "0"
}
