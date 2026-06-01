FROM postgres:15-bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    redis-server \
    supervisor \
    gcc \
    libsqlite3-dev \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

ENV GO_VERSION=1.23.1
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/root/go
ENV PATH=$PATH:$GOPATH/bin

WORKDIR /app

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres123
ENV POSTGRES_DB=distribution
ENV PGDATA=/var/lib/postgresql/data

ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=postgres123
ENV DB_NAME=distribution
ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379
ENV REDIS_PASSWORD=
ENV JWT_SECRET=your_jwt_secret_key_here
ENV JWT_EXPIRE_HOURS=24
ENV COMMISSION_LEVEL1_RATE=10
ENV COMMISSION_LEVEL2_RATE=5
ENV COMMISSION_LEVEL3_RATE=3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

RUN mkdir -p /var/log/supervisor /var/run/postgresql /var/log/postgresql \
    && chown -R postgres:postgres /var/run/postgresql /var/log/postgresql \
    && chmod +x /app/wait-for-services.sh

COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN git init \
    && git config user.email "docker@example.com" \
    && git config user.name "Docker" \
    && git add -A \
    && git commit -m "Initial commit"

EXPOSE 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
