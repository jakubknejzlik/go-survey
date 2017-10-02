FROM ubuntu

RUN apk --update add sqlite

COPY bin/go-survey-alpine /usr/local/bin/go-survey

ENTRYPOINT ["go-survey"]
