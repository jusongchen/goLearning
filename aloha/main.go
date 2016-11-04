package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/kataras/go-template/html"

	"github.com/kataras/iris"
	"salesforce.com/dsp/dspRepo/irisx"

	"io/ioutil"
)

const (
	alohaHomeURL            = "https://aloha.my.salesforce.com"
	alohaDotForceURL        = "https://aloha.force.com"
	pathAlohav3__SAML_LOGIN = "/alohav3__SAML_LOGIN"
	pathLogin               = "/login"
	pathVerifyLogin         = "/verify_login"
	sfdcAlohaLoginURL       = alohaDotForceURL + pathAlohav3__SAML_LOGIN
)

func main() {
	startIris(3000)
}

//Start starts an http Server
func startIris(consolePort int) error {

	// Serve
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%d", consolePort))
	if err != nil {
		panic(err)
	}
	irisFW := irisx.NewFramework()
	registerHandles(irisFW, nil)

	if err := irisFW.Serve(ln); err != nil {
		panic(err)
	}
	return nil
}

//Get is a customerized http.Get
func Get(urlStr string) (*http.Response, error) {
	req := http.Request{
		Header: http.Header{},
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
	req.Method = "GET"
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("Get:%v", err)
	}
	req.URL = u
	return http.DefaultClient.Do(&req)
}

func registerHandles(irisFW *iris.Framework, db *sqlx.DB) {

	// directory and extensions defaults to ./templates, .html for all template engines
	irisFW.UseTemplate(html.New(html.Config{Layout: "layouts/layout.html"})).Directory("./templates", ".html")

	//svr.Config.Render.Template.Gzip = true
	// irisFW.Post("/message", func(ctx *iris.Context) {

	// 	log.Printf("Header: %v\n", ctx.Response.Header)
	// 	// r.FormValue("userid"), r.FormValue("message")
	// })

	//svr.Config.Render.Template.Gzip = true
	irisFW.Get("/", func(ctx *iris.Context) {

		ctx.MustRender("index.html", nil)
	})

	//svr.Config.Render.Template.Gzip = true
	irisFW.Get(pathLogin, func(ctx *iris.Context) {

		// ctx.MustRender("login.html", nil)

		// return

		bodyBytes, err := Get(alohaHomeURL)
		if err != nil {
			log.Fatal(err)
		}
		defer bodyBytes.Body.Close()

		// io.Copy(os.Stdout, bodyBytes.Body)
		resp, err := ioutil.ReadAll(bodyBytes.Body)

		if err != nil {
			log.Fatal(err)
		}
		// http.PostForm(sfdcAlohaLoginURL,)
		s := strings.Replace(string(resp), sfdcAlohaLoginURL, pathLogin, -1)

		// s := string(resp)
		log.Println(s)
		ctx.HTML(iris.StatusOK, string(s))
	})

	irisFW.Post(pathLogin, func(ctx *iris.Context) {

		data := ctx.PostValuesAll()
		// data :=url.Values{}
		// data[]
		resp, err := http.PostForm(sfdcAlohaLoginURL, data)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		// io.Copy(os.Stdout, bodyBytes.Body)
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		s := strings.Replace(string(bodyBytes), sfdcAlohaLoginURL, pathVerifyLogin, -1)
		// s := string(resp)
		// log.Println(s)
		ctx.HTML(iris.StatusOK, string(s))

	})

	irisFW.Post(pathVerifyLogin, func(ctx *iris.Context) {

		data := ctx.PostValuesAll()
		// data :=url.Values{}
		// data[]

		log.Printf("PostForm data %v", data)
		resp, err := http.PostForm(sfdcAlohaLoginURL, data)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		// io.Copy(os.Stdout, bodyBytes.Body)
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		s := string(bodyBytes)
		log.Println(s)
		ctx.HTML(iris.StatusOK, string(s))

	})

}
