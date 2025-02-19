package externalServices

import (
	"fmt"

	"net/http"
	"sync"
	"time"

	"bauman-poker/config"

	log "github.com/sirupsen/logrus"
)

type HealthTemplate func(url string) bool

type BreakerContext struct {
	urlH            string
	healthCheck     HealthTemplate
	mutex           *sync.Mutex
	isOpen          bool
	numOfFails      int64
	lastCall        time.Time
	lastSuccessCall time.Time
}

func NewBreakerContext(urlH string) *BreakerContext {
	mu := new(sync.Mutex)
	return &BreakerContext{
		isOpen:     true,
		numOfFails: 0,
		healthCheck: func(url string) bool {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.WithError(err).Error(fmt.Sprintf("healthCheck-error: in request\nurl: %s", url))
				return false
			}

			resp, err2 := http.DefaultClient.Do(req)
			if err2 != nil {
				log.WithError(err).Error(fmt.Sprintf("healthCheck-error: in getting response\nurl: %s", url))
				return false
			}

			log.Info(fmt.Sprintf("url: %s\nstatus = %d", url, resp.StatusCode))
			return resp.StatusCode == 200
		},
		urlH:  urlH,
		mutex: mu,
	}
}

func NewCircuitBreakerDecorator[T func(*http.Request) (*http.Response, error)](ctx *BreakerContext, sendReq T) T {
	log.Debug(fmt.Sprintf("Decorated url: %s", ctx.urlH))
	return func(req *http.Request) (*http.Response, error) {
		if !ctx.isOpen {
			//log.Info(fmt.Sprintf("endpoint %s is in close state", ctx.urlH))
			return nil, fmt.Errorf("503")
		}

		response, err := sendReq(req)
		if err != nil {
			ctx.mutex.Lock()
			ctx.lastCall = time.Now()
			ctx.numOfFails++

			log.WithError(err).Error(fmt.Sprintf("url: %s has %d fails", ctx.urlH, ctx.numOfFails))
			if ctx.numOfFails >= config.MaxNumOfFails {
				log.Info(fmt.Sprintf("url: %s is closed", ctx.urlH))
				ctx.isOpen = false
				retryTimeout(ctx, ctx.mutex)
			}
			ctx.mutex.Unlock()
			return response, fmt.Errorf("503")
		} else {
			log.Infof("Resp status code: %d", response.StatusCode)
			/*if response.StatusCode >= 400 {
				return response, fmt.Errorf("%d", response.StatusCode)
			}*/
			return response, err
		}
	}
}

func retryTimeout(ctx *BreakerContext, mutex *sync.Mutex) {
	time.AfterFunc(config.Timeout, func() {
		isAvailable := ctx.healthCheck(ctx.urlH)
		if isAvailable {
			log.Info(fmt.Sprintf("url: %s is opened", ctx.urlH))

			mutex.Lock()
			ctx.isOpen = true
			ctx.numOfFails = 0
			mutex.Unlock()
		} else {
			retryTimeout(ctx, mutex)
		}
	})
}
