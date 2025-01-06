import { Cloudflare } from ".";
import { Database } from "../database";
import type { Record } from "cloudflare/resources/dns/records.mjs";

interface DNSList {
  domains: string[];
  records: Record[];
}

export class DNS {
  private cloudflare = new Cloudflare();
  private db = new Database();

  async populateDNS(assignWildcard: Boolean = false) {
    const servers = await this.db.getServers();
    const wildcards = await this.db.getWildcards();
    for (const srv of servers) {
      const server: any = srv;
      await this.postDNS(server.domain, server.ip, false);

      if (assignWildcard) {
        let wildcardFetchs: Promise<Record>[] = [];
        for (const wildcard of wildcards) {
          wildcardFetchs.push(this.postDNS(`${wildcard.domain}.${server.domain}`, server.ip, true));

          if (wildcardFetchs.length >= 10) {
            await Promise.all(wildcardFetchs).finally(() => {
              wildcardFetchs = [];
            });
          }
        }
        await Promise.all(wildcardFetchs);
      }
    }
  }

  async getDNSList() {
    const dns: DNSList = {
      domains: [],
      records: [],
    };

    const records = await this.cloudflare.getDNSRecords();
    for (const record of records) {
      dns.domains.push(record.name);
      dns.records.push(record);
    }

    return dns;
  }

  async postDNS(name: string, content: string, proxied: boolean) {
    return await this.cloudflare.postDNSRecord(name, content, proxied);
  }

  async flushDNS() {
    const dns = await this.getDNSList();
    for (const record of dns.records) {
      await this.cloudflare.deleteDNSRecord(record.id as string);
    }
  }

  async flushAndPopulateDNS() {
    await this.flushDNS();
    await this.populateDNS();
  }
}
