import { CronJob } from "cron";

export class Cron {
  private cronJobList: CronJob<any, any>[] = [];

  register(func: Function, cronTime: string) {
    const job = new CronJob(cronTime, () => func(), null, false, "Asia/Jakarta");
    this.cronJobList.push(job);
  }

  start() {
    for (const job of this.cronJobList) {
      job.start();
    }
  }
}
