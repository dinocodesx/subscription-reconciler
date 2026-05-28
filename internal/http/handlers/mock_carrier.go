package handlers

import (
	"math/rand/v2"
	"sync"
)

type MockCarrierHandler struct {
	random *rand.Rand
	mu     sync.Mutex
}
