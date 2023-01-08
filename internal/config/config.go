package config

import (
	"flag"
	"time"
)

type Config struct {
	StoreFile     string
	StoreInterval time.Duration // 0 - disable save data
	Restore       bool
}

func NewConfig() *Config {
	f := flag.String("STORE_FILE", "db/stash.data", "store file")
	i := flag.Duration("STORE_INTERVAL", time.Second*5, "store interval")
	r := flag.Bool("RESTORE", true, "restore DB from disk on startup")
	flag.Parse()

	return &Config{
		StoreFile:     *f,
		StoreInterval: *i,
		Restore:       *r,
	}
}
