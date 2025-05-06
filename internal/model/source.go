package model

import (
	"net/url"
)

type Source struct {
	URL       *url.URL
	RateLimit *rateLimit
}

type rateLimit struct {
	RespectRobotsTxt
}
