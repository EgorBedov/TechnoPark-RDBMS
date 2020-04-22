FROM golang:1.13-stretch AS build

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone


WORKDIR /home/egogoger-rdbms
COPY . .
RUN go build -o /bin/egogoger-rdbms ./cmd/server/main.go

# Until this all okay

FROM ubuntu:18.04 AS release

MAINTAINER Egor A. Bedov

# Set the timezone
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get -y update
# RUN apt-get -y upgrade
RUN apt install -y git wget gcc gnupg

ENV PGVER 11

RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main" > /etc/apt/sources.list.d/pgdg.list

RUN wget https://www.postgresql.org/media/keys/ACCC4CF8.asc
RUN apt-key add ACCC4CF8.asc

RUN apt-get update

RUN apt-get install -y  postgresql-$PGVER


WORKDIR /home/egogoger-rdbms
COPY . .

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f /home/egogoger-rdbms/scripts/init.sql &&\
    /etc/init.d/postgresql stop

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

# https://habr.com/ru/post/444018/
RUN echo "random_page_cost = 1.0" >> /etc/postgresql/$PGVER/main/postgresql.conf

# http://www.gilev.ru/postgres-%D0%BF%D0%B0%D1%80%D0%B0%D0%BC%D0%B5%D1%82%D1%80-synchronous_commit/
RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf

RUN echo "max_connections = 100" >> /etc/postgresql/$PGVER/main/postgresql.conf

# Expose the PostgreSQL port
# EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

# Объявлем порт сервера
EXPOSE 5000

# Собранный ранее сервер
COPY --from=build /bin/egogoger-rdbms /bin/egogoger-rdbms

#
# Запускаем PostgreSQL и сервер
#
CMD service postgresql start && ./bin/egogoger-rdbms