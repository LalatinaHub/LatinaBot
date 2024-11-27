import { createClient } from "@supabase/supabase-js";
import { v4 as uuidv4 } from "uuid";
import pswd from "generate-password";

export class Database {
  private client = createClient(process.env.SUPABASE_URL as string, process.env.SUPABASE_KEY as string);

  async getServers() {
    const { data, error } = await this.client.from("domains").select();
    if (error) throw error;

    return data;
  }

  async getUser(id: number) {
    const { data, error } = await this.client.from("users").select("*, premium (*)").eq("id", id);
    if (error) throw error;

    return data[0];
  }

  async postUser(id: number) {
    const password = pswd.generate({ length: 8, numbers: true });
    const { error } = await this.client.from("users").insert({ id: id, expired: new Date(), password: password });
    if (error) throw error;

    await this.postPremium(id);
  }

  async putUser(obj: any) {
    const { error } = await this.client.from("users").upsert(obj);
    if (error) throw error;
  }

  async getDonation(orderId: string) {
    const { data, error } = await this.client.from("donations").select().eq("order_id", orderId);
    if (error) throw error;

    return data[0];
  }

  async postDonation(orderId: string) {
    const { error } = await this.client.from("donations").insert({
      order_id: orderId,
    });
    if (error) throw error;
  }

  async postPremium(id: number) {
    const servers = await this.getServers();
    const { error } = await this.client.from("premium").insert({
      id: id,
      password: uuidv4(),
      type: "vmess",
      domain: servers[0].domain,
      quota: 1000,
      cc: "",
      adblock: true,
    });
    if (error) throw error;
  }

  async putPremium(obj: any) {
    const { error } = await this.client.from("premium").upsert(obj);
    if (error) throw error;
  }
}
