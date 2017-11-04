package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/handler"
	"github.com/jakubknejzlik/go-survey/model"
	"github.com/jinzhu/gorm"
)

type Answer struct {
	UID    string `json:"uid"`
	Data   string `json:"data"`
	Survey Survey `json:"survey"`
}

type Survey struct {
	UID     string `json:"uid"`
	Data    string `json:"data"`
	Answers []Answer
}

func getRouter(db *gorm.DB) *mux.Router {

	r := mux.NewRouter()

	r.HandleFunc("/", handleIndex())
	r.HandleFunc("/editor", handleEditor()).Methods("GET")
	r.HandleFunc("/survey", handleSurvey()).Methods("GET")

	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataPUT(db)).Methods("PUT")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataPUT(db)).Methods("PUT")

	schema := getSchema(db)

	graphqlHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	r.HandleFunc("/graphql", graphqlHandler.ServeHTTP)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}

func handleIndex() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	}
}

func handleEditor() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadFile("views/editor.html")
		w.Write(data)
	}
}

func handleSurvey() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadFile("views/survey.html")
		w.Write(data)
	}
}

func handleSurveyDataGET(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["surveyUID"]
		survey := model.Survey{}
		err := db.Where("Uid = ?", uid).First(&survey).Error
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "not found!", http.StatusNotFound)
			return
		}
		data := survey.Data
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(data))
	}
}

func handleSurveyDataPUT(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["surveyUID"]

		data, _ := ioutil.ReadAll(r.Body)
		survey := model.Survey{
			Uid:  uid,
			Data: string(data),
		}

		err := db.Save(&survey).Error
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleAnswersDataGET(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		uid := vars["answerUID"]
		answer := model.Answer{}
		err := db.Where("Uid = ?", uid).First(&answer).Error
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "answer not found!", http.StatusNotFound)
			return
		}
		data := answer.Data
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(data))
	}
}

func handleAnswersDataPUT(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		surveyUID := vars["surveyUID"]
		answerUID := vars["answerUID"]

		err := db.Where("Uid = ?", surveyUID).First(&model.Survey{}).Error
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "answer not found!", http.StatusNotFound)
			return
		}

		data, _ := ioutil.ReadAll(r.Body)

		answer := model.Answer{
			Uid:       answerUID,
			Data:      string(data),
			SurveyUid: surveyUID,
		}

		db.Save(&answer)

		w.WriteHeader(http.StatusNoContent)
	}
}
