# Auth service

# Table of Contents

- [Description](#description)
- [Configuration](#configuration)
- [Testing](#testing)
- [Docker](#docker)
- [Attention](#attention)

## Description

This service is responsible for authorizing users.

## Configuration

The service can be configured using the following environment variables:

| Name | Description | Required | Default | Type | Possible Values |
|------|-------------|----------|---------|------|-----------------|
| PORT | The port to run the service | No | 8400 | int | any valid port number |
| LOG_LEVEL | The log level of the service | No | INFO | string | DEBUG, INFO, WARN, ERROR |
| LOG_KIND | The kind of log to output | No | TEXT | string | TEXT, JSON |
| ENVIRONMENT | The environment to run the service | No | dev | string | dev, prod, test |
| USER_SERVICE_URL | The URL of the user service endpoint | Yes | | string | any valid URL |
| JWT_SECRET | The secret key to sign the JWT tokens | Yes | | string | any valid string |
| ACCESS_TOKEN_DURATION_MS | The duration of the access token in milliseconds | No | 900000 (15 minutes) | int | any valid unsigned int |
| REFRESH_TOKEN_DURATION_MS | The duration of the refresh token in milliseconds | No | 172800000 (2 days) | int | any valid unsigned int |
| PASSWORD_HASH_COST | The cost of the password hashing | No | 12 | int | any valid integer between 4 and 31 |
| DOMAIN_NAME | The domain name of the service | No | localhost | string | any valid string |
| TOKEN_ISSUER | The issuer of the JWT tokens | No | auth-service | string | any valid string |
| TOKEN_INTERNAL_AUDIENCE | The internal audience of the JWT tokens | No | auth-service-internal | string | any valid string |
| TOKEN_FRONTEND_AUDIENCE | The frontend audience of the JWT tokens | No | auth-service-frontend | string | any valid string |
| ID_GENERATOR_SERVICE_ADDR | The address of the id generator service | Yes | | string | any valid string made of address and port (e.g. localhost:8501) |
| ID_GENERATOR_SERVICE_CERT_DIR | The directory to the certificate of the id generator service | No | | string | any valid directory |
| ID_GENERATOR_SERVICE_EPOCH | The epoch to use for the id generator service | No | 1672531200000 | int | any valid epoch in milliseconds |
| REDIS_MASTER_ADDR | The master address of the Redis server | Yes | | string | any valid address made of host and port (e.g. localhost:6379) |
| REDIS_REPLICA_ADDR | The replica address of the Redis server, default to the master address if not provided | No | | string | any valid address made of host and port (e.g. localhost:6379) |
| REDIS_PASSWORD | The password for the Redis server | No | | string | any valid string |
| PASSWORD_RESET_RATE_LIMIT_TTL | The TTL for the password reset rate limit | No | 1 hour | string | any valid duration. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (e.g. "300ms", "1.5h" or "2h45m") |
| PASSWORD_RESET_RATE_LIMIT_MAX | The maximum number of password reset requests allowed in the TTL | No | 5 | int | any valid unsigned int | 
| PASSWORD_RESET_CODE_TTL | The TTL for the password reset code | No | 15 minutes | string | any valid duration. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (e.g. "300ms", "1.5h" or "2h45m") |
| PASSWORD_RESET_ATTEMPT_TTL | The TTL for the password reset attempt counter | No | 1 hour | string | any valid duration. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (e.g. "300ms", "1.5h" or "2h45m") |
| PASSWORD_RESET_ATTEMPT_MAX | The maximum number of password reset attempts allowed in the TTL | No | 5 | int | any valid unsigned int |
| PASSWORD_RESET_TOKEN_TTL | The TTL for the password reset token | No | 15 minutes | string | any valid duration. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" (e.g. "300ms", "1.5h" or "2h45m") |
| SMTP_EMAIL_FROM_ADDRESS | The email address for the email sender | Yes | | string | any valid email address |
| SMTP_EMAIL_FROM_NAME_DISPLAY | The display name for the email sender | No | Same as SMTP_EMAIL_FROM_ADDRESS | string | any valid string |
| SMTP_HOST | The SMTP email host | Yes | | string | any valid string |
| SMTP_PORT | The SMTP email port | Yes | | int | any valid port number |

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
