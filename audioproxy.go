package main

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

const AudioProxyPrefix = "/audioproxy/"

// AudioProxy is the proxy for audio file store.
type AudioProxy struct {
	Name    string
	Address string
}

func (p *AudioProxy) String() string {
	return p.Name + "=" + p.Address
}

// Hijack hijacks url to proxied url.
//
//	p := &AudioProxy{ Name: "foo", Address: "http://127.0.0.1:6666/bar/" }
//	p.Hijack("http://127.0.0.1:6666/bar/song.mp3") // -> "/audioproxy/bar/song.mp3", true
//	p.Hijack("http://example.com/bar/song.mp3")    // -> "http://example.com/bar/song.mp3", false
//
// The hijacked url is started with "/".
// It's client's duty to add the scheme and host (to this server) at the beginning.
func (p *AudioProxy) Hijack(url string) (newurl string, ok bool) {
	if !strings.HasPrefix(url, p.Address) {
		return url, false
	}

	file, _ := strings.CutPrefix(url, p.Address)

	return AudioProxyPrefix + p.Name + "/" + file, true
}

// proxyFromString parse "Name=Address" to AudioFileStoreProxy.
func proxyFromString(s string) (*AudioProxy, error) {
	ss := strings.Split(s, "=")
	if len(ss) != 2 {
		return nil, errors.New("invalid proxy string")
	}

	name := strings.TrimSpace(ss[0])
	address := strings.TrimSpace(ss[1])

	return &AudioProxy{
		Name:    name,
		Address: address,
	}, nil
}

// proxiesFromString parse "Name=Address,Name=Address" to []*AudioFileStoreProxy.
func proxiesFromString(s string) ([]*AudioProxy, error) {
	ss := strings.Split(s, ",")
	proxies := make([]*AudioProxy, 0, len(ss))
	for _, s := range ss {
		p, err := proxyFromString(s)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, p)
	}
	return proxies, nil
}

var (
	proxies        = []*AudioProxy{}
	proxyByName    = map[string]*AudioProxy{}
	proxyByAddress = map[string]*AudioProxy{} // unused
)

func SetupAudioProxies(s string) {
	if s == "" {
		return
	}
	proxies, _ = proxiesFromString(s)
	for _, p := range proxies {
		proxyByName[p.Name] = p
		proxyByAddress[p.Address] = p
	}
}

func init() {
	s := os.Getenv("AUDIO_PROXIES")
	if s != "" {
		SetupAudioProxies(s)
	}
}

// RregisterAudioProxyTo register audio proxy handlers.
//
//	GET /audioproxy/:proxyName/*file
func RregisterAudioProxyTo(r gin.IRouter) {
	r.GET(AudioProxyPrefix+":proxyName/*file", func(c *gin.Context) {
		proxyName := c.Param("proxyName")

		// log.Printf("[DBG] handle GET /audioproxy: proxyName: %v", proxyName)
		// log.Printf("[DBG] proxies: %#v", proxies)

		if proxy, ok := proxyByName[proxyName]; ok {
			httpProxy(c, proxy.Address, AudioProxyPrefix+proxyName)
			return
		}

		c.JSON(404, gin.H{
			"error": "not found",
		})
	})
}

// httpProxy proxies http request to upstream
func httpProxy(c *gin.Context, upstream string, trimPrefix string) {
	remote, err := url.Parse(upstream)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	//Define the director func
	//This is a good place to log, for example
	stdDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// req.Header = c.Request.Header
		// req.Host = remote.Host
		// req.URL.Scheme = remote.Scheme
		// req.URL.Host = remote.Host

		stdDirector(req)
		req.URL.Path = strings.Replace(req.URL.Path, trimPrefix, "", 1)
	}

	// proxy.ServeHTTP will panic on "net/http: abort Handler"
	// when it failed to write response to client.
	// e.g. curl get a binary file
	//  => "Warning: Binary output can mess up your terminal"
	//  => panic: net/http: abort Handler
	// So we recover it.

	defer func() {
		if err := recover(); err != nil {
			slog.Error("httpProxy panic (recovered)",
				"err", err,
				"url", c.Request.URL.String(),
				"upstream", upstream)
		}
	}()

	proxy.ServeHTTP(c.Writer, c.Request)
}
