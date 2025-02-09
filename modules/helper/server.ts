import { fetch, sleep } from "bun";
import { Database } from "../database";
import type { ServerStatus } from "../../common/context/status";

const db = new Database();

export async function reloadServers() {
  const servers = await db.getServers();
  const apiToken = await db.getKV("apiToken");
  const serverFetchs = [];

  await sleep(2000);
  for (const server of servers) {
    serverFetchs.push(fetch(`http://${server.domain}/api/v1/${apiToken.value}`));
    await sleep(200);
  }

  await Promise.all(serverFetchs);
}

export async function getServerStatus(domain: string) {
  const res = await fetch(`http://${domain}/api/v1/status`);
  if (res.status == 200) {
    return (await res.json()) as ServerStatus;
  }
}

export async function getServerProxies(domain: string) {
  const res = await fetch(`http://${domain}/api/v1/relay`);

  if (res) {
    return await res.json();
  }
}

export async function assignServerTenants() {
  const servers = await db.getServers();
  const users = await db.getUsers();

  const fetchsList = [];
  for (let server of servers) {
    server.users_count = 0;
    for (const user of users) {
      if (user.server_code == server.code && (user.quota as number) > 10) {
        (server.users_count as number) += 1;
      }
    }

    fetchsList.push(db.putServer(server));
  }

  await Promise.all(fetchsList);
}
