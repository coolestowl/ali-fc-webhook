FROM alpine:latest

COPY ./dist/main /main

EXPOSE 8000

CMD [ "/main" ]
