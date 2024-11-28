import { Database } from "../../modules/database";
import { type FoolishConversation, type FoolishContext } from "../context";
import { DNS } from "../../modules/cloudflare/dns";

export async function createWildcard(conversation: FoolishConversation, ctx: FoolishContext) {
  const dns = new DNS();
  const db = new Database();
  const text = await conversation.waitFor(":text");
  const wildcardDomain = text.message?.text;

  if (wildcardDomain && wildcardDomain.match(/^.+\.\D+$/)) {
    ctx.reply("OK proses...");
    await db.postWildcard(wildcardDomain);
    await dns.populateDNS();
    await ctx.reply("Done!");
  } else {
    await ctx.reply("Waduh error!\nProses batal!");
  }
}
