import { toBase64 } from "cloudflare/core.mjs";

export function convertProxyToUrl(proxy: any) {
  if (proxy.vpn == "vmess") {
    const option = {
      v: 2,
      ps: proxy.remark,
      add: proxy.server,
      port: proxy.server_port,
      id: proxy.uuid,
      aid: 0,
      net: proxy.transport,
      path: proxy.path,
      type: "none",
      host: proxy.host,
      tls: proxy.tls ? "tls" : "",
    };

    return `vmess://${toBase64(JSON.stringify(option))}`;
  } else if (proxy.vpn == "trojan" || proxy.vpn == "vless") {
    let uri = new URL(`${proxy.vpn}://`);
    uri.host = `${proxy.password || proxy.uuid}@${proxy.server}`;
    uri.port = proxy.server_port;
    uri.searchParams.append("path", proxy.path);
    uri.searchParams.append("security", proxy.tls ? "tls" : "");
    uri.searchParams.append("type", proxy.transport);
    uri.searchParams.append("sni", proxy.sni);
    uri.hash = proxy.remark;

    return uri.toString();
  }
}
