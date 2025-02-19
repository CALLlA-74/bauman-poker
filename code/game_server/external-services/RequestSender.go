package externalServices

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"bauman-poker/config"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type RequestSender struct {
	sendReq     func(req *http.Request) (*http.Response, error)
	mutex       *sync.Mutex
	context     *BreakerContext
	queue       *list.List
	isExecuting bool
}

func NewRequestSender(ctx *BreakerContext) *RequestSender {
	log.Info("RequestSender init. health-check url: ", ctx.urlH)
	return &RequestSender{
		sendReq: NewCircuitBreakerDecorator(ctx, http.DefaultClient.Do),
		context: ctx,
		mutex:   new(sync.Mutex),
		queue:   list.New(),
	}
}

func (sender RequestSender) SendRequest(req *http.Request) (*http.Response, error) {
	resp, err2 := sender.sendReq(req)
	if err2 != nil {
		if err2.Error() == "503" {
			log.Errorf("error in sending request. url: %s: %s\n", req.Method, req.URL.String())
		} else {
			log.Errorf("error response was got. url: %s: %s; status code = %d; status: %s\n", req.Method, req.URL.String(), resp.StatusCode, resp.Status)
		}
	} else {
		log.Infof("response received successful. url: %s: %s; status code = %d; status: %s\n", req.Method, req.URL.String(), resp.StatusCode, resp.Status)
	}
	return resp, err2
}

func Unpack(respBody []byte, respStructPtr any) error {
	if err4 := json.Unmarshal(respBody, respStructPtr); err4 != nil {
		log.WithError(err4).Errorf("error in unmarshalling response. ResponseBody: %#v\n", respBody)
		return fmt.Errorf("500")
	}

	if err := validator.New().Struct(respStructPtr); err != nil {
		log.WithError(err).Errorf("error in validate required fields in unpacking")
		return fmt.Errorf("500")
	}

	log.Infof("return response for request url")
	return nil
}

func (sender RequestSender) ReadAll(resp *http.Response) ([]byte, error) {
	respBody, err3 := io.ReadAll(resp.Body)
	if err3 != nil {
		log.WithError(err3).Errorf("error in reading response body. url: %s\n", resp.Request.URL.String())
		return nil, fmt.Errorf("500")
	}
	return respBody, nil
}

func (sender RequestSender) SendRequestForever(req *http.Request) error {
	resp, err2 := sender.sendReq(req)
	if err2 != nil {
		if err2.Error() == "503" {
			log.Errorf("error in sending request. url: %s: %s\n", req.Method, req.URL.String())
			sender.pushRequest(req)
			return nil
		} else {
			log.Errorf("error response was got. url: %s: %s; status code = %d; status: %s\n", req.Method, req.URL.String(), resp.StatusCode, resp.Status)
		}
	}
	return err2
}

func (sender RequestSender) pushRequest(request *http.Request) {
	sender.mutex.Lock()
	sender.queue.PushBack(request)
	log.Info(fmt.Sprintf("Request has added to queue. Requests in queue: %d", sender.queue.Len()))
	sender.mutex.Unlock()

	if !sender.isExecuting {
		sender.retryTimeout()
	}
}

func (sender RequestSender) retryTimeout() {
	sender.mutex.Lock()
	sender.isExecuting = true
	sender.mutex.Unlock()
	go func() {
		mutex := sender.mutex
		for !sender.isEmpty() {
			mutex.Lock()
			request := sender.queue.Front().Value.(*http.Request)
			mutex.Unlock()
			_, err := sender.sendReq(request)
			if err == nil || err.Error() != "503" {
				mutex.Lock()
				sender.queue.Remove(sender.queue.Front())
				log.Info(fmt.Sprintf("Request removed from queue. Requests in queue: %d", sender.queue.Len()))
				mutex.Unlock()
			} else {
				time.AfterFunc(config.Timeout*2, func() {
					sender.retryTimeout()
				})
				return
			}
		}

		if sender.isEmpty() {
			mutex.Lock()
			sender.isExecuting = false
			mutex.Unlock()
		}
	}()
}

func (sender RequestSender) isEmpty() bool {
	sender.mutex.Lock()
	len := sender.queue.Len()
	sender.mutex.Unlock()
	return len <= 0
}
