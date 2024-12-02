import { fetch } from "bun";
import { Database } from "../database";

export async function reloadServers() {
  const db = new Database();
  const servers = await db.getServers();

  const serverFetchs = [];
  for (const server of servers) {
    serverFetchs.push(fetch(`http://${server.domain}/${process.env.SERVER_PASSWORD}`));
  }

  await Promise.all(serverFetchs);
}
