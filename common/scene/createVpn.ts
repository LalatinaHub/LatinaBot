import { InlineKeyboard } from "grammy";
import { Database } from "../../modules/database";
import { getServerProxies } from "../../modules/helper/server";
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

  let message: string = "Pilih protokol sesuai kebutuhan:\n";
  message += "• VMess: Cepat dan stabil, cocok buat streaming atau browsing.\n";
  message += "• VLESS: Hemat data, ideal untuk jaringan kurang stabil.\n";
  message += "• Trojan: Super aman, cocok untuk kamu yang peduli privasi.\n\n";
  message += "Bebas ganti protokol kapan aja!";
  await ctx.editMessageCaption({
    caption: message,
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
  const serversInfo: Promise<{ org: string; ping: number }>[] = [];
  for (const server of servers) {
    const timeStart = new Date().getTime();
    serversInfo.push(
      fetch("https://" + server.domain + "/api/v1/info").then(async (res) => {
        let result = {
          org: "Unknown",
          ping: 0,
        };
        if (res.status == 200) {
          const json = await res.json();
          result = {
            org: json.org,
            ping: new Date().getTime() - timeStart,
          };
        }

        return result;
      })
    );
  }
  await Promise.all(serversInfo);
  message = "Pilih server kesukaanmu!\n";
  message += "Sorry kalo masih dikit, kalo banyak yang donasi pasti makin banyak nantinya 😉\n\n";
  message += "<blockquote expandable>";
  message += "<b>Server's Stats</b>\n\n";
  for (const i in servers) {
    const server = servers[i];
    const serverInfo = serversInfo[i];
    message += `• <b>${server.code}</b> [${server.tenant}/${server.max_tenant}]\n`;
    message += `•• 📡 ${(await serverInfo).ping} ms\n`;
    message += `•• 🪪 ${(await serverInfo).org}\n`;
    if (server.tenant >= server.max_tenant) {
      message += "•• 😭 Penuh\n";
    } else if (server.tenant <= parseInt((server.max_tenant / 2).toString())) {
      message += "•• 😋 Aman\n";
    } else {
      message += "•• 😁 Boleh lah\n";
    }
    message += "\n";
  }
  message += "</blockquote>";
  await ctx.editMessageCaption({
    caption: message,
    parse_mode: "HTML",
    reply_markup: InlineKeyboard.from([
      (() => {
        return servers
          .filter((server) => server.tenant < server.max_tenant)
          .map((server) => InlineKeyboard.text(server.code, server.domain));
      })(),
    ]),
  });

  createVpnData.domain = (
    await conversation.waitForCallbackQuery(servers.filter((data) => data.domain))
  ).callbackQuery.data;

  const res = await getServerProxies(createVpnData.domain);
  if (res) {
    createVpnData.relaysCC = Object.entries(res.proxies)
      .filter((data) => data[0].length < 5 && (data[1] as any)?.history?.[0]?.delay)
      .map((data) => data[0])
      .sort();
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

  message = "Fool VPN menyediakan jalur relay dengan skema:\n";
  message += "Kamu ➡️ Server Fool VPN ➡️ Server Relay ➡️ Internet\n\n";
  message += "Keuntungan relay:\n";
  message += "• Privasi Lebih Terjaga: Alamat IP asli kamu gak kelihatan.\n";
  message += "• Anti Blokir: Bebas akses meski ada pembatasan jaringan.\n";
  message += "• Koneksi Stabil: Relay bikin koneksi tetap lancar meski di jaringan ketat.\n\n";
  message += "Catatan\n";
  message += "• Relay nambahin latensi/ping\n";
  message += "• Relay gak bikin lemot (tergantung server relay)";
  while (!createVpnData.relay) {
    await ctx.editMessageCaption({
      caption: `${message}\n\n${(createVpnData.page || 0) + 1}/${Math.round(
        (createVpnData.relaysCC?.length || 6) / 6
      )}`,
      reply_markup: InlineKeyboard.from([
        ...(() => {
          const keyboard = [];
          for (let i = (createVpnData.page || 0) * 3; i < (createVpnData.page || 0) * 3 + 3; i++) {
            if (keyboardMap[i]) keyboard.push(keyboardMap[i]);
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
    reply_markup: InlineKeyboard.from([
      [InlineKeyboard.text("Gajadi", "cancel"), InlineKeyboard.text("Yakin", "confirm")],
    ]),
  });
}
