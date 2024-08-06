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

you need to have `.env` file in root folder, in the file you
need `key=value` each line. See list of required environment
variables [here](#-list-of-available-environment-variables):<br>

## ⇁ List of available environment variables

# Environment variables

| Variable                | Required | Purpose                                                                                                                   |
|-------------------------|----------|---------------------------------------------------------------------------------------------------------------------------|
| SPRING_PROFILES_ACTIVE  | YES      | your environment, currently it can be `dev` or `prod`                                                                     |
| DATABASE                | YES      | your chosen database. For example: `postgresql`                                                                           |
| DRIVER_CLASS_NAME       | YES      | example for postgresql: `org.postgresql.Driver`                                                                           |                                                                          
| DB_DRIVER               | YES      | for example, postgresql driver will be `jdbc:postgresql:/`                                                                |
| DB_HOST                 | YES      | host name                                                                                                                 |
| DB_NAME                 | YES      | database schema name                                                                                                      |
| DB_PORT                 | YES      | for example, postgresql will be `5432`                                                                                    |
| DB_ROOT_CERT            | YES      | for example, linux will be `/etc/ssl/certs/ca-certificates.crt`                                                           |
| USERNAME                | YES      | admin username (for whole system control)                                                                                 |
| PASSWORD                | YES      | admin password. Notice that the password must be encrypted by bcrypt with correct configuration, not raw password         |
| PORT                    | YES      | server's port. Note that it must match container port if you want to use docker compose to run your app                   |
| PROTOCOL                | YES      | can be `http` or `https`. In production, you need to set `http` if your port is `80`, `https` if your port is `443`       |
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
| LOG_PATH                | NO       | path to log folder                                                                                                        |

For the full .env file example, check
out [this template](./templates/.env.template)

## ⇁ Getting Started

if you change the code, then run this first to format the code:

```shell
make clean
```

first, you need to have `.env` file inside root folder. See more
in [here](#-list-of-available-environment-variables)<br>

you can run the development server by this command:

```shell
make run # (only if you have set `MAVEN_OPTS` in `.env` file. See more in [here](#-list-of-available-environment-variables))
```
NOTE: if your `.env` file has different name, add `ENV_FILE` option in the command (this can work to all `make` command). For example:
```shell
make run ENV_FILE=.env.dev
```

if you want to build docker image, use:
```shell
make build_% # for example: make `build_dev` or make `build_prod`
```

then you can run docker with:
```shell
make server_%
```

after you run the app, you can go to `/swagger-ui/index.html#/` endpoint to see
swagger (this can only work if `SPRING_PROFILES_ACTIVE=dev`)

## ⇁ Run tests

first, you need to have `.env` file inside `environments/dev` folder. See more
in [here](#-list-of-available-environment-variables)<br>

### ⇁ Run all tests

```shell
make test
```

### ⇁ Run test of 1 method in the class

you can do this by running `./mvnw_wrapper.sh test -Dtest=class#method`. For
example:

```shell
./mvnw_wrapper.sh test -Dtest=ChatAppAuthServiceApplicationTests#testLogoutRequestWithValidRefreshTokenShouldGet200OkAndCannotReuseRT
```

## ⇁ Deploy

comming soon
