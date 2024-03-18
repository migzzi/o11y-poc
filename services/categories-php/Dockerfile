FROM --platform=arm64 php:8.3

ARG PORT=80

# Create app directory
WORKDIR /var/www/html

RUN apt-get update && apt-get install -y \
    libzip-dev \
    zip \
    make \
    autoconf \
    && docker-php-ext-configure zip \
    && docker-php-ext-install zip

RUN pecl install opentelemetry \
    && docker-php-ext-enable opentelemetry

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer


# Install app dependencies
COPY . .

ENV APP_PORT=8080
EXPOSE $APP_PORT

CMD php -S 0.0.0.0:${APP_PORT}
