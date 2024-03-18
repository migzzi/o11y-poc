FROM --platform=arm64 node:18.3.0-alpine3.15

# Create app directory
WORKDIR /usr/src/app

# Install app dependencies
COPY package*.json ./

RUN npm install

ARG NODE_ENV=production
ARG PORT=8080

# Bundle app source
COPY . .

EXPOSE $PORT

CMD [ "npm", "start"]