package handler

import (
	"fmt"
	"github.com/go-chassis/go-chassis/control"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/invocation"
	"github.com/go-chassis/go-chassis/core/qps"
	"net/http"
)

// ConsumerRateLimiterHandler consumer rate limiter handler
type ConsumerRateLimiterHandler struct{}

// Handle is handles the consumer rate limiter APIs
func (rl *ConsumerRateLimiterHandler) Handle(chain *Chain, i *invocation.Invocation, cb invocation.ResponseCallBack) {
	rlc := control.DefaultPanel.GetRateLimiting(*i, common.Consumer)
	if !rlc.Enabled {
		chain.Next(i, cb)

		return
	}
	//qps rate <=0
	if rlc.Rate <= 0 {
		r := newErrResponse(i, rlc)
		cb(r)
		return
	}
	//get operation meta info ms.schema, ms.schema.operation, ms
	if qps.GetRateLimiters().TryAccept(rlc.Key, rlc.Rate) {
		chain.Next(i, cb)
	} else {
		r := newErrResponse(i, rlc)
		cb(r)
		return
	}

}

func newErrResponse(i *invocation.Invocation, rlc control.RateLimitingConfig) *invocation.Response {
	switch i.Reply.(type) {
	case *http.Response:
		resp := i.Reply.(*http.Response)
		resp.StatusCode = http.StatusTooManyRequests
	}
	r := &invocation.Response{}
	r.Status = http.StatusTooManyRequests
	r.Err = fmt.Errorf("%s | %v", rlc.Key, rlc.Rate)
	return r
}

func newConsumerRateLimiterHandler() Handler {
	return &ConsumerRateLimiterHandler{}
}

// Name returns name
func (rl *ConsumerRateLimiterHandler) Name() string {
	return "consumerratelimiter"
}
