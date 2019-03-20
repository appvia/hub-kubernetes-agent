FROM alpine:3.8
LABEL Name=hub-kubernetes-agent \
      Release=https://github.com/appvia/hub-kubernetes-agent \
      Maintainer=danielwhatmuff@gmail.com \
      Url=https://github.com/appvia/hub-kubernetes-agent \
      Help=https://github.com/appvia/hub-kubernetes-agent/issues

RUN apk add --no-cache ca-certificates curl

ADD bin/hub-kubernetes-agent /hub-kubernetes-agent

USER 65534

ENTRYPOINT [ "/hub-kubernetes-agent" ]
