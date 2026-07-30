package main

import "os"

// envAlias returns GRU_* when set, else legacy GORY_*.
func envAlias(gru, gory string) string {
	if v := os.Getenv(gru); v != "" {
		return v
	}
	return os.Getenv(gory)
}

func envAliasSet(gru, gory string) bool {
	return envAlias(gru, gory) != ""
}

func envAliasEq(gru, gory, want string) bool {
	return envAlias(gru, gory) == want
}
