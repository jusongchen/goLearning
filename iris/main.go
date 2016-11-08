package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/kataras/go-template/html"
	"github.com/kataras/iris"
	"salesforce.com/dsp/irisx"
)

func main() {

	irisx.DefaultConfig.IsDevelopment = true
	fw := irisx.NewFramework(&irisx.DefaultConfig)
	registerHandles(fw)
	err := irisx.ListenAndServe(":3000", fw)
	log.Fatal(err)

}

func registerHandles(irisFW *iris.Framework) {

	// , db *sqlx.DB

	// directory and extensions defaults to ./templates, .html for all template engines
	irisFW.UseTemplate(html.New(html.Config{Layout: "layouts/layout.html"})).Directory("./templates", ".html")

	//svr.Config.Render.Template.Gzip = true
	irisFW.Get("/", func(ctx *iris.Context) {

		ctx.MustRender("index.html", nil)
	})

	const (
		pathLogin = "/login"
	)

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
		// s := string(resp)
		s := strings.Replace(string(resp), actionURL, pathLogin, -1)

		// s := string(resp)
		// log.Println(s)
		// ctx.Text(iris.StatusOK, string(s))
		ctx.HTML(iris.StatusOK, string(s))
	})

	irisFW.Post(pathLogin, func(ctx *iris.Context) {

		data := ctx.PostValuesAll()
		// data :=url.Values{}
		// data[]
		resp, err := http.PostForm(actionURL, data)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		// io.Copy(os.Stdout, bodyBytes.Body)
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		s := strings.Replace(string(bodyBytes), actionURL, pathLogin, -1)

		// s := string(bodyBytes)
		log.Println(s)

		if strings.Contains(s, "https://org62.my.salesforce.com/home/home.jsp") {
			log.Printf("\n********password verified\n")
		} else {
			log.Printf("\n********password NOT verified\n")
		}

		ctx.HTML(iris.StatusOK, string(s))

	})

}
