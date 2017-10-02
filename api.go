package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jakubknejzlik/go-survey/model"
	"github.com/jinzhu/gorm"
)

func getRouter(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	// f := r.PathPrefix("/files").Subrouter()
	r.HandleFunc("/", handleIndex())
	r.HandleFunc("/editor", handleEditor()).Methods("GET")
	r.HandleFunc("/survey", handleSurvey()).Methods("GET")

	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataPUT(db)).Methods("PUT")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataPUT(db)).Methods("PUT")

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
		survey := model.Survey{}
		db.FirstOrInit(&survey, model.Survey{Uid: uid})
		data, _ := ioutil.ReadAll(r.Body)
		survey.Data = string(data)
		db.Save(&survey)
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
		survey := model.Survey{}
		answer := model.Answer{}

		err := db.Where("Uid = ?", surveyUID).First(&survey).Error
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "answer not found!", http.StatusNotFound)
			return
		}

		db.FirstOrInit(&answer, model.Answer{Uid: answerUID})

		data, _ := ioutil.ReadAll(r.Body)
		answer.Data = string(data)
		answer.Survey = survey
		db.Save(&answer)
		w.WriteHeader(http.StatusNoContent)
	}
}
