FROM golang:1.14 as build
RUN mkdir /action
COPY . /app/
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/action .

RUN apt-get update && apt-get -y install upx

RUN upx -q -9 /bin/action

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/action /bin/action

# Specify the container's entrypoint as the action
ENTRYPOINT ["/bin/action"]