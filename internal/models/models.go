package models

import "time"

type Limit struct {
	Capacity   int64         `yaml:"capacity" env:"CAPACITY"`
	RefillRate time.Duration `yaml:"refill_rate_seconds" env:"REFILL_RATE_SECONDS"`
}

type ClientLimit struct {
	Key        string        `yaml:"key"`
	Capacity   int64         `yaml:"capacity"`
	RefillRate time.Duration `yaml:"refill_rate_seconds"`
	Unlimited  bool          `yaml:"unlimited"`
}
