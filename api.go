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

type Query struct {
	Survey  string `json:"survey"`
	Surveys string `json:"surveys"`
}

var SurveyList []Survey
var AnswerList []Answer
var schema graphql.Schema
var SurveyType *graphql.Object

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

var AnswerType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AnswerType",
	Fields: graphql.FieldsThunk(func() graphql.Fields {
		return graphql.Fields{
			"uid": &graphql.Field{
				Type: graphql.String,
			},
			"data": &graphql.Field{
				Type: graphql.String,
			},
			"survey": &graphql.Field{
				Type: SurveyType,
			},
		}
	}),
})

func getRouter(db *gorm.DB) *mux.Router {

	r := mux.NewRouter()

	r.HandleFunc("/", handleIndex())
	r.HandleFunc("/editor", handleEditor()).Methods("GET")
	r.HandleFunc("/survey", handleSurvey()).Methods("GET")

	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}", handleSurveyDataPUT(db)).Methods("PUT")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataGET(db)).Methods("GET")
	r.HandleFunc("/surveys/{surveyUID}/answers/{answerUID}", handleAnswersDataPUT(db)).Methods("PUT")

	SurveyType = graphql.NewObject(graphql.ObjectConfig{
		Name: "SurveyType",
		Fields: graphql.Fields{
			"uid": &graphql.Field{
				Type: graphql.String,
			},
			"data": &graphql.Field{
				Type: graphql.String,
			},
			"answers": &graphql.Field{
				Type: graphql.NewList(AnswerType),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					survey, ok := params.Source.(Survey)
					if !ok {
						return nil, errors.New("Error with answers")
					}
					// TODO: Return all answers related with survey

					dbAnswers := []model.Answer{}
					err := db.Find(&dbAnswers).Error

					if err == gorm.ErrRecordNotFound {
						return nil, errors.New("answer not found")
					}

					var answers []Answer
					for _, v := range dbAnswers {
						answers = append(answers, Answer{UID: v.Uid, Data: v.Data, Survey: survey})
					}

					return answers, nil

				},
			},
		},
	})

	RootQuery := graphql.NewObject(
		graphql.ObjectConfig{
			Name:        "Query",
			Description: "Testing",
			Fields: graphql.Fields{

				"surveys": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(SurveyType))),
					Args: graphql.FieldConfigArgument{
						"uid": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
					Resolve: func(params graphql.ResolveParams) (interface{}, error) {
						uid, ok := params.Args["uid"].(string)
						if !ok {
							var surveys []Survey
							dbSurveys := []model.Survey{}

							db.Find(&dbSurveys)

							for _, v := range dbSurveys {
								surveys = append(surveys, Survey{UID: v.Uid, Data: v.Data})
							}
							return surveys, nil
						}

						survey := model.Survey{}
						err := db.Where("Uid = ?", uid).First(&survey).Error
						if err == gorm.ErrRecordNotFound {
							return nil, errors.New("survey not found")
						}
						return []Survey{Survey{UID: survey.Uid, Data: survey.Data}}, nil

					},
				},
			}})

	schema, _ := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: RootQuery,
		},
	)

	r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
