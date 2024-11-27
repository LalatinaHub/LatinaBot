import { Bot, session } from "grammy";
import { limit } from "@grammyjs/ratelimiter";
import { hydrateFiles } from "@grammyjs/files";
import { conversations, createConversation } from "@grammyjs/conversations";
import { generateUpdateMiddleware } from "telegraf-middleware-console-time";
import { encode } from "html-entities";
import { v4 as uuidv4 } from "uuid";
import pswd from "generate-password";
import "dotenv/config";

// Import Modules
import { Database } from "./modules/database";
import type { FoolishContext } from "./common/context";
import { templateStart } from "./template/start";
import { createVpn } from "./common/scene/createVpn";
import { scanOcrUrl } from "./modules/helper/ocr";
import { Trakteer } from "./modules/trakteer";
import { reloadServers } from "./modules/helper/server";

const bot = new Bot<FoolishContext>(process.env.BOT_TOKEN as string);
const db = new Database();

// Config
bot.api.config.use(hydrateFiles(bot.token));

// Error handler
bot.catch((err) => {
  err.ctx.reply("Runtime error happened!");

  let message: string = "<b>Error Report</b>\n";
  message += `From: ${err.ctx.from?.id}\n`;
  message += `Message: ${encode(err.ctx.message?.text)}\n`;
  message += `Error Message: ${encode(err.message)}\n`;
  message += `Error Stack: \n${encode(err.stack)}`;

  return bot.api.sendMessage(process.env.ADMIN_ID as string, message, {
    parse_mode: "HTML",
  });
});

bot.use(async (ctx, next) => {
  ctx.replyWithChatAction("typing");

  if (ctx.chat?.type != "private") return;
  if (ctx.from?.id != 732796378) {
    return ctx.reply("Bot dalam perbaikan!");
  }

  ctx.foolish = {
    user: async () => {
      while (true) {
        let tempUserData = await db.getUser(ctx.from?.id as number);
        if (tempUserData) return tempUserData;

        await db.postUser(ctx.from?.id as number);
      }
    },
  };

  return next();
});
bot.use(limit());
bot.use(session({ initial: () => ({}) }));
bot.use(conversations());
bot.use(createConversation(createVpn));
bot.use(generateUpdateMiddleware());

// Commands
bot.command("start", async (ctx) => {
  return templateStart(ctx);
});

// On
bot.on("message:photo", async (ctx) => {
  const photo = await ctx.getFile();
  const photoText = await scanOcrUrl(photo.getUrl());
  const photoTextMatch = photoText?.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/gm) || [];
  const orderId = photoTextMatch[photoTextMatch.length - 1];

  if (await db.getDonation(orderId)) {
    return ctx.reply("Order ID Expired!");
  }

  const donations = await Trakteer.getDonations();
  for (const donation of donations.result.data) {
    if (donation.order_id == orderId) {
      const user = await ctx.foolish.user();

      let now = new Date();
      let expired = new Date(user.expired);
      if (expired.getTime() - now.getTime() < 0) expired = now;
      expired.setDate(expired.getDate() + 30);

      await db.putPremium({
        ...user.premium,
        quota: (user.premium?.quota as number) > 0 ? (user.premium?.quota as number) + 100000 : 100000,
      });
      delete user.premium;
      await db.putUser({
        ...user,
        expired: expired.toISOString().split("T")[0],
      });

      await db.postDonation(orderId);
      return templateStart(ctx);
    }
  }
  return ctx.reply("Order ID Invalid!");
});

// Callback query
bot.callbackQuery("m/refresh", (ctx) => {
  return templateStart(ctx, true);
});
bot.callbackQuery("c/vpn", async (ctx) => {
  await ctx.conversation.enter("createVpn");
});

bot.callbackQuery("confirm", async (ctx) => {
  const user = await ctx.foolish.user();
  const data = JSON.parse(ctx.callbackQuery.message?.caption || "{}");

  await db.putPremium({
    ...user.premium,
    type: data.vpn,
    domain: data.domain,
    cc: data.relay,
  });

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
  const user = await ctx.foolish.user();
  await db.putPremium({
    ...user.premium,
    password: uuidv4(),
  });

  await reloadServers();
  return templateStart(ctx, true);
});

bot.callbackQuery("s/adblock", async (ctx) => {
  const user = await ctx.foolish.user();
  await db.putPremium({
    ...user.premium,
    adblock: !user.premium?.adblock,
  });

  return templateStart(ctx, true);
});

bot.start({
  drop_pending_updates: true,
  onStart: () => {
    const server = Bun.serve({
      port: 8080,
      fetch(request) {
        return new Response("Welcome to Bun!");
      },
    });

    console.log("Bot ready!");
    console.log(`Listening on localhost:${server.port}`);
  },
});
