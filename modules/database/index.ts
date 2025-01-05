import { createClient } from "@libsql/client";

import { v4 as uuidv4 } from "uuid";
import pswd from "generate-password";

export class Database {
  private client = createClient({
    url: process.env.TURSO_DATABASE_URL as string,
    authToken: process.env.TURSO_AUTH_TOKEN as string,
  });

  async getProxy() {
    const proxies = await this.client.execute(
      "SELECT * FROM proxies WHERE vpn != 'shadowsocks' AND conn_mode == 'cdn' ORDER BY RANDOM() LIMIT 1"
    );

    return proxies.rows[0] as any;
  }

  async getServers() {
    const data = await this.client.execute("SELECT * FROM servers;");
    return data.rows;
  }

  async putServer(obj: any) {
    const keys = Object.keys(obj);
    await this.client.execute({
      sql: `UPDATE servers SET ${keys.map((key) => key + " = :" + key).join(", ")} WHERE id = ${obj.id}`,
      args: obj,
    });
  }

  async getWildcards() {
    const data = await this.client.execute("SELECT * FROM wildcards;");
    return data.rows;
  }

  async postWildcard(domain: string) {
    await this.client.execute({
      sql: "INSERT INTO wildcards(domain) VALUES (:domain);",
      args: {
        domain: domain,
      },
    });
  }

  async getUsers() {
    const data = await this.client.execute("SELECT * FROM users;");
    return data.rows;
  }

  async getUser(id: number) {
    const data = await this.client.execute(`SELECT * FROM users WHERE id = ${id};`);
    return data.rows[0] as any;
  }

  async deleteUser(ids: number[]) {
    await this.client.batch(ids.map((id) => `DELETE FROM users WHERE id = ${id};`) as unknown as string[]);
  }

  async postUser(id: number) {
    const expired = new Date();
    expired.setDate(expired.getDate() + 7);

    const token = pswd.generate({ length: 8, numbers: true });

    await this.client.execute({
      sql: "INSERT INTO users VALUES (:id, :token, :password, :expired, :server_code, :quota, :relay, :adblock :vpn)",
      args: {
        id: id,
        token: token,
        password: uuidv4(),
        expired: expired.toISOString().split("T")[0],
        server_code: "",
        quota: 10000,
        relay: "",
        adblock: 1,
        vpn: "vless",
      },
    });
  }

  async putUser(obj: any) {
    const keys = Object.keys(obj);
    await this.client.execute({
      sql: `UPDATE users SET ${keys.map((key) => key + " = :" + key).join(", ")} WHERE id = ${obj.id}`,
      args: obj,
    });
  }

  async getDonation(orderId: string) {
    const data = await this.client.execute(`SELECT FROM donations WHERE order_id = ${orderId};`);
    return data.rows[0];
  }

  async postDonation(orderId: string) {
    await this.client.execute({
      sql: "INSERT INTO donations(order_id) VALUES (:order_id);",
      args: {
        order_id: orderId,
      },
    });
  }
}
