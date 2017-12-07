# go-survey

Service build using SurveyJS library. Create and edit

[![Build Status](https://travis-ci.org/jakubknejzlik/go-survey.svg?branch=master)](https://travis-ci.org/jakubknejzlik/go-survey)

# Envorinment variables

* `DATABASE_URL` - database url where survey and answers are stored (default:
  sqlite3://:memory:)
* `PORT` - port on which app listens (default: 80)
* `PROPERTIES_URL` - url to load properties from. See below for example
* `ANSWER_WEBHOOK_URL` - url which is called when answer is created/updated
  (payload is in format {"survey":"XXX","answer":"YYY"})

## Properties

Content has following structure:

```
[
  {
    key: "categories",
    type: "multiplevalues",
    choices: [
        { value: "category_a", text: "Category A" },
        { value: "category_b", text: "Category B" }
    ]
  }
]
```

# Endpoints

This service doesn't handle user authorization for each endpoints, instead you
should access for each endpoint using your proxy and this service should not be
publicly accessible

* `/editor?survey=XXX` - editing detail for survey XXX
* `/survey?survey=XXX&answer=YYY` - answering survey XXX with answers grouped
  under id YYY
* `/graphql` - graphql api where you can scrape data based on
  `ANSWER_WEBHOOK_URL` webhook
