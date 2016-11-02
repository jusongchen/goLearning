// Package main provide one-line integration with letsencrypt.org
package main

import "github.com/kataras/iris"

func main() {
	iris.Get("/", func(ctx *iris.Context) {
		ctx.Write("Hello from SECURE SERVER!")
	})

	iris.Get("/test2", func(ctx *iris.Context) {
		ctx.Write("Welcome to secure server from /test2!")
	})

	// This will provide you automatic certification & key from letsencrypt.org's servers
	// it also starts a second 'http://' server which will redirect all 'http://$PATH' requests to 'https://$PATH'
	iris.ListenLETSENCRYPT(":443")
}

// func (s *Framework) ListenLETSENCRYPT(addr string) {
// 	addr = ParseHost(addr)
// 	if s.Config.VHost == "" {
// 		s.Config.VHost = addr
// 		// this will be set as the front-end listening addr
// 	}
// 	ln, err := LETSENCRYPT(addr)
// 	if err != nil {
// 		s.Logger.Panic(err)
// 	}

// 	// starts a second server which listening on :80 to redirect all requests to the :443 (https://)
// 	Proxy(":80", "https://"+addr)
// 	s.Must(s.Serve(ln))
// }
