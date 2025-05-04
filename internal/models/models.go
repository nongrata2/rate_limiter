package models

type Limit struct {
	Capacity   int64
	RefillRate int
}

type ClientLimit struct {
	Key        string
	Capacity   int64
	RefillRate int
	Unlimited  bool
}
