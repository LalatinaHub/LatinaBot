FROM oven/bun:latest
 
WORKDIR /app
COPY . .
 
RUN bun install
 
EXPOSE 8080

ENTRYPOINT [ "bun", "run", "index.ts" ]