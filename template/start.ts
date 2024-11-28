import { InlineKeyboard, type CallbackQueryContext, type CommandContext } from "grammy";
import type { FoolishContext } from "../common/context";
import { Trakteer } from "../modules/trakteer";

type CtxContext = CommandContext<FoolishContext> | CallbackQueryContext<FoolishContext> | FoolishContext;

export async function templateStart(ctx: CtxContext, edit: boolean = false) {
  const now = new Date();
  const user = await ctx.foolish.user();
  const donations = await Trakteer.getDonations();

  const expired = new Date(user.expired).getTime() - now.getTime();
  let message: string = "❡ <b>Foolish on Service</b> ❡\n";
  message += "<i><u>Embracing the Journey, Trusting the Unknown</u></i>\n\n";

  message += "<blockquote expandable>";
  message += "❂ <b>Profil Kamu</b> ❂\n";
  message += "-".repeat(15) + "\n";
  message += `ID: <code>${ctx.from?.id}</code>\n`;
  message += `Nama: ${ctx.from?.first_name}\n`;
  message += `Password: <code>${user.password}</code>\n`;
  message += `Status: <b>${expired >= 0 ? "Donator" : "Penikmat"}</b>\n`;
  message += `Expired: ${user.expired}\n`;
  message += "</blockquote>\n\n";

  message += "<blockquote>";
  message += "❁ <b>Akun VPN</b> ❁\n";
  message += "-".repeat(15) + "\n";
  message += `Tipe: ${user.premium?.type}\n`;
  message += `UUID: <tg-spoiler>${user.premium?.password}</tg-spoiler>\n`;
  message += `Domain: ${user.premium?.domain}\n`;
  message += `Relay: ${user.premium?.cc}\n`;
  message += `Quota: ${user.premium?.quota} MB\n`;
  message += `Adblock: ${user.premium?.adblock ? "Hidup" : "Mati"}\n`;
  message += "</blockquote>\n\n";

  message += "<blockquote>";
  message += "✤ <b>Informasi</b> ✤\n";
  message += "-".repeat(15) + "\n";
  message += "☾ Maksimal 10 akun API\n";
  message += "☾ 1x donasi untuk 29 hari premium, berapapun jumlahnya";
  message += "</blockquote>\n\n";

  message += "<blockquote expandable>";
  message += "※ <b>Sultan Permen Titid</b> ※\n";
  message += "-".repeat(15) + "\n";
  for (const donation of donations.result.data) {
    message += `<b>${donation.supporter_name} [${donation.quantity}]</b>\n`;
    message += `<tg-spoiler>${donation.support_message}</tg-spoiler>\n\n`;
  }
  message += "</blockquote>\n\n\n";

  message += new Date();

  const keyboard = InlineKeyboard.from([
    [InlineKeyboard.text("Refresh", "m/refresh")],
    [
      InlineKeyboard.url(
        "Ambil Akun",
        `http://fool.azurewebsites.net/get?format=raw&cdn=104.18.2.2&sni=google.com&mode=cdn,sni&region=Asia&vpn=vmess,vless,trojan&pass=${user.password}`
      ),
      InlineKeyboard.text("Buat Akun", "c/vpn"),
    ],
    [
      InlineKeyboard.url("Docs", "https://fool.azurewebsites.net"),
      InlineKeyboard.url("Converter", "t.me/subxfm_bot"),
      InlineKeyboard.url("Grup", "t.me/foolvpn"),
    ],
    [InlineKeyboard.text("Ganti Password", "c/pass"), InlineKeyboard.text("Ganti UUID", "c/uuid")],
    [InlineKeyboard.text(`${user.premium?.adblock ? "Matikan" : "Hidupkan"} Adblock`, "s/adblock")],
    [InlineKeyboard.text("List Wildcard", "l/wildcard"), InlineKeyboard.text("Buat Wildcard", "c/wildcard")],
    [InlineKeyboard.text("❗️ Desclaimer ❗️", "t/desclaimer")],
    [
      InlineKeyboard.text("Cara Donasi", "t/donasi"),
      InlineKeyboard.url("Donasi", "https://trakteer.id/dickymuliafiqri/tip"),
    ],
  ]);

  if (edit) {
    ctx.editMessageCaption({
      caption: message,
      parse_mode: "HTML",
      reply_markup: keyboard,
    });
  } else {
    ctx.replyWithPhoto("https://i0.wp.com/altargods.com/wp-content/uploads/2024/01/altar-gods-tarot-the-fool.jpg", {
      caption: message,
      parse_mode: "HTML",
      reply_markup: keyboard,
    });
  }
}
