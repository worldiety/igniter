FROM golang:1.12.9-alpine3.10
RUN apk add git
ADD $PWD /code

RUN cd /code && go build . && mv igniter /bin/ && cd .. && rm -rf /code

ENTRYPOINT /bin/igniter
