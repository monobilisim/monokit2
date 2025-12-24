package lib

import "os"

func IsTestMode() bool {
	v, ok := os.LookupEnv("TEST")
	return ok && (v == "1" || v == "true" || v == "yes")
}
