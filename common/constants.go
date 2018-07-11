package common

import "time"

const (
	K8S_VERSION                = "1.10.4"
	DEFAULT_APISERVER_PORT     = 6443
	DRAIN_TIMEOUT              = 5 * time.Minute
	DRAIN_GRACE_PERIOD_SECONDS = -1
)
