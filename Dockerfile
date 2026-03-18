FROM node:20-slim AS frontend
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM golang:1.25.4-alpine AS backend
WORKDIR /src

COPY bck/go.mod bck/go.sum ./bck/
WORKDIR /src/bck
RUN go mod download

WORKDIR /src
COPY bck ./bck
COPY --from=frontend /app/bck/internal/ui/dist ./bck/internal/ui/dist

WORKDIR /src/bck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -tags prod -o /out/timetable ./cmd

FROM alpine:3.21 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /out/timetable /timetable

ENV GO_ENV=production
ENV PORT=8080

EXPOSE 8080
USER 10001:10001
ENTRYPOINT ["/timetable"]
