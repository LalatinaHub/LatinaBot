import { InlineKeyboard, type CallbackQueryContext, type CommandContext } from "grammy";
import type { FoolishContext } from "../common/context";
import { Trakteer } from "../modules/trakteer";
import { getCard } from "../modules/tarot";
import { Database } from "../modules/database";

type CtxContext = CommandContext<FoolishContext> | CallbackQueryContext<FoolishContext> | FoolishContext;

export async function templateStart(ctx: CtxContext, edit: boolean = false) {
  const db = new Database();
  const card = getCard();
  const now = new Date();
  const user = await ctx.foolish.user();
  const donations = await Trakteer.getDonations();
  const servers = await db.getServers();

  const expired = new Date(user.expired).getTime() - now.getTime();
  let message: string = `❡ <b>${card.name} on Service</b> ❡\n`;
  message += `<i><u>${card.message}</u></i>\n\n`;

  message += "<blockquote expandable>";
  message += "❂ <b>Profil Kamu</b> ❂\n";
  message += "-".repeat(15) + "\n";
  message += `ID: <code>${ctx.from?.id}</code>\n`;
  message += `Nama: ${ctx.from?.first_name}\n`;
  message += `Password: <code>${user.token}</code>\n`;
  message += `Status: <b>${expired >= 0 ? "Donator" : "Penikmat"}</b>\n`;
  message += `Expired: ${user.expired}\n`;
  message += "</blockquote>\n\n";

  message += "<blockquote>";
  message += "❁ <b>Akun VPN</b> ❁\n";
  message += "-".repeat(15) + "\n";
  message += `Tipe: ${user.vpn}\n`;
  message += `UUID: <tg-spoiler>${user.password}</tg-spoiler>\n`;
  if (user.server_code) {
    message += `Server Code: ${user.server_code}\n`;
    message += `Domain: ${servers.filter((server: any) => server.code == user.server_code)?.[0]?.domain}\n`;
  }
  message += `Relay: ${user.relay}\n`;
  message += `Quota: ${user.quota} MB\n`;
  message += `Adblock: ${user.adblock ? "Hidup" : "Mati"}\n`;
  message += "</blockquote>\n\n";

  message += "<blockquote>";
  message += "✤ <b>Informasi</b> ✤\n";
  message += "-".repeat(15) + "\n";
  message += "☾ Maksimal 10 akun API\n";
  message += "☾ 1x donasi untuk 29 hari premium, berapapun jumlahnya";
  message += "</blockquote>\n\n";

  message += "<blockquote expandable>";
  message += "※ <b>Para Supreme Being</b> ※\n";
  message += "-".repeat(15) + "\n";
  for (const donation of donations.result.data) {
    message += `<b>${donation.supporter_name} [${donation.quantity}]</b>\n`;
    message += `<tg-spoiler>${donation.support_message}</tg-spoiler>\n\n`;
  }
  message += "</blockquote>\n\n\n";

  message += new Date();

  const keyboard = InlineKeyboard.from([
    [
      InlineKeyboard.url(
        "Ambil Akun",
        `http://api.foolvpn.me/sub?format=raw&cdn=104.18.2.2&sni=google.com&mode=cdn,sni&region=Asia&vpn=vmess,vless,trojan&pass=${user.token}`
      ),
      InlineKeyboard.text("Buat Akun", "c/vpn"),
    ],
    [
      InlineKeyboard.url("Docs", "https://fool.azurewebsites.net"),
      InlineKeyboard.url("Converter", "t.me/subxfm_bot"),
      InlineKeyboard.url("Grup", "t.me/foolvpn"),
    ],
    [InlineKeyboard.text("Ganti Password", "c/pass"), InlineKeyboard.text("Ganti UUID", "c/uuid")],
    [InlineKeyboard.text(`${user.adblock ? "Matikan" : "Hidupkan"} Adblock`, "s/adblock")],
    [
      InlineKeyboard.text("List Wildcard", "l/wildcard"),
      //  InlineKeyboard.text("Buat Wildcard", "c/wildcard")
    ],
    [InlineKeyboard.text("❗️ Desclaimer ❗️", "t/desclaimer")],
    [
      InlineKeyboard.text("Cara Donasi", "t/donasi"),
      InlineKeyboard.url("Trakteer", "https://trakteer.id/dickymuliafiqri/tip"),
      // InlineKeyboard.text("Midtrans (Beta)", "m/donasi"),
    ],
    [InlineKeyboard.text("🔄", "m/refresh"), InlineKeyboard.text("ℹ️", "m/info")],
  ]);

  if (edit) {
    ctx.editMessageCaption({
      caption: message,
      parse_mode: "HTML",
      reply_markup: keyboard,
    });
  } else {
    ctx.replyWithPhoto(card.image, {
      caption: message,
      parse_mode: "HTML",
      reply_markup: keyboard,
    });
  }
}
