# Read Me First

PLEASE DO NOT USE "EXTENDS" KEYWORD

The following was discovered as part of building this project:

* The original package name 'org.service.auth.chat-app-authentication-service'
  is invalid and this project uses '
  org.service.auth.chatappauthservice' instead.

---

# About project

this is a chat app's service that supports authentication and authorization

# Table of content

[prerequisites](#-prerequisites)<br>
[setup](#-setup)<br>
[getting started](#-getting-started)<br>
[run test](#-run-test)<br>
[database schema](#-database-schema)<br>
[deployment (comming soon)](#-deploy)<br>

## ⇁ Prerequisites

you must have npm installed<br>
jdk's required version: 21<br>
database of your choice<br>

## ⇁ Setup

first, clone this project<br>
next, config git hooks<br>

```shell
git config core.hooksPath '.git-hooks'
```

verify right hook directory:

```shell
git rev-parse --git-path hooks
```

you need to have `.env` file in root project, in the file you need `key=value`
each line. See list of required
environment variables [here](#-list-of-available-environment-variables):<br>

## ⇁ List of available environment variables

# Environment variables

| Variable                | Required | Purpose                                                                                                                   |
|-------------------------|----------|---------------------------------------------------------------------------------------------------------------------------|
| ENVIRONMENT             | NO       | can be `development` or `production`. Default: `development`                                                              |
| DATABASE                | YES      | your chosen database. For example: `postgresql`                                                                           |
| DRIVER_CLASS_NAME       | YES      | example for postgresql: `org.postgresql.Driver`                                                                           |                                                                          
| DB_DRIVER               | YES      | for example, postgresql driver will be `jdbc:postgresql:/`                                                                |
| DB_HOST                 | YES      | host name                                                                                                                 |
| DB_NAME                 | YES      | database schema name                                                                                                      |
| DB_PORT                 | YES      | for example, postgresql will be `5432`                                                                                    |
| DB_ROOT_CERT            | YES      | for example, linux will be `/etc/ssl/certs/ca-certificates.crt`                                                           |
| USERNAME                | YES      | admin username (for whole system control)                                                                                 |
| PASSWORD                | YES      | admin password. Notice that the password must be encrypted by bcrypt with correct configuration, not raw password         |
| PORT                    | YES      | server's port                                                                                                             |
| PROTOCOL                | YES      | can be `http` or `https`                                                                                                  |
| DOMAIN                  | YES      | your domain name                                                                                                          |
| ACCESS_JWT_LIFESPAN_MS  | NO       | duration of access token in millisecond. Default: `600000` ms                                                             |
| REFRESH_JWT_LIFESPAN_MS | NO       | duration of refresh token in millisecond. Default: `86400000` ms                                                          |
| DB_USER                 | YES      | database user                                                                                                             |
| DB_PASSWORD             | YES      | password of database user                                                                                                 |
| REFRESH_JWT_SECRET      | YES      | secret to en/decrypt refresh token                                                                                        |
| ACCESS_JWT_SECRET       | YES      | secret to en/decrypt access token                                                                                         |
| BCRYPT_STRENGTH         | NO       | this service will en/decrypt token using `bcrypt`, you can define how strong you want this algorithm to be. Default: `12` |
| API_VERSION             | YES      | API version. For example: `v1`                                                                                            |                                                                                              
| MAVEN_OPTS              | NO       | set this to `--enable-preview` to run on terminal                                                                         |

## ⇁ Getting Started

run the development server:

```shell
./mvnw_wrapper.sh exec:java # (only if you have set `MAVEN_OPTS` in `.env` file. See more in [here](#-list-of-available-environment-variables))
```

or you can run this command if you have docker compose:

```shell
docker compose up
```

and turn the service down with:

```shell
docker compose down
```

if you change the code, then run this first to format the code:

```shell
./mvnw_wrapper.sh spring-javaformat:apply
```

after you run the app, you can go to `/swagger-ui/index.html#/` endpoint to see
swagger

---

## ⇁ Run test

```shell
./mvnw_wrapper.sh clean verify
```

## ⇁ Database schema

![Schema](./assets/db_schema.png)

## ⇁ Deploy