FROM oven/bun:latest
 
WORKDIR /app
COPY . .
 
RUN bun install
RUN apt install pm2 -g
 
EXPOSE 8080

ENTRYPOINT [ "pm2-runtime", "--interpreter", "bun", "index.ts" ]