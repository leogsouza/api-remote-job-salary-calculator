package main

import (
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*visitor
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*visitor),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
	return i
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = &visitor{limiter, time.Now()}

	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	v, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}
	i.mu.Unlock()
	return v.limiter
}

func (i *IPRateLimiter) CleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		i.mu.Lock()
		defer i.mu.Unlock()
		for ip, v := range i.ips {
			if time.Now().Sub(v.lastSeen) > 30*time.Second {
				log.Printf("Cleaning ip %s", ip)
				delete(i.ips, ip)
			}
		}
	}
}