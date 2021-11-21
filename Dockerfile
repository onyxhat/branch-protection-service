FROM alpine:3.15
ARG DOCKER_BIN

LABEL MAINTAINER="onyxhat"
LABEL REPO="https://github.com/onyxhat/branch-protection-service"
LABEL FORKED_FROM="https://github.com/So-Sahari/branch-protection-service"

ENV TOKEN ORG

RUN apt update && apt upgrade -y

COPY ./bin/${DOCKER_BIN} /app/branch-protection-service
RUN chmod -R +x /app

CMD [ "/app/branch-protection-service", "-token", ${TOKEN}, "-org", ${ORG} ]
