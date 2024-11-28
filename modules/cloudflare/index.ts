import { Database } from "../database";
import CF from "cloudflare";
import type { Zone } from "cloudflare/resources/zones/zones.mjs";

export class Cloudflare {
  private client = new CF({
    apiEmail: process.env.CLOUDFLARE_EMAIL,
    apiKey: process.env.CLOUDFLARE_API_KEY,
  });
  private db = new Database();
  private _zone: Zone | undefined;

  async getZone() {
    if (this._zone == undefined) {
      const servers = await this.db.getServers();
      const domain = (servers[0].domain as string).match(/(^\w+\.)(.+$)/)?.[2];

      const zones = await this.client.zones.list({
        name: domain,
      });

      this._zone = zones.result[0];
    }
    return this._zone;
  }

  async getDNSRecords() {
    const zone = await this.getZone();
    const records = await this.client.dns.records.list({
      zone_id: zone.id,
    });

    return records.result;
  }

  async postDNSRecord(name: string, content: string) {
    const zone = await this.getZone();
    const records = await this.getDNSRecords();
    for (const record of records) {
      if (record.name == name) return record;
    }

    const record = await this.client.dns.records.create({
      zone_id: zone.id,
      name: name,
      type: "A",
      content: content,
      proxied: true
    });

    return record;
  }

  async deleteDNSRecord(recordId: string) {
    const zone = await this.getZone();
    return await this.client.dns.records.delete(recordId, {
      zone_id: zone.id,
    });
  }
}
