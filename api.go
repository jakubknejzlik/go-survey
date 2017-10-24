package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
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
	// Demo data for testing
	survey1 := Survey{UID: "j", Data: "Survey1"}
	survey2 := Survey{UID: "p", Data: "Survey2"}
	survey3 := Survey{UID: "k", Data: "Survey3"}
	SurveyList = append(SurveyList, survey1, survey2, survey3)

	ans1 := Answer{UID: "j", Data: "Question1"}
	ans2 := Answer{UID: "p", Data: "Question2"}
	ans3 := Answer{UID: "k", Data: "Question3"}
	ans4 := Answer{UID: "k", Data: "Question4"}
	ans5 := Answer{UID: "k", Data: "Question5"}
	AnswerList = append(AnswerList, ans1, ans2, ans3, ans4, ans5)

	var SurveyType *graphql.Object

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
					var tmp []Answer
					for _, v := range AnswerList {
						if v.UID == survey.UID {
							v.Data = "Hello Bro"
							tmp = append(tmp, v)
						}
					}
					return tmp, nil

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
							return SurveyList, nil
						}
						for _, v := range SurveyList {
							if v.UID == uid {
								return []Survey{v}, nil
							}
						}
						return nil, errors.New("Incrorrect survey uid!")
					},
				},
			}})

	schema, _ := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: RootQuery,
		},
	)

	r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		json.NewEncoder(w).Encode(result)
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}
