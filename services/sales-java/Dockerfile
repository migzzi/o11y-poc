FROM openjdk:17-jdk-slim as builder

WORKDIR /app

COPY *.gradle.kts gradle ./
COPY gradle gradle

RUN ./gradlew build --no-daemon || return 0

COPY . .

RUN ./gradlew build --no-daemon


FROM openjdk:17-jdk-slim

WORKDIR /app

ARG APP_PORT=8080

EXPOSE $APP_PORT

RUN apt-get -y update; apt-get -y install curl

RUN curl -L -O https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar

COPY --from=builder /app/build/libs/sales-0.0.1.jar /app

ENV JAVA_TOOL_OPTIONS="-javaagent:./opentelemetry-javaagent.jar"

ENTRYPOINT ["java", "-jar", "/app/sales-0.0.1.jar"]
