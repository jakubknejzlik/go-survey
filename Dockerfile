FROM kontena/cli

COPY bin/go-survey-alpine /usr/local/bin/go-survey

ENTRYPOINT ["go-survey"]
