import { Database } from "../database";

const db = new Database();

export async function cleanExpiredUsers() {
  const now = new Date();
  const users = await db.getUsers();
  const expiredIds: number[] = [];
  const oneWeekInMilliseconds = 7 * 24 * 60 * 60 * 1000;
  let servers = await db.getServers();

  const fetchsList = [];
  for (const user of users) {
    if (now.getTime() - new Date(user.expired).getTime() > oneWeekInMilliseconds) {
      expiredIds.push(user.id);

      for (let server of servers) {
        if (server.domain == user.premium.domain) {
          server.tenant = server.tenant > 0 ? server.tenant - 1 : 0;
        }
      }
    }
  }

  if (expiredIds) {
    fetchsList.push(db.deletePremium(expiredIds));
    fetchsList.push(db.deleteUser(expiredIds));
    for (const server of servers) {
      fetchsList.push(db.putServer(server));
    }

    await Promise.all(fetchsList);
  }
}

export async function cleanExceededQuota() {
  let users = await db.getUsers();
  let servers = await db.getServers();

  const fetchsList = [];
  for (const user of users) {
    if (user.premium.quota <= 10) {
      for (let server of servers) {
        if (server.domain == user.premium.domain) {
          server.tenant = server.tenant > 0 ? server.tenant - 1 : 0;
          fetchsList.push(
            db.putPremium({
              ...user.premium,
              domain: "",
            })
          );
        }
      }
    }
  }

  for (const server of servers) {
    fetchsList.push(db.putServer(server));
  }

  await Promise.all(fetchsList);
}
