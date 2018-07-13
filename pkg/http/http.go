package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	ContentTypeXML  = "application/xml; charset=utf-8"
	ContentTypeJSON = "application/xml; charset=utf-8"
)

type requestHandler func(reqData map[string]interface{}) io.Reader
type responseHandler func(*http.Response) (rawData []byte)

type pool struct {
	httpClient sync.Pool
	tlsClient  sync.Pool
}

type HttpClient struct {
	clientPool     pool
	certFile       string
	keyFile        string
	rootcaFile     string
	requestHandle  requestHandler
	responseHandle responseHandler
	options        []func(*http.Client)
}

// NewHttpClient
func NewHttpClient() *HttpClient {

	return &HttpClient{

		clientPool: pool{
			httpClient: sync.Pool{
				New: func() interface{} {
					return &http.Client{
						Timeout: 5 * time.Second,
					}
				},
			},
			tlsClient: sync.Pool{
				New: func() interface{} {
					return &http.Client{
						Timeout: 5 * time.Second,
					}
				},
			},
		},
		requestHandle: func(data map[string]interface{}) io.Reader {
			bts, _ := json.Marshal(data)
			return bytes.NewBuffer(bts)
		},

		responseHandle: func(resp *http.Response) (rawData []byte) {
			if resp == nil {
				return nil
			}
			bts, _ := ioutil.ReadAll(resp.Body)
			return bts
		},
	}
}

// SetTLS 设置tls参数
func (h *HttpClient) SetTLS(certFile, keyFile, rootcaFile string) {
	h.certFile = certFile
	h.keyFile = keyFile
	h.rootcaFile = rootcaFile
}

// SetOptions 设置http client相关参数
func (h *HttpClient) SetOptions(options ...func(*http.Client)) {
	h.options = options
}

func (h *HttpClient) httpRequest(withTLS bool, do func(*http.Client)) error {
	var err error
	var trans *http.Transport
	var client *http.Client
	if withTLS {
		trans, err = h.withTLS()
		if err != nil {
			return err
		}
	}

	if withTLS {
		client = h.clientPool.tlsClient.Get().(*http.Client)
		client.Transport = trans
	} else {
		client = h.clientPool.httpClient.Get().(*http.Client)
	}

	do(client)

	if withTLS {
		h.clientPool.tlsClient.Put(client)
	} else {
		h.clientPool.httpClient.Put(client)
	}

	return nil
}

func (h *HttpClient) withTLS() (*http.Transport, error) {

	if h.certFile == "" || h.keyFile == "" {
		return nil, errors.New("certFile and keyFile cannot empty")
	}

	cert, err := tls.LoadX509KeyPair(h.certFile, h.keyFile)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(h.rootcaFile)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(data)

	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}

	conf := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		RootCAs:            pool,
	}

	trans := &http.Transport{
		TLSClientConfig: conf,
	}
	return trans, nil
}

// SetRequestHandle
func (h *HttpClient) SetRequestHandle(fn requestHandler) {
	if fn != nil {
		h.requestHandle = fn
	}

}

// SetResponseHandle 设置请求结束回调
func (h *HttpClient) SetResponseHandle(fn responseHandler) {
	if fn != nil {
		h.responseHandle = fn
	}
}

// Post http post提交 设置withTLS为true开启tls双向认证时务必先.setTLS配置参数
func (h *HttpClient) Post(url string, contentType string, withTLS bool, data map[string]interface{}) ([]byte, error) {

	var httpResp *http.Response
	var err error
	var outerErr error
	body := h.requestHandle(data)
	outerErr = h.httpRequest(withTLS, func(client *http.Client) {
		httpResp, err = client.Post(url, contentType, body)
	})
	if outerErr != nil {
		err = outerErr
	}

	if err == nil {
		defer httpResp.Body.Close()
	}

	return h.responseHandle(httpResp), err
}
