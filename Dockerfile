FROM golang AS builder

WORKDIR /kynes

ADD ./src .

RUN go build

FROM hashicorp/terraform:0.12.29

COPY --from=builder /kynes/kynes .

ENTRYPOINT [ "./kynes" ]
