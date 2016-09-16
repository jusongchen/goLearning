package main
	
	
import (
	"database/sql"
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/binding"
	"net/http"
	"log"
	"flag"
	_ "code.google.com/p/odbc"
	"fmt"
	"strconv"
	"os"
)

var (
	mssrv    = flag.String("mssrv", "localhost", "ms sql server name")
	msdb     = flag.String("msdb", "goDemo", "ms sql server database name")
	msuser   = flag.String("msuser", "", "ms sql server user name; trusted connection used if not specified ")
	mspass   = flag.String("mspass", "", "ms sql server password")
//	msdriver = flag.String("msdriver", "sql server", "ms sql odbc driver name")
	port   = flag.String("port", "80", "web server port number")
)



type Book struct {
	Title		string	`form:"title"`
	Author      string	`form:"author"`
	Description string	`form:"description"`
}



// This method implements binding.Validator and is executed by the binding.Validate middleware
func (bk Book) Validate(errors *binding.Errors, req *http.Request) {

	if len(bk.Title) < 4 {
		errors.Fields["title"] = "Field Title Too short; minimum 4 characters"
		}
	if len(bk.Title) > 255 {
		errors.Fields["title"] = "Field Title Too long; maximum 255 characters"
		}
	if len(bk.Author) < 2 {
		errors.Fields["Author"] = "Field Author too short; minimum 2 characters"
		}
	if  len(bk.Author) > 40 {
		errors.Fields["Author"] = "Field Author too long; maximum 40 characters"
		}
	if len(bk.Description) > 2000 {
		errors.Fields["Description"] = "Field Author too long; maximum 2000 characters"
		}
}


type SearchPattern  struct {
	Pattern string	`form:"pattern"`
}

func (p SearchPattern) Validate(errors *binding.Errors, req *http.Request) {
	if len(p.Pattern) < 1 {
		errors.Fields["Pattern"] = "Field Too short; minimum 1 characters"
		}
}


func serverVersion(db *sql.DB) (sqlVersion string, err error) {
	var v string
	if err = db.QueryRow("select @@version").Scan(&v); err != nil {
		return "", err
	}	else {
		return v,nil
	}
}	
	
	

func SetupDB() (db *sql.DB,err error) {

	params := map[string]string{
		"driver":   "sql server",
		"server":   *mssrv,
		"database": *msdb,
		//"port": 	*msport,
		}
		
	if len(*msuser) == 0 {
			params["trusted_connection"] = "yes"
	} else {
			params["uid"] = *msuser
			params["pwd"] = *mspass
	}
	
	var c string
	for n, v := range params {
		c += n + "=" + v + ";"
	}

	return  sql.Open("odbc", c)
	
}



func SearchBooks (pattern SearchPattern,ren render.Render,r *http.Request) {
ren.Redirect("/demo/books/?search="+pattern.Pattern)
}


func NewBooks(r render.Render) {
	r.HTML(200, "create", nil)
}


func getLastInsertId(db *sql.DB) (id int64,err error) {
	if err = db.QueryRow("select @@IDENTITY").Scan(&id); err != nil {
		return 0, err
	}	else {
		return id,nil
	}
}


func CreateBook(book Book,ren render.Render, r *http.Request, db *sql.DB) {
	//log.Print("Title:",book.Title);
	result, err := db.Exec("INSERT INTO books (title, author, description) VALUES (?, ?, ?)",
		book.Title,book.Author,book.Description)

	var id int64 // id of new create book
	if err != nil {
		log.Printf("SQL statement execution failed:\n%s",err)
	} else {
		if id,err = result.LastInsertId(); err == nil {
			log.Printf("SQL result.LastInsertId()=%d",id)			
		} else {
			id,err = getLastInsertId(db)
		}
	}
	if err == nil {
		ren.Redirect(fmt.Sprintf("/demo/books/id/%d",id))
	} else {	
		ren.Redirect("/demo/books")
	}	
}

func ShowBooks(ren render.Render, r *http.Request,db *sql.DB,params martini.Params) {
	var (
		rows *sql.Rows
		err error
		)
	
	bookId,ok := params["id"]
	if  ok {
		//book id passed in URL
		if id,err := strconv.Atoi(bookId); err == nil {
			//the id is a number
			rows, err = db.Query(`SELECT title, author, description 
			   FROM books 
			   WHERE id =?`, id)
			  } else{
			// bad book id passed in
			ren.HTML(200, "books", nil)
			return
		}
		
	} else {
		//search by pattern
		search := r.URL.Query().Get("search")
		if search == "" {
			ren.HTML(200, "books", nil)
			return
			}
		search = "%" +search + "%"
		rows, err = db.Query(`SELECT title, author, description 
						   FROM books 
						   WHERE title like ?
						   OR author like ?
						   OR description like ?`, search,search,search)
	}
	
	if err != nil {
		log.Fatalf("SQL execution failed:%s\n",err)
		}
		
	defer rows.Close()
	books := []Book{}
	for rows.Next() {
		book := Book{}
		err := rows.Scan(&book.Title, &book.Author, &book.Description)
		if err != nil {
			log.Fatalf("DB rows scan failed\n%s",err)
		}
		books = append(books, book)
	}
	ren.HTML(200, "books", books)
}

func main() {

	flag.Parse() //   SetupDB()
	db,err := SetupDB(); 
	if err != nil {
		log.Fatal("connect to DB failed:\n",err)
		return 
	}
	
	if ver,err := serverVersion(db); err != nil {
		log.Fatal("connect to SQL server failed:\n",err)
		return 
	} else {	
		log.Printf("connected to SQL server (host=%s, database=%s)\n%s", *mssrv,*msdb,ver)
	}
	
	m := martini.Classic()
	m.Map(db)
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))
	
	m.Get("/demo/books", ShowBooks)
	m.Get("/demo/books/id/:id", ShowBooks)
	m.Get("/demo/books/search",binding.Bind(SearchPattern{}) ,SearchBooks)
	m.Post("/demo/books/create",binding.Bind(Book{}), CreateBook)
	m.Get("/demo/books/create", NewBooks)

	os.Setenv("PORT",*port)
	
	m.Run()
}
