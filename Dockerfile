FROM golang:alpine AS builder

WORKDIR /kynes

ADD . .

RUN go build

FROM hashicorp/terraform:0.12.29

WORKDIR /kynes

COPY --from=builder /kynes/kynes .

ENTRYPOINT [ "/kynes/kynes" ]