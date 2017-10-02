FROM golang:1.8.3-onbuild

RUN apt-get update && apt-get install sqlite3
# COPY bin/go-survey-alpine /usr/local/bin/go-survey

ENTRYPOINT ["go-survey"]
