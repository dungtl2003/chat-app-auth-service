FROM eclipse-temurin:22.0.1_8-jre AS builder
WORKDIR extracted
ADD target/*.jar app.jar

FROM eclipse-temurin:22.0.1_8-jre
WORKDIR application
COPY --from=builder extracted/dependencies/ ./
COPY --from=builder extracted/spring-boot-loader/ ./
COPY --from=builder extracted/snapshot-dependencies/ ./
COPY --from=builder extracted/application/ ./

EXPOSE 8080
ENTRYPOINT ["java", "org.springframework.boot.loader.JarLauncher"]