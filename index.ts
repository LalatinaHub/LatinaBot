import { Bot, InputFile, session } from "grammy";
import { run } from "@grammyjs/runner";
import { limit } from "@grammyjs/ratelimiter";
import { hydrateFiles } from "@grammyjs/files";
import { conversations, createConversation } from "@grammyjs/conversations";
import { generateUpdateMiddleware } from "telegraf-middleware-console-time";
import { encode } from "html-entities";
import { v4 as uuidv4 } from "uuid";
import { fetch } from "bun";
import { exists, mkdir } from "node:fs/promises";
import pswd from "generate-password";
import "dotenv/config";

// Import Modules
import { Database } from "./modules/database";
import type { FoolishContext } from "./common/context";
import { templateStart } from "./template/start";
import { createVpn } from "./common/scene/createVpn";
import { createWildcard } from "./common/scene/createWildcard";
import { scanOcrUrl } from "./modules/helper/ocr";
import { Trakteer } from "./modules/trakteer";
import { reloadServers } from "./modules/helper/server";
import { cleanExceededQuota, cleanExpiredUsers } from "./modules/helper/users";
import sharp from "sharp";

const bot = new Bot<FoolishContext>(process.env.BOT_TOKEN as string);
const db = new Database();
const adminId = process.env.ADMIN_ID as unknown as number;
let localOrderId: string = "";

// Config
bot.api.config.use(hydrateFiles(bot.token));

// Error handler
bot.catch((err) => {
  err.ctx.reply(`Waduh error!\nCoba bilang <a href="tg://user?id=${adminId}">admin</a>`, {
    parse_mode: "HTML",
  });

  let message: string = "<b>Error Report</b>\n";
  message += `From: ${err.ctx.from?.id}\n`;
  message += `Message: ${encode(err.ctx.message?.text)}\n`;
  message += `Error Message: ${encode(err.message)}\n`;
  message += `Error Stack: \n${encode(err.stack)}`;

  return bot.api.sendMessage(adminId, message, {
    parse_mode: "HTML",
  });
});

bot.use(async (ctx, next) => {
  ctx.replyWithChatAction("typing");

  if (ctx.chat?.type != "private") return;

  ctx.foolish = {
    user: async () => {
      while (true) {
        let tempUserData = await db.getUser(ctx.from?.id as number);
        if (tempUserData) return tempUserData;

        await db.postUser(ctx.from?.id as number);
      }
    },
    timeBetweenRestart: (ctx) => {
      return (new Date().getTime() - ctx.session.lastRestart.getTime()) / 60 / 1000;
    },
    fetchsList: [],
  };

  return next();
});

bot.use(limit());
bot.use(session({ initial: () => ({ lastRestart: new Date() }) }));
bot.use(conversations());
bot.use(createConversation(createVpn));
bot.use(createConversation(createWildcard));
bot.use(generateUpdateMiddleware());

// Commands
bot.command("start", async (ctx) => {
  return templateStart(ctx);
});

bot.command("new_order", async (ctx) => {
  localOrderId = uuidv4();
  await sharp({
    create: {
      width: 1200,
      height: 200,
      channels: 4,
      background: { r: 255, g: 255, b: 255, alpha: 1 },
    },
  })
    .composite([
      {
        input: {
          text: {
            text: localOrderId,
            width: 1200,
            height: 64,
            font: "sans",
            rgba: true,
          },
        },
      },
    ])
    .jpeg()
    .toFile("./temp/localOrder.jpeg");
  return ctx.replyWithPhoto(new InputFile("./temp/localOrder.jpeg"));
});

// On
bot.on("message:photo", async (ctx) => {
  const photo = await ctx.getFile();
  const photoText = await scanOcrUrl(photo.getUrl());
  const photoTextMatch = photoText?.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/gm);
  if (photoTextMatch) {
    const orderId = photoTextMatch[photoTextMatch.length - 1];

    if (await db.getDonation(orderId)) {
      return ctx.reply("Order ID Expired!");
    }

    const donations = await Trakteer.getDonations();
    let isVerified: boolean = false;
    if (localOrderId == orderId) {
      isVerified = true;
    } else {
      for (const donation of donations.result.data) {
        if (donation.order_id == orderId) {
          isVerified = true;
          break;
        }
      }
    }

    if (isVerified) {
      const user = await ctx.foolish.user();

      let now = new Date();
      let expired = new Date(user.expired);
      if (expired.getTime() - now.getTime() < 0) expired = now;
      expired.setDate(expired.getDate() + 30);

      ctx.foolish.fetchsList.push(
        db.putPremium({
          ...user.premium,
          quota: (user.premium?.quota as number) > 0 ? (user.premium?.quota as number) + 250000 : 250000,
        })
      );
      delete user.premium;
      ctx.foolish.fetchsList.push(
        db.putUser({
          ...user,
          expired: expired.toISOString().split("T")[0],
        })
      );

      ctx.foolish.fetchsList.push(db.postDonation(orderId));
      await Promise.all(ctx.foolish.fetchsList);
      await reloadServers();
      return templateStart(ctx);
    }
    return ctx.reply("Order ID Invalid!");
  }

  return ctx.reply("Gagal membaca Order ID!\nCoba crop fotonya biar lebih jelas.");
});

// Callback query
bot.callbackQuery("m/refresh", (ctx) => {
  return templateStart(ctx, true);
});
bot.callbackQuery("c/vpn", async (ctx) => {
  const timeBetween = ctx.foolish.timeBetweenRestart(ctx);
  if (timeBetween < 5) {
    return ctx.answerCallbackQuery({
      text: `Ups, ada delay ${5 - parseInt(timeBetween.toString())} menit`,
      show_alert: true,
    });
  }

  const user = await ctx.foolish.user();
  if ((user.premium?.quota as number) <= 10) {
    return ctx.answerCallbackQuery({
      text: "Kuota kamu habis kak!",
      show_alert: true,
    });
  }

  await ctx.conversation.enter("createVpn");
});

bot.callbackQuery("confirm", async (ctx) => {
  const user = await ctx.foolish.user();
  const servers = await db.getServers();
  const data = JSON.parse(ctx.callbackQuery.message?.caption || "{}");
  for (const server of servers) {
    if (server.domain == data.domain) {
      ctx.foolish.fetchsList.push(
        db.putServer({
          ...server,
          tenant: server.tenant + 1,
        })
      );
    }
    if (server.domain == user.premium?.domain) {
      ctx.foolish.fetchsList.push(
        db.putServer({
          ...server,
          tenant: server.tenant > 0 ? server.tenant - 1 : 0,
        })
      );
    }
  }

  ctx.foolish.fetchsList.push(
    db.putPremium({
      ...user.premium,
      type: data.vpn,
      domain: data.domain,
      cc: data.relay,
    })
  );

  ctx.session.lastRestart = new Date();
  await Promise.all(ctx.foolish.fetchsList);
  await reloadServers();
  return templateStart(ctx, true);
});

bot.callbackQuery("cancel", async (ctx) => {
  return templateStart(ctx, true);
});

bot.callbackQuery("t/desclaimer", (ctx) => {
  let message: string = "Semua akun yang disediakan oleh API, merupakan akun yang tersedia secara bebas di Internet.\n";
  message += "Layanan ini tidak melakukan aktivitas tracing, logging, atau semacamnya!.";

  ctx.answerCallbackQuery({
    text: message,
    show_alert: true,
  });
});

bot.callbackQuery("t/donasi", (ctx) => {
  let message: string = "<b>Cara Melakukan Donasi</b>\n";
  message += "1. Buka halaman donasi\n";
  message += "2. Selesaikan donasi\n";
  message += "3. Simpan bukti donasi\n";
  message += "4. Kirimkan bukti donasi ke bot\n";
  message += "5. Crop fotonya biar lebih jelas (optional)\n\n\n";
  message += "↓ Tombol Donasi";

  ctx.replyWithPhoto("https://okzpqehvbvtzrjzohjtw.supabase.co/storage/v1/object/public/assets/donasi_1.png", {
    caption: "↑ Contoh bukti donasi ↑\n\n" + message,
    parse_mode: "HTML",
  });
});

bot.callbackQuery("c/pass", async (ctx) => {
  const user = await ctx.foolish.user();
  delete user.premium;

  await db.putUser({
    ...user,
    password: pswd.generate({
      length: 8,
      numbers: true,
    }),
  });

  return templateStart(ctx, true);
});

bot.callbackQuery("c/uuid", async (ctx) => {
  const timeBetween = ctx.foolish.timeBetweenRestart(ctx);
  if (timeBetween < 5) {
    return ctx.answerCallbackQuery({
      text: `Ups, ada delay ${5 - parseInt(timeBetween.toString())} menit`,
      show_alert: true,
    });
  }

  const user = await ctx.foolish.user();
  await db.putPremium({
    ...user.premium,
    password: uuidv4(),
  });

  ctx.session.lastRestart = new Date();
  await reloadServers();
  return templateStart(ctx, true);
});

bot.callbackQuery("c/wildcard", async (ctx) => {
  const user = await ctx.foolish.user();

  let now = new Date();
  let expired = new Date(user.expired);
  if (expired.getTime() - now.getTime() < 0) {
    return ctx.answerCallbackQuery({
      text: "YDDA\nYang Donasi Donasi Ajah :>",
      show_alert: true,
    });
  }

  await ctx.editMessageCaption({
    caption: "OK, mau domain apa ?\n\ncontoh: zoom.us",
  });
  ctx.conversation.enter("createWildcard");
});

bot.callbackQuery("l/wildcard", async (ctx) => {
  const servers = await db.getServers();
  const wildcardList = await db.getWildcards();

  let message = "Daftar Wildcard:\n";
  message += "<blockquote expandable>";
  for (const wildcard of wildcardList) {
    message += `• ${wildcard.domain}\n`;
  }
  message += "</blockquote>\n\n";

  message += "Contoh:\n";
  message += `<code>${wildcardList[0].domain}.${servers[0].domain}</code>`;

  ctx.reply(message, {
    parse_mode: "HTML",
    link_preview_options: {
      is_disabled: true,
    },
  });
});

bot.callbackQuery("s/adblock", async (ctx) => {
  const user = await ctx.foolish.user();
  await db.putPremium({
    ...user.premium,
    adblock: !user.premium?.adblock,
  });

  ctx.answerCallbackQuery({
    text: "Kamu harus bikin akun vpn baru, uuid baru, atau tunggu server restart untuk merasakan perubahannya.",
    show_alert: true,
  });
  return templateStart(ctx, true);
});

(async () => {
  if (!(await exists("./temp"))) {
    await mkdir("./temp");
  }

  run(bot);

  const server = Bun.serve({
    port: 8080,
    fetch(request) {
      return new Response("Welcome to Bun!");
    },
  });

  const cred = await fetch(process.env.SERVICE_ACCOUNT_URL || "");
  Bun.write("./gcloud-cred.json", cred);

  console.log("Bot ready!");
  console.log(`Listening on localhost:${server.port}`);

  // Interval functions
  setInterval(async () => {
    console.log("Running interval functions...");
    await cleanExpiredUsers();
    await cleanExceededQuota();
  }, 1000 * 60 * 5);
})();
