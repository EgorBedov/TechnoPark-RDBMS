FROM golang:1.13-stretch AS build

WORKDIR /home/egogoger-rdbms
COPY . .
RUN go build -o /bin/egogoger-rdbms ./cmd/server/main.go


FROM ubuntu:18.04
MAINTAINER Egor A. Bedov
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y gnupg
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y git

USER root
WORKDIR /home/egogoger-rdbms
RUN cd /home/egogoger-rdbms
COPY . .

RUN apt-get -y update
RUN apt-get -y install apt-transport-https git wget
RUN echo 'deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main' >> /etc/apt/sources.list.d/pgdg.list
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN apt-get -y update
ENV PGVERSION 12
RUN apt-get -y install postgresql-$PGVERSION postgresql-contrib



USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f /home/egogoger-rdbms/scripts/init.sql &&\
    /etc/init.d/postgresql stop

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

# https://habr.com/ru/post/444018/
RUN echo "random_page_cost = 1.0" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

# http://www.gilev.ru/postgres-%D0%BF%D0%B0%D1%80%D0%B0%D0%BC%D0%B5%D1%82%D1%80-synchronous_commit/
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

RUN echo "max_connections = 100" >> /etc/postgresql/$PGVERSION/main/postgresql.conf

# Expose the PostgreSQL port
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
EXPOSE 5432
EXPOSE 5000


# Back to the root user
USER root
# Собранный ранее сервер
COPY --from=build /bin/egogoger-rdbms /bin/egogoger-rdbms

#
# Запускаем PostgreSQL и сервер
#
CMD service postgresql start && cd / && ./bin/egogoger-rdbms
