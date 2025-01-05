import { Database } from "../database";

const db = new Database();

export async function cleanExpiredUsers() {
  const now = new Date();
  const users = await db.getUsers();
  const expiredIds: number[] = [];
  const oneWeekInMilliseconds = 7 * 24 * 60 * 60 * 1000;

  const fetchsList = [];
  for (const user of users) {
    if (now.getTime() - new Date((user as any).expired).getTime() > oneWeekInMilliseconds) {
      expiredIds.push(user.id as number);
    }
  }

  if (expiredIds) {
    fetchsList.push(db.deleteUser(expiredIds));

    await Promise.all(fetchsList);
  }
}

export async function cleanExceededQuota() {
  let users = await db.getUsers();

  const fetchsList = [];
  for (const user of users) {
    if ((user.quota as number) <= 10) {
      fetchsList.push(
        db.putUser({
          ...user,
          server_code: "",
        })
      );
    }
  }

  await Promise.all(fetchsList);
}
