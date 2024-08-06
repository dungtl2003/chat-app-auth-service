FROM eclipse-temurin:21-jdk-alpine@sha256:ebfc28d35b192c55509e3c7cc597d91136528f1a9d3261965b44663af9eb4b4b AS compile
WORKDIR /project
COPY . .
RUN ./mvnw clean package -DskipTests

FROM eclipse-temurin:21-jdk-alpine@sha256:ebfc28d35b192c55509e3c7cc597d91136528f1a9d3261965b44663af9eb4b4b AS builder
ARG JAR_FILE=target/*.jar
COPY --from=compile /project/${JAR_FILE} application.jar
RUN java -Djarmode=layertools -jar application.jar extract

FROM eclipse-temurin:21-jdk-alpine@sha256:ebfc28d35b192c55509e3c7cc597d91136528f1a9d3261965b44663af9eb4b4b AS final
LABEL authors="ilikeblue"
WORKDIR /app
RUN addgroup --system javauser && \
    adduser --system --shell /bin/false --ingroup javauser javauser
COPY --from=builder dependencies/ ./
COPY --from=builder snapshot-dependencies/ ./
COPY --from=builder spring-boot-loader/ ./
COPY --from=builder application/ ./
RUN chown -R javauser:javauser /app
USER javauser

EXPOSE 80
ENTRYPOINT ["java", "--enable-preview", "org.springframework.boot.loader.launch.JarLauncher"]
