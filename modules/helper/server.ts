import { fetch } from "bun";
import { Database } from "../database";
import type { ServerStatus } from "../../common/context/status";

export async function reloadServers() {
  const db = new Database();
  const servers = await db.getServers();

  const serverFetchs = [];
  for (const server of servers) {
    serverFetchs.push(fetch(`https://${server.domain}/${process.env.SERVER_PASSWORD}`));
  }

  await Promise.all(serverFetchs);
}

export async function getServerStatus(domain: string) {
  const res = await fetch(`https://${domain}/status`);
  if (res.status == 200) {
    return (await res.json()) as ServerStatus;
  }
}
