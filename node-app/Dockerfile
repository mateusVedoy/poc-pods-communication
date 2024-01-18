#Run app 
FROM node:18-alpine

RUN apk add --no-cache tzdata && \
    cp -fv /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && \
    echo "America/Sao_Paulo" > /etc/timezone && \
    apk del tzdata

WORKDIR /app

COPY package*.json ./

RUN npm install --production

ENV COMMIT=@@commit

COPY . .

ADD k8s k8s

CMD ["node", "src/index.js"]
