package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
)

type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

const dbName = "sql_sample"

func Connect() *sql.DB {
	const driverName = "mysql"
	const userName = "root"
	const password = "heas3real9ract2ZIRK"
	const dataSourceName = userName + ":" + password + "@/" + dbName
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err.Error())
	}

	return db
}

func Disconnect(db *sql.DB) {
	err := db.Close()
	if err != nil {
		println(err.Error())
	}
	fmt.Println("close db")
}

func RequestQuery(db *sql.DB, query string) *sql.Rows {
	rows, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	return rows
}

func CloseRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		println(err.Error())
	}
	fmt.Println("close rows")
}

func PrintColumns(rows *sql.Rows) {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			fmt.Println(columns[i], ": ", value)
		}
		fmt.Println("--------------------")
	}
}

func PrintArticles(rows *sql.Rows) {
	for rows.Next() {
		var article Article
		err := rows.Scan(&article.ID, &article.Title, &article.Desc, &article.Content)
		if err != nil {
			panic(err.Error())
		}

		fmt.Println(article.ID, article.Title, article.Desc, article.Content)
		fmt.Println("--------------------")
	}
}

func handleRequests() {
	server := http.Server{
		Addr: ":8000",
	}
	http.HandleFunc("/", homePage)
	http.HandleFunc("/articles", articles)
	http.HandleFunc("/users", users)
	err := server.ListenAndServe()
	if err != nil {
		panic(err.Error())
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		return
	}
	_, err := fmt.Fprintf(w, "Welcome to the HomePage")
	if err != nil {
		log.Fatalf("%v", err)
		return
	}
	fmt.Println("Endpoint Hit: homePage")
}

func articles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getArticles(w, r)
	case "POST":
		postArticles(w, r)
	default:
		w.WriteHeader(405)
	}
}

func getArticles(w http.ResponseWriter, r *http.Request) {
	db := Connect()
	defer Disconnect(db)

	rows := RequestQuery(db, "SELECT * FROM articles")
	defer CloseRows(rows)

	articles := make([]Article, 0)
	for rows.Next() {
		var article Article
		err := rows.Scan(&article.ID, &article.Title, &article.Desc, &article.Content)
		if err != nil {
			panic(err.Error())
		}

		articles = append(articles, article)
		fmt.Println(article.ID, article.Title, article.Desc, article.Content)
		fmt.Println("--------------------")
	}

	fmt.Println("Endpoint Hit: articles")
	err := json.NewEncoder(w).Encode(articles)
	if err != nil {
		log.Fatalf("%v", err)
	}
	w.WriteHeader(http.StatusOK)
}

func postArticles(w http.ResponseWriter, r *http.Request) {
	db := Connect()
	defer Disconnect(db)

	// `desc`が予約後のためバッククォートで括る必要がある
	stmt, err := db.Prepare("INSERT INTO articles (title, `desc`, content) VALUES (?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
			panic(err.Error())
		}
	}()

	uri := r.RequestURI
	fmt.Println(uri)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("%v", err)
	}
	var article Article
	err = json.Unmarshal(body, &article)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println(article.Title)
	fmt.Println(article.Desc)
	fmt.Println(article.Content)

	result, err := stmt.Exec(article.Title, article.Desc, article.Content)
	if err != nil {
		panic(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}

	err = json.NewEncoder(w).Encode(map[string]int{"id": int(id)})
	if err != nil {
		panic(err.Error())
	}
	w.WriteHeader(http.StatusCreated)
}

func users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getUsers(w, r)
	default:
		w.WriteHeader(405)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	db := Connect()
	defer Disconnect(db)

	rows := RequestQuery(db, "SELECT * FROM users")
	defer CloseRows(rows)

	users := make([]User, 0)
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name)
		if err != nil {
			panic(err.Error())
		}

		users = append(users, user)
		fmt.Println(user.ID, user.Name)
		fmt.Println("--------------------")
	}

	fmt.Println("Endpoint Hit: articles")
	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Fatalf("%v", err)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	handleRequests()

	//db := Connect()
	//defer Disconnect(db)
	//
	//rows1 := RequestQuery(db, "SELECT * FROM users")
	//defer CloseRows(rows1)

	// case 1
	//fmt.Println("PrintColumns")
	//PrintColumns(rows1)
	//fmt.Println("")
	//
	//rows2 := RequestQuery(db, "SELECT * FROM articles")
	//defer CloseRows(rows2)
	//// case 2
	//fmt.Println("PrintArticles")
	//PrintArticles(rows2)
}
