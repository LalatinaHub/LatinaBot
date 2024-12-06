import { fetch, sleep } from "bun";
import { Database } from "../database";
import type { ServerStatus } from "../../common/context/status";

export async function reloadServers() {
  const db = new Database();
  const servers = await db.getServers();
  const serverFetchs = [];

  await sleep(2000);
  for (const server of servers) {
    serverFetchs.push(fetch(`https://${server.domain}/api/v1/${process.env.SERVER_PASSWORD}`));
  }

  await Promise.all(serverFetchs);
}

export async function getServerStatus(domain: string) {
  const res = await fetch(`https://${domain}/api/v1/status`);
  if (res.status == 200) {
    return (await res.json()) as ServerStatus;
  }
}

export async function getServerProxies(domain: string) {
  const res = await fetch(`https://${domain}/yacd/proxies`, {
    headers: {
      Authorization: "Bearer YACD_PASSWORD",
    },
  });

  if (res) {
    return await res.json();
  }
}
