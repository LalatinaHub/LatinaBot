import Cloudflare from "cloudflare";
import { Database } from "../database";

export class Worker {
  private cloudflare = new Cloudflare();
  private db = new Database();

  private accountId = process.env.CLOUDFLARE_ACCOUNT_ID || "";
  private zoneId = process.env.CLOUDFLARE_ZONE_ID || "";
  private environment = process.env.WORKER_ENVIRONMENT || "";
  private serviceName = process.env.WORKER_SERVICE_NAME || "";

  async getWorkersSubdomains() {
    const data = await this.cloudflare.workers.domains.list({
      account_id: this.accountId,
      environment: this.environment,
      service: this.serviceName,
      zone_id: this.zoneId,
    });

    return data.result;
  }

  async registerWorkersSubdomains() {
    const workersSubdomains = await this.getWorkersSubdomains();
    const wildcards = await this.db.getWildcards();
    const servers = await this.db.getServers();

    const fetchsList = [];
    for (const server of servers) {
      for (const wildcard of wildcards) {
        const subdomain = `${wildcard.domain}.${this.serviceName}.${server.domain}`;
        let isExists = false;
        for (const workderSubdomain of workersSubdomains) {
          if (subdomain == workderSubdomain.hostname) {
            isExists = true;
          }
        }

        if (!isExists) {
          fetchsList.push(
            this.cloudflare.workers.domains.update({
              account_id: this.accountId,
              environment: this.environment,
              service: this.serviceName,
              zone_id: this.zoneId,
              hostname: subdomain,
            })
          );
        }
      }
    }

    await Promise.all(fetchsList);
  }

  async flushWorkersSubdomains() {
    const workersSubdomains = await this.getWorkersSubdomains();

    const fetchsList = [];
    for (const workersSubdomain of workersSubdomains) {
      fetchsList.push(
        this.cloudflare.workers.domains.delete(workersSubdomain.id as string, { account_id: this.accountId })
      );
    }
  }
}
