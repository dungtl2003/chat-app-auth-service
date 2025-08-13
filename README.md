# Auth service

Last updated: 2025-08-13

# Table of Contents

- [Description](#description)
- [Endpoints](#endpoints)
  - [GET /healthcheck](#get-healthcheck)
  - [POST /login](#post-login)
  - [GET /check](#get-check)
  - [GET /refresh](#get-refresh)
  - [GET /logout](#get-logout)
  - [POST /signup](#post-signup)
- [Configuration](#configuration)
- [Testing](#testing)
- [Docker](#docker)
- [Attention](#attention)

## Description

This service is responsible for authorizing users.

## Endpoints

### GET /healthcheck

- **Description**: Check the health status of the service.
- **Response**: JSON object with the health status.

### POST /login

- **Description**: Login a user and return an access token and refresh token.
- **Request Body**: JSON object with the following fields:
  - `identifier`: The identifier of the user (email or username)
  - `password`: The password of the user
  - `device_info`: The device information of the user (in the format of a JSON object)
- **Response**: JSON object with the following fields:
  - `access_token`: The access token of the user
  - `session_id`: The current session id of the user
  - `user`: The user object (with some fields removed for security reasons)
- **Cookies**: This endpoint will also set the `refresh_token` cookie in the response.

### GET /check

- **Description**: Authorize a user.
- **Request Header**: The request should contain the following fields:
  - `Authorization`: The access token of the user (Bearer token)
- **Response**: JSON object with the following fields:
    - `user`: The user object (with some fields removed for security reasons)

### GET /refresh

- **Description**: Refresh the access token and refresh token of the user.
- **Request Header**: The request should contain the following fields:
  - `Cookie`: The refresh token of the user. This should be set in the `refresh_token` cookie.
- **Response**: JSON object with the following fields:
    - `access_token`: The access token of the user
    - `user`: The user object (with some fields removed for security reasons)
    - `session_id`: The current session id of the user
- **Cookies**: This endpoint will also set the `refresh_token` cookie in the response.

### GET /logout

- **Description**: Logout a user and invalidate the session.
- **Request Header**: The request should contain the following fields:
  - `Authorization`: The access token of the user (Bearer token)
- **Response**: message indicating that the user has been logged out successfully.

### POST /signup

- **Description**: Sign up a new user and log them in.
- **Request Body**: JSON object with the following fields:
  - `email`: The email of the user
  - `username`: The username of the user
  - `password`: The password of the user
  - `device_info`: The device information of the user (in the format of a JSON object)
  - `role`: The role of the user
- **Response**: JSON object with the following fields:
  - `access_token`: The access token of the user
  - `session_id`: The current session id of the user
  - `user`: The user object (with some fields removed for security reasons)
- **Cookies**: This endpoint will also set the `refresh_token` cookie in the response.

## Configuration

The service can be configured using the following environment variables:

| Name | Description | Required | Default | Type | Possible Values |
|------|-------------|----------|---------|------|-----------------|
| PORT | The port to run the service | No | 8400 | int | any valid port number |
| LOG_LEVEL | The log level of the service | No | INFO | string | DEBUG, INFO, WARN, ERROR |
| LOG_KIND | The kind of log to output | No | TEXT | string | TEXT, JSON |
| ENVIRONMENT | The environment to run the service | No | dev | string | any valid string |
| USER_SERVICE_URL | The URL of the user service endpoint | Yes | | string | any valid URL |
| JWT_SECRET | The secret key to sign the JWT tokens | Yes | | string | any valid string |
| ACCESS_TOKEN_DURATION_MS | The duration of the access token in milliseconds | No | 900000 (15 minutes) | int | any valid unsigned int |
| REFRESH_TOKEN_DURATION_MS | The duration of the refresh token in milliseconds | No | 172800000 (2 days) | int | any valid unsigned int |
| PASSWORD_HASH_COST | The cost of the password hashing | No | 12 | int | any valid integer between 4 and 31 |
| DOMAIN_NAME | The domain name of the service | No | localhost | string | any valid string |
| ID_GENERATOR_ADDR | The address of the id generator service | Yes | | string | any valid string made of address and port (e.g. localhost:8501) |
| ID_GENERATOR_CERT_DIR | The directory to the certificate of the id generator service | No | | string | any valid directory |

You can see the full configuration example in `./template/env-template` file.

## Testing

Run the tests with the following command:

```bash
make test
```

Or run the tests with analysis enabled:

```bash
make test JSON=1
```

## Docker

Use this command to build the Docker image:

```bash
make ci_%env # e.g. make ci_dev
```

There are scripts that support logging services and result to file. If in the future, you want to add a new service log, you need to update `LOG_META` variable in `./scripts/test_local.sh` script.

## Attention

2025-08-13: `__test_with_services` script container healthcheck sometimes fails due to the fact that the service is not ready yet but for some reason, docker still said that the container is healthy. This will be fixed in the future.
