package main

import (
	"context"
	"math"
	"strings"

	"time"

	"fmt"

	"strconv"

	"log"

	"net/http"

	"github.com/gorilla/mux"

	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"

	"html/template"

	"Personal-Web/connection"

	"Personal-Web/middleware"
)

// tipe data 'map' dengan key bertipe string dan value bertipe interface
var Data = map[string]interface{}{
	"Title":   "Personal Web",
	"IsLogin": false,
}

// kumpulan definisi variables
type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

// kumpulan definisi variables
// var pointer => beisi alamat memori/reference suatu nilai
type Project struct {
	Id           int
	Title        string
	Start_date   time.Time
	End_date     time.Time
	Description  string
	Technologies []string
	Image        string
	DurationText string
	IsLogin      bool
}

//type Projects []Project

//func NewProject() *Project {
//	return &Project{
//		Start_date: time.Date(2022, 5, 12, 21, 0, 0, 0, time.Local),
//		End_date:   time.Date(2022, 5, 12, 21, 0, 0, 0, time.Local),
//	}
//}

//var Projects = []Project{
//	{
//		Title:       "Pembelajaran Online",
//		Duration:    "Duration : 3 Weeks",
//		Author:      " | Bagas",
//		Description: "Sangat sulit sekali hehehehe",
//	},
//}

// function routing
func main() {
	router := mux.NewRouter()

	// Connect database
	connection.DatabaseConnect()

	// Create Folder
	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/Project", project).Methods("GET")
	router.HandleFunc("/addProject", middleware.UploadFile(addProject)).Methods("POST")
	router.HandleFunc("/contactMe", contactMe).Methods("GET")
	router.HandleFunc("/projectDetail/{id}", projectDetail).Methods("GET")
	router.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	router.HandleFunc("/edit-project-new/{id}", EditProject).Methods("GET")
	router.HandleFunc("/edit-project/{id}", middleware.UploadFile(EditProjectForm)).Methods("POST")
	router.HandleFunc("/signForm", signupForm).Methods("GET")
	router.HandleFunc("/signUp", signUp).Methods("POST")
	router.HandleFunc("/loginForm", loginForm).Methods("GET")
	router.HandleFunc("/loginForm", login).Methods("POST")
	router.HandleFunc("/logout", Logout).Methods("GET")

	fmt.Println("Server Running Successfully")
	http.ListenAndServe("localhost:5000", router)
}

// function handling index.html
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//Parsing template html file
	var tmpl, err = template.ParseFiles("index.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	fm := session.Flashes("message")

	var flashes []string

	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
	Data["FlasData"] = strings.Join(flashes, "")

	//membaca record data, query menghasilkan semua baris dan kolom
	rows, _ := connection.Conn.Query(context.Background(), "SELECT id, title, start_date, end_date, description, technologies, image FROM public.tb_project;")
	//context.Background() => menjadi parent/induk dari content context yang lain.
	var result []Project
	for rows.Next() {
		var each = Project{}
		//untuk mendapatkan value
		var err = rows.Scan(&each.Id, &each.Title, &each.Start_date, &each.End_date, &each.Description, &each.Technologies, &each.Image)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.DurationText = CalculateDuration(each.Start_date, each.End_date)

		var store = sessions.NewCookieStore([]byte("SESSION_ID"))
		session, _ := store.Get(r, "SESSION_ID")

		if session.Values["IsLogin"] != true {
			each.IsLogin = false
		} else {
			each.IsLogin = session.Values["IsLogin"].(bool)
		}

		//each.Format_start = each.Start_date.Format("12 May 2001")
		//each.Format_end = each.End_date.Format("12 May 2001")

		//		if session.Values["IsLogin"] != true {
		//			each.IsLogin = false
		//		} else {
		//			each.IsLogin = session.Values ["IsLogin"].(bool)
		//		}

		result = append(result, each)
	}

	resp := map[string]interface{}{
		"Title":    Data,
		"Data":     Data,
		"Projects": result,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

// function handling myproject.html
func project(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//Parsing template html file
	var tmpl, err = template.ParseFiles("myproject.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	resp := map[string]interface{}{
		"Title": Data,
		"Data":  Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

// function handling contactMe.html
func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//Parsing template html file
	var tmpl, err = template.ParseFiles("contactMe.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	resp := map[string]interface{}{
		"Title": Data,
		"Data":  Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

// function handling myproiect-detail.html
func projectDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	//Parsing template html file
	var tmpl, err = template.ParseFiles("myproject-detail.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ProjectDetail := Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM public.tb_project WHERE id=$1", id).
		Scan(&ProjectDetail.Id, &ProjectDetail.Title, &ProjectDetail.Start_date, &ProjectDetail.End_date, &ProjectDetail.Description, &ProjectDetail.Technologies, &ProjectDetail.Image)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message: " + err.Error()))
		return
	}
	DurationText := CalculateDuration(ProjectDetail.Start_date, ProjectDetail.End_date)
	StartDate := ProjectDetail.Start_date.Format("Jan 2, 2006")
	EndDate := ProjectDetail.End_date.Format("Jan 2, 2006")

	resp := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
		"DurationText":  DurationText,
		"StartDate":     StartDate,
		"EndDate":       EndDate,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

// function handling add myproject
func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("title")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("description")
	technologies := r.Form["tec"]

	dataContext := r.Context().Value("dataFile")
	image := dataContext.(string)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(title, start_date, end_date, description, technologies, image) VALUES($1, $2, $3, $4, $5, $6)", title, startDate, endDate, description, technologies, image)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// function handling delete project
func deleteProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// function handling sign up form
func signupForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//Parsing template html file
	var tmpl, err = template.ParseFiles("signUp.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	tmpl.Execute(w, nil)
}

// function handling sign up
func signUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 15)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_user (name, email, password) VALUES($1, $2, $3)", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/loginForm", http.StatusMovedPermanently)
}

// function handling login form
func loginForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//Parsing template html file
	var tmpl, err = template.ParseFiles("login.html")
	//Error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

// function handling login
func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).
		Scan(&user.Id, &user.Name, &user.Email, &user.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Options.MaxAge = 10800

	session.AddFlash("Login Successful", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func CalculateDuration(StartDate time.Time, EndDate time.Time) string {
	Duration := EndDate.Sub(StartDate)
	DurationHours := Duration.Hours()
	DurationDays := math.Floor(DurationHours / 24)
	DurationWeeks := math.Floor(DurationDays / 7)
	DurationMonths := math.Floor(DurationDays / 30)
	var DurationText string
	if DurationMonths > 1 {
		DurationText = strconv.FormatFloat(DurationMonths, 'f', 0, 64) + " months"
	} else if DurationMonths > 0 {
		DurationText = strconv.FormatFloat(DurationMonths, 'f', 0, 64) + " month"
	} else {
		if DurationWeeks > 1 {
			DurationText = strconv.FormatFloat(DurationWeeks, 'f', 0, 64) + " weeks"
		} else if DurationWeeks > 0 {
			DurationText = strconv.FormatFloat(DurationWeeks, 'f', 0, 64) + " week"
		} else {
			if DurationDays > 1 {
				DurationText = strconv.FormatFloat(DurationDays, 'f', 0, 64) + " days"
			} else if DurationDays > 0 {
				DurationText = strconv.FormatFloat(DurationDays, 'f', 0, 64) + " day"
			} else {
				DurationText = "less than a day"
			}
		}
	}
	return DurationText
}

func EditProjectForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	title := r.PostForm.Get("title")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("description")
	technologies := r.Form["tec"]

	StartDate, _ := time.Parse("2006-01-02", startDate)
	EndDate, _ := time.Parse("2006-01-02", endDate)

	dataContex := r.Context().Value("dataFile")
	Image := dataContex.(string)

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_project SET title=$1, start_date=$2, end_date=$3, description=$4, technologies=$5, image=$6, WHERE id=$;", title, StartDate, EndDate, description, technologies, Image, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func EditProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var tmpl, err = template.ParseFiles("myproject-edit.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data["IsLogin"] = false
	} else {
		Data["IsLogin"] = session.Values["IsLogin"].(bool)
		Data["Username"] = session.Values["Name"].(string)
	}

	ProjectDetail := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", id).Scan(
		&ProjectDetail.Id, &ProjectDetail.Title, &ProjectDetail.Start_date, &ProjectDetail.End_date, &ProjectDetail.Description, &ProjectDetail.Technologies, &ProjectDetail.Image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var Node, Java, Php, Laravel bool
	for _, technology := range ProjectDetail.Technologies {
		if technology == "node" {
			Node = true
		}
		if technology == "java" {
			Java = true
		}
		if technology == "php" {
			Php = true
		}
		if technology == "laravel" {
			Laravel = true
		}
	}

	StartDateString := ProjectDetail.Start_date.Format("2006-01-02")
	EndDateString := ProjectDetail.End_date.Format("2006-01-02")

	resp := map[string]interface{}{
		"Data":            Data,
		"Id":              id,
		"ProjectDetail":   ProjectDetail,
		"StartDateString": StartDateString,
		"EndDateString":   EndDateString,
		"Node":            Node,
		"Java":            Java,
		"Php":             Php,
		"Laravel":         Laravel,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-chace, no-store, must-revalidate")

	var store = sessions.NewCookieStore([]byte("SESSION_ID"))
	session, _ := store.Get(r, "SESSION_ID")

	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
