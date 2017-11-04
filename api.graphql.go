package main

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/jakubknejzlik/go-survey/model"
	"github.com/jinzhu/gorm"
)

func getSchema(db *gorm.DB) graphql.Schema {
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

					sur := model.Survey{Uid: survey.UID, Data: survey.Data}

					err := db.Model(&sur).Related(&sur.Answers, "SurveyUid").Error
					if err == gorm.ErrRecordNotFound {
						return nil, errors.New("answer not found")
					}

					var answers []Answer
					for _, v := range sur.Answers {
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

	return schema
}
