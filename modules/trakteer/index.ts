import { fetch } from "bun";

export class Trakteer {
  static async getDonations() {
    const res = await fetch("https://api.trakteer.id/v1/public/supports?limit=5&include=order_id", {
      headers: {
        Accept: "application/json",
        "X-Requested-With": "XMLHttpRequest",
        key: process.env.TRAKTEER_TOKEN || "",
      },
    });

    if (res.status == 200) {
      return await res.json();
    }

    return;
  }
}
