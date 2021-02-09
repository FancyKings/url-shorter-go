package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Database struct and json map
type UrlMap struct {
	LongMd5  string `json:"longmd5"`
	LongUri  string `json:"longurl"`
	InnerUrl string `json:"innerurl"`
	Time     string `json:"time"`
	Count    int    `json:"count"`
}

func main() {

	// Echo instance
	e := echo.New()

	// Redirect to https
	//e.Pre(middleware.HTTPSRedirect())

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	// Intercept routing parameters uuid
	// Reference: https://www.yuque.com/qiuquanwu/lz8ker/gwt254
	// Reference: https://www.cnblogs.com/tomtellyou/p/12874530.html
	e.GET("/:innerUrl", urlRedirect)
	// Temp redirect /new
	e.POST("/", addUrl)

	// Start call Database
	//sqliteDb()

	// Start server port 80
	go func() {
		e.Logger.Fatal(
			e.Start(":80"),
		)
	}()
	//Start server tls ERROR
	e.Logger.Fatal(
		e.StartTLS(":443",
			"crt/local.pem",
			"crt/local-key.pem"),
	)
}

// Handler url redirect
func urlRedirect(c echo.Context) error {

	// Get database pointer
	db := initSqliteConnect()

	// Parsing url param uuid(inner url)
	innerUrl := c.Param("innerUrl")
	fmt.Println("fuc urlRedirect innerUrl : " + innerUrl)

	row, err := db.Query(
		"SELECT * FROM dbtable WHERE innerurl ='" + innerUrl + "'",
	)
	checkErr(err)
	defer row.Close()

	// Test
	singleMap := new(UrlMap)
	for row.Next() {
		err = row.Scan(&singleMap.LongMd5, &singleMap.LongUri, &singleMap.InnerUrl,
			&singleMap.Time, &singleMap.Count,
		)
		checkErr(err)
		fmt.Println("func Url Redirect : " + singleMap.LongUri)
	}

	// Use HTTP status code 307 for URL redirection
	err = c.Redirect(http.StatusTemporaryRedirect, singleMap.LongUri)
	checkErr(err)

	return c.JSON(http.StatusOK, singleMap)
	//return c.String(http.StatusOK, singleMap.LongUri)
}

// Handler add url
func addUrl(c echo.Context) error {

	// Get database pointer
	db := initSqliteConnect()

	// Get current time
	currTime := time.Now().UTC()
	currTimeFormat := currTime.String()

	// Parsing fields origin_url
	originUrl := c.FormValue("origin_url")
	fmt.Println("Func addUrl LongUrl : " + originUrl)

	// Prepare database templates
	// For the initial, Url count field initialize to 1
	// Since the original url calculates md5 sum as the primary key,
	// to detect duplicates
	insertDataTemplate, err := db.Prepare(
		`INSERT INTO dbtable (longmd5,longurl, innerurl, time, count) values(?, ?, ?, ?, 1)`,
	)
	checkErr(err)

	// Generate short jump links, temporarily use UUID instead
	innerUrl := uuid.NewString()
	// Json parse "-" error, so delete "-"
	innerUrl = strings.ReplaceAll(innerUrl, "-", "")

	// Calculate original link Md5
	longUrlHash := md5.Sum([]byte(originUrl))
	longUrlMd5 := hex.EncodeToString(longUrlHash[:])

	// Just For Test
	fmt.Println("Func addUrl InnerUrl : " + innerUrl)
	fmt.Println("Func addUrl LongUrlMd5 : " + longUrlMd5)

	// Performing database insert operations
	_, err = insertDataTemplate.Exec(
		longUrlMd5, originUrl, innerUrl, currTimeFormat)

	// https://draveness.me/golang-defer/
	// To learn
	defer db.Close()

	return c.JSON(
		http.StatusOK,
		UrlMap{
			longUrlMd5, originUrl, innerUrl, currTimeFormat, 1,
		},
	)
}

// Sqlite database initialization
// Refer: https://www.codeproject.com/Articles/5261771/Golang-SQLite-Simple-Example
// No use gorm
func initSqliteConnect() *sql.DB {
	db, err := sql.Open("sqlite3",
		"db/url-shorter-go.db")
	checkErr(err)
	return db
}

// Deprecate: Print the complete row of the database
func printDbLine(id int, time string, originUrl string, innerUrl string, count int) {
	fmt.Print(
		strconv.Itoa(id) + "---" +
			time + "---" +
			originUrl + "---" +
			innerUrl + "---" +
			strconv.Itoa(count),
	)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("ERROR")
		panic(err)
	}
}
