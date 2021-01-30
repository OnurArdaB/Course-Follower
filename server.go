package main

import (
	"strings"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"io/ioutil"
	"time"
	"database/sql"
	"strconv"
)
import _ "github.com/go-sql-driver/mysql"

type Tag struct {
    id   int    `json:"id"`
	c_id string `json:"c_id"`
	email string `json:"email"`
	timestamp string `json:"reg_date"`
}

func QueryMaker(){

	time.Sleep(10 * time.Second)
	fmt.Println("Go Routine Starting...")

	db, err := sql.Open("mysql", "")
	
	if err != nil {
        panic(err.Error())
	}

	defer db.Close()

	for true {
		fmt.Println("Queue restarting...")
		selected, err_Q := db.Query("SELECT * FROM QUEUE")
	
		if err_Q != nil {
			panic(err_Q.Error())
		}
		for selected.Next() {
			var tag Tag
			err = selected.Scan(&tag.id,&tag.c_id ,&tag.email,&tag.timestamp)
			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}
			//start making a query
			url:="https://suis.sabanciuniv.edu/prod/bwckschd.p_disp_detail_sched?term_in=202001&crn_in=" + tag.c_id
			fmt.Println(url)
			resp, err := http.Get(url)
			if err != nil {
				panic(err.Error())
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			if strings.Contains(bodyString,"Seats"){
				seated:=strings.Split(bodyString,"Seats")[1]
				reparsed:=strings.Split(seated,"dddefault")
				if strings.Contains(reparsed[3],">0<"){
					//not empty
					fmt.Println("Not Empty :(")
				}else{
					//mail user
					from := ""
					password := ""

					// Receiver email address.
					to := []string{
						tag.email,
					}

					// smtp server configuration.
					smtpHost := "smtp.gmail.com"
					smtpPort := "587"

					// Message.
					message := []byte("Course with CRN " + tag.c_id + " has an empty seat.")
					
					// Authentication.
					auth := smtp.PlainAuth("", from, password, smtpHost)
					
					// Sending email.
					err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println("Email Sent Successfully!")
					fmt.Println("DELETE FROM QUEUE WHERE id="+strconv.Itoa(tag.id))
					_, err_D :=db.Query("DELETE FROM QUEUE WHERE id="+strconv.Itoa(tag.id))
					if err_D != nil {
						panic(err_D.Error())
					}
				}
			}else{
				fmt.Println(bodyString)
			}
		}
		defer selected.Close()
		time.Sleep(60 * time.Second)
	}
}

func OperationSuccessfull(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET"{
		http.ServeFile(w,r,"static/message.html")
	}else{
		http.Error(w,"Method is not supported.",http.StatusNotFound)
		return
	}
}

func LandingHandler(w http.ResponseWriter, r *http.Request){

	if r.Method == "GET"{
		http.ServeFile(w,r,"static/index.html")
	}else if r.Method == "POST"{
		//controller
		err := r.ParseForm()
		if err != nil {
			// in case of any error
			return
		}
		crn:= r.Form.Get("crn")
		email:= r.Form.Get("email")
		term:=r.Form.Get("term")
		fmt.Println(crn,email,term)
		if crn!=""&&term!=""&&email!="" {
			//ok
			db, err := sql.Open("mysql", "")

			// if there is an error opening the connection, handle it
			if err != nil {
				panic(err.Error())
			}
		
			// defer the close till after the main function has finished
			// executing
			defer db.Close()
			query := "INSERT INTO QUEUE(c_id,email) VALUES ( '" + crn + "', '" + email +"' )"
			insert, err := db.Query(query)
			
			if err != nil {
				panic(err.Error())
			}
			defer insert.Close()
			http.Redirect(w, r, "/message", http.StatusSeeOther)
		}else{
			http.Error(w, "You have given wrong format of input", http.StatusBadRequest)
		}
	}else{
		http.Error(w,"Method is not supported.",http.StatusNotFound)
		return
	}
}

func main(){
	http.HandleFunc("/",LandingHandler)
	http.HandleFunc("/message",OperationSuccessfull)
	go QueryMaker()
	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080",nil); err != nil{
	log.Fatal(err)
	}
}
