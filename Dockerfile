FROM node:25-alpine AS builder-frontend
WORKDIR /app/frontend
COPY frontend/package*.json .
RUN npm install
COPY frontend/ .
RUN npx ng build --configuration production

FROM golang:1.25-alpine AS builder 
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder-frontend /app/frontend/dist/frontend frontend/dist/frontend
RUN go build -o monitor main.go

FROM alpine:latest
RUN apk add --no-cache libcap
WORKDIR /app
COPY --from=builder /app/monitor .
RUN setcap cap_net_raw+eip /app/monitor
ENTRYPOINT ["./monitor"]