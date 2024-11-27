import { fetch } from "bun";
import { Database } from "../database";

export async function reloadServers() {
  const db = new Database();
  const servers = await db.getServers();

  for (const server of servers) {
    await fetch(`https://${server.domain}/${process.env.SERVER_PASSWORD}`);
  }
}
