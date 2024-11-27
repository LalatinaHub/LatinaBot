import { InlineKeyboard } from "grammy";
import { Database } from "../../modules/database";
import { type FoolishConversation, type FoolishContext } from "../context";
import { fetch } from "bun";

export async function createVpn(conversation: FoolishConversation, ctx: FoolishContext) {
  const db = new Database();
  let createVpnData: {
    vpn: string;
    domain: string;
    relay: string;
    page?: number;
    relaysCC?: string[];
  } = {
    vpn: "",
    domain: "",
    relay: "",
    page: 0,
    relaysCC: [""],
  };

  await ctx.editMessageCaption({
    caption: "Silahkan pilih protokol",
    reply_markup: InlineKeyboard.from([
      [
        InlineKeyboard.text("VMess", "vmess"),
        InlineKeyboard.text("VLESS", "vless"),
        InlineKeyboard.text("Trojan", "trojan"),
      ],
    ]),
  });

  createVpnData.vpn = (await conversation.waitForCallbackQuery(["vmess", "vless", "trojan"])).callbackQuery.data;

  const servers = await db.getServers();
  await ctx.editMessageCaption({
    caption: "Silahkan pilih server",
    reply_markup: InlineKeyboard.from([
      (() => {
        return servers.map((data) => InlineKeyboard.text(data.code, data.domain));
      })(),
    ]),
  });

  createVpnData.domain = (
    await conversation.waitForCallbackQuery(servers.filter((data) => data.domain))
  ).callbackQuery.data;

  const res = await fetch("https://" + createVpnData.domain + "/relay");
  if (res.status == 200) {
    createVpnData.relaysCC = [...new Set(((await res.json()) as []).map((data: any) => data.country_code))].sort();
  } else {
    createVpnData.relay = "Tanpa Relay";
  }

  createVpnData.relaysCC?.unshift("Tanpa Relay");
  const keyboardMap: any = [[]];
  const relayCC = createVpnData.relaysCC;
  for (const i in relayCC) {
    keyboardMap[keyboardMap.length - 1].push(InlineKeyboard.text(relayCC[parseInt(i)]));
    if (parseInt(i) && (!(parseInt(i) % 6) || parseInt(i) == relayCC.length - 1)) {
      keyboardMap.push([
        parseInt(i) > 6 || parseInt(i) == relayCC.length - 1
          ? InlineKeyboard.text("〈 Prev", `${Math.round(parseInt(i) / 6 - 2)}`)
          : InlineKeyboard.text("⛔️ End", "End"),
        parseInt(i) < relayCC.length - 1
          ? InlineKeyboard.text("Next 〉", `${Math.round(parseInt(i) / 6)}`)
          : InlineKeyboard.text("End ⛔️", "End"),
      ]);
    }
    if (parseInt(i) && !(parseInt(i) % 3)) keyboardMap.push([]);
  }

  while (!createVpnData.relay) {
    await ctx.editMessageCaption({
      caption: `Silahkan pilih relay\n\nPage: ${(createVpnData.page || 0) + 1}/${Math.round(
        (createVpnData.relaysCC?.length || 6) / 6
      )}`,
      reply_markup: InlineKeyboard.from([
        ...(() => {
          const keyboard = [];
          for (let i = (createVpnData.page || 0) * 3; i < (createVpnData.page || 0) * 3 + 3; i++) {
            keyboard.push(keyboardMap[i]);
          }

          return keyboard;
        })(),
      ]),
    });

    const relay = (
      await conversation.waitForCallbackQuery(
        keyboardMap
          .map((keyboard: any) => keyboard.map((data: any) => data.callback_data))
          .flat()
          .filter((data: any) => data != "End")
      )
    ).callbackQuery.data;
    if (!isNaN(parseInt(relay))) {
      createVpnData.page = parseInt(relay);
    } else {
      createVpnData.relay = relay;
    }
  }

  delete createVpnData.relaysCC;
  delete createVpnData.page;
  createVpnData.relay = createVpnData.relay == "Tanpa Relay" ? "" : createVpnData.relay;

  return ctx.editMessageCaption({
    caption: `${JSON.stringify(createVpnData, null, "\t")}`,
    reply_markup: InlineKeyboard.from([[InlineKeyboard.text("Ya", "confirm"), InlineKeyboard.text("Tidak", "cancel")]]),
  });
}
