FROM golang AS builder

WORKDIR /kynes

ADD . .

RUN go build

FROM hashicorp/terraform:0.12.29

COPY --from=builder /kynes/kynes .

ENTRYPOINT [ "./kynes" ]
