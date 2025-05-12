FROM oven/bun:latest
 
WORKDIR /app
COPY . .
 
RUN bun install
RUN bun install pm2 -g
 
EXPOSE 8080

ENTRYPOINT [ "pm2-runtime", "--interpreter", "bun", "index.ts" ]