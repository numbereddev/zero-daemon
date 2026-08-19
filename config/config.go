package config

import "sync/atomic"

// TODO: configuration file

type Config struct{}

var config atomic.Pointer[Config]
