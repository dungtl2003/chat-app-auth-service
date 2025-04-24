# Auth service

## Table of Contents

- [Description](#description)
- [Endpoints](#endpoints)
  - [GET /healthcheck](#get-healthcheck)
  - [POST /login](#post-login)
  - [GET /check](#get-check)
  - [GET /refresh](#get-refresh)
  - [GET /logout](#get-logout)
  - [POST /signup](#post-signup)
- [Configuration](#configuration)
- [Running](#running)
- [Dockerize](#dockerize)
- [Testing](#testing)

## Description

This service is responsible for authorizing users.

## Endpoints

### GET /healthcheck

Check if the service is up and running. Always returns 200.

### POST /login

Login a user. The request body should contain the following fields:

- `identifier`: The identifier of the user (email or username)
- `password`: The password of the user
- `device_id`: The device id of the user

The response will contain the following fields:

- `access_token`: The access token of the user
- `user`: The user object (with some fields removed for security reasons)

This endpoint will also set the `refresh_token` cookie in the response.

### GET /check

Authorize a user. The request header should contain the following fields:

- `Authorization`: The access token of the user (Bearer token)

The response will contain the following fields:

- `user`: The user object (with some fields removed for security reasons)

### GET /refresh

Refresh the access token and refresh token of the user. The request header should contain the following fields:

- `Cookie`: The refresh token of the user. This should be set in the `refresh_token` cookie.

The response will contain the following fields:

- `access_token`: The access token of the user
- `user`: The user object (with some fields removed for security reasons)

This endpoint will also set the `refresh_token` cookie in the response.

### GET /logout

Logout a user. The request header should contain the following fields:

- `Authorization`: The access token of the user (Bearer token)

### POST /signup

Sign up and also login a user. The request body should contain the required fields
for creating a user (check the user service for the required fields).

The response will contain the following fields:

- `access_token`: The access token of the user
- `user`: The user object (with some fields removed for security reasons)

This endpoint will also set the `refresh_token` cookie in the response.

## Configuration

The service can be configured using the following environment variables:

| Name | Description | Required | Default | Type | Possible Values |
|------|-------------|----------|---------|------|-----------------|
| PORT | The port to run the service | No | 8400 | int | any valid port number |
| LOG_LEVEL | The log level of the service | No | INFO | string | DEBUG, INFO, WARN, ERROR |
| LOG_KIND | The kind of log to output | No | TEXT | string | TEXT, JSON |
| ENV | The environment to run the service | No | dev | string | any valid string |
| USER_SERVICE_URL | The URL of the user service endpoint | Yes | | string | any valid URL |
| JWT_SECRET | The secret key to sign the JWT tokens | Yes | | string | any valid string |
| ACCESS_TOKEN_DURATION_MS | The duration of the access token in milliseconds | No | 900000 (15 minutes) | int | any valid unsigned int |
| REFRESH_TOKEN_DURATION_MS | The duration of the refresh token in milliseconds | No | 172800000 (2 days) | int | any valid unsigned int |
| DOMAIN_NAME | The domain name of the service | No | localhost | string | any valid string |
| COST | The cost of the password bcrypt hashing algorithm | No | 12 | int | any valid unsigned int |

## Running

First, run the following command to generate certificates:

``` bash
./scripts/gen_certs.sh
```

Then, run the following command to run the service:

``` bash
make run # or make run ENV=? for a specific environment
```

If you want to run multiple dependencies (e.g. database, id generator), you can run the following command:

``` bash
make run_with_services # or make run_with_services ENV=? for a specific environment
```

Make sure to have the coresponding compose file in the `compose` directory and correct `env.%ENV%` file in the `environments` directory (details on environment variables can be found in the [Configuration](#configuration) section).

## Dockerize

There are many useful commands in makefile (e.g. `make dbuild_%ENV%`, `make dpush_%ENV%`, `make drun_%ENV%`, etc.) that can be used to build, push and run the service in Docker. You can also use `make ci_%ENV%` to build, push, release the service on Docker Hub.

## Testing

First, run the following command to generate certificates:

``` bash
./scripts/gen_certs.sh
```

Then, run the following command to run the tests:

``` bash
make test # or make test JSON=1 for json output (this can be used to get statistics as well)
```
