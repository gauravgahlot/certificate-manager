package controller

import "time"

const (
	reconcileInAMinute = 1 * 60 * time.Second
	reconcileShortly   = 5 * time.Second
	reconcileNone      = time.Duration(0)
	tlsKey             = "tls.key"
	tlsCert            = "tls.crt"
)

var isImmutable = true
