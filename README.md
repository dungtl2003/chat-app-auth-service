# Read Me First

PLEASE DO NOT USE "EXTENDS" KEYWORD

The following was discovered as part of building this project:

- The original package name 'org.service.auth.chat-app-authentication-service' is invalid and this project uses '
  org.service.auth.chatappauthservice' instead.

---

# Getting Started

### Reference Documentation

For further reference, please consider the following sections:

- [Official Apache Maven documentation](https://maven.apache.org/guides/index.html)
- [Spring Boot Maven Plugin Reference Guide](https://docs.spring.io/spring-boot/docs/3.2.3/maven-plugin/reference/html/)
- [Create an OCI image](https://docs.spring.io/spring-boot/docs/3.2.3/maven-plugin/reference/html/#build-image)
- [Spring Web](https://docs.spring.io/spring-boot/docs/3.2.3/reference/htmlsingle/index.html#web)

### Guides

The following guides illustrate how to use some features concretely:

- [Building a RESTful Web Service](https://spring.io/guides/gs/rest-service/)
- [Serving Web Content with Spring MVC](https://spring.io/guides/gs/serving-web-content/)
- [Building REST services with Spring](https://spring.io/guides/tutorials/rest/)

### Requirements

- jdk 21
- npx

### Config git hooks

```shell
git config core.hooksPath '.git-hooks'
```

Verify right hook directory:

```shell
git rev-parse --git-path hooks
```

---

# Swagger

- PORT: port number when app is runnning

- http://localhost:{PORT}/swagger-ui/index.html#/

---

# Environment variables

In order to run the project using the below command, you have to have `.env` file in your root directory.<br>
File must contain:
`ACCESS_JWT_LIFESPAN_MS`, `ACCESS_JWT_SECRET`, `BCRYPT_STRENGTH`, `DB_DRIVER`, `DB_HOST`, `DB_NAME`, `DB_PASSWORD`, `DB_PORT`, `DB_ROOT_CERT`, `DB_USER`, `PORT`, `REFRESH_JWT_LIFESPAN_MS`, `REFRESH_JWT_SECRET`.<br>
File format must be: `KEY=VALUE`, one line for each pair.

---

# Commands

### Run format

```shell
./mvnw_wrapper.sh spring-javaformat:apply
```

### Run tests

```shell
./mvnw_wrapper.sh clean verify
```
