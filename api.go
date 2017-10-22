package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/jakubknejzlik/go-survey/model"
	"github.com/jinzhu/gorm"
)

var schema graphql.Schema

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v\n", result.Errors)
	}
	return result
}

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

	type Answer struct {
		UID    string `json:"uid"`
		Data   string `json:"data"`
		Survey string `json:"survey"`
	}

	var AnswerList []Answer

	answer1 := Answer{UID: "a", Data: "Do not forget do that!", Survey: "false"}
	answer2 := Answer{UID: "b", Data: "This is the most important stuff", Survey: "false"}
	answer3 := Answer{UID: "c", Data: "Please do not do this in a past!", Survey: "false"}
	AnswerList = append(AnswerList, answer1, answer2, answer3)

	// survey := graphql.NewObject(graphql.ObjectConfig{})

	var AnswerType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Answer",
		Fields: graphql.Fields{
			"uid": &graphql.Field{
				Type: graphql.String,
			},
			"data": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	type Survey struct {
		UID    string `json:"uid"`
		Data   string `json:"data"`
		Answer string `json:"survey"`
	}

	var SurveyList []Survey

	survey1 := Survey{UID: "j", Data: "Question1", Answer: "1"}
	survey2 := Survey{UID: "p", Data: "Question2", Answer: "2"}
	survey3 := Survey{UID: "k", Data: "Question3", Answer: "3"}
	SurveyList = append(SurveyList, survey1, survey2, survey3)

	SurveyType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Survey",
			Fields: graphql.Fields{
				"uid": &graphql.Field{
					Type: graphql.String,
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						fmt.Println("Survey Data!!!")
						return SurveyList[params.Source.(int)].UID, nil
					},
				},
				"data": &graphql.Field{
					Type: graphql.String,
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {

						fmt.Println("Survey Data!!!")
						return SurveyList[params.Source.(int)].Data, nil
					},
				},
				"answers": &graphql.Field{
					Type:        graphql.NewList(AnswerType),
					Description: "List of answers",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						return AnswerList, nil
					},
				},
			},
		})

	type Query struct {
		survey  string `json:"survey"`
		surveys string `json:"surveys"`
	}

	QueryType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:        "Query",
			Description: "Testing",
			Fields: graphql.Fields{
				"survey": &graphql.Field{
					Type: SurveyType,
					Args: graphql.FieldConfigArgument{
						"uid": &graphql.ArgumentConfig{
							Type: graphql.Int,
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						if uid, ok := params.Args["uid"].(int); ok {
							return uid, nil // ??
						}
						return nil, errors.New("Some error, but i need sleep ... ")
					},
				},
				"surveys": &graphql.Field{
					Type: graphql.NewList(SurveyType),
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						return SurveyList, nil // ??
					},
				},
			},
		})

	schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: QueryType,
		},
	)

	r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		json.NewEncoder(w).Encode(result)
	}).Methods("GET")

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
