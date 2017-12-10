package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
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

	r.HandleFunc("/properties.json", handleProperties()).Methods("GET")

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
		if err := validateToken(r); err != nil {
			fmt.Fprint(w, "invalid access token")
			return
		}
		data, _ := ioutil.ReadFile("views/editor.html")
		w.Write(data)
	}
}

func handleProperties() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		propertiesUrl := os.Getenv("PROPERTIES_URL")
		if propertiesUrl == "" {
			fmt.Fprint(w, "[]")
		} else {

			client := &http.Client{}

			req, err := http.NewRequest("GET", propertiesUrl, nil)
			if err != nil {
				fmt.Fprintf(w, "console.log(\"invalid url: %s\")", propertiesUrl)
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", getToken(r)))

			fmt.Println(req.Header.Get("Authorization"))
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(w, "console.log(\"error: %s\")", err.Error())
				return
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(w, "console.log(\"error: %s\")", err.Error())
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
	}
}

func handleSurvey() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := validateToken(r); err != nil {
			fmt.Fprint(w, "invalid access token")
			return
		}
		data, _ := ioutil.ReadFile("views/survey.html")
		w.Write(data)
	}
}

func handleSurveyDataGET(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := validateToken(r); err != nil {
			fmt.Fprint(w, "invalid access token")
			return
		}
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

		if err := validateToken(r); err != nil {
			fmt.Fprint(w, "invalid access token")
			return
		}

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

		if err := validateToken(r); err != nil {
			fmt.Fprint(w, "invalid access token")
			return
		}

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

		answerWebhookURL := os.Getenv("ANSWER_WEBHOOK_URL")
		if answerWebhookURL != "" {
			data = []byte(fmt.Sprintf("{\"survey\":\"%s\",\"asnwer\":\"%s\"}", surveyUID, answerUID))
			reader := bytes.NewReader(data)
			http.Post(answerWebhookURL, "application/json", reader)
		}
	}
}

func getToken(r *http.Request) string {
	tokenString := r.URL.Query().Get("access_token") //"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU"
	if tokenString == "" {
		tokenString = r.Header.Get("Authorization")
	}

	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	return tokenString
}

func validateToken(r *http.Request) error {
	tokenString := getToken(r)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "JWT_SECRET"
	}
	if tokenString == "" {
		return errors.New("missing token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}
