import tls from "tls";

async function sendRequest(host: string, path: string, proxy: any = null) {
  return new Promise((resolve, reject) => {
    const options = {
      host: proxy ? proxy.host : host,
      port: proxy ? proxy.port : 443,
      servername: host,
    };

    const socket = tls.connect(options, () => {
      const request =
        `GET ${path} HTTP/1.1\r\n` + `Host: ${host}\r\n` + `User-Agent: Mozilla/5.0\r\n` + `Connection: close\r\n\r\n`;
      socket.write(request);
    });

    let responseBody = "";
    socket.on("data", (data) => (responseBody += data.toString()));
    socket.on("end", () => {
      const body = responseBody.split("\r\n\r\n")[1] || "";
      resolve(body);
    });

    socket.on("error", (error) => reject(error));

    socket.setTimeout(5000, () => {
      reject(new Error("Request timeout"));
      socket.end();
    });
  });
}

export async function checkIP(proxyIP: string | null) {
  if (proxyIP == undefined) return {};

  const [proxy, port] = proxyIP.split(":");
  const proxyInfo = { host: proxy, port: port };

  try {
    const [ipinfo, myip] = await Promise.all([
      sendRequest("cloudflare-ip.html.zone", "/geo", proxyInfo),
      sendRequest("cloudflare-ip.html.zone", "/geo", null),
    ]);

    const parsedIpInfo = JSON.parse(ipinfo as string);
    const parsedMyIp = JSON.parse(myip as string);

    if (parsedIpInfo.ip && parsedIpInfo.ip !== parsedMyIp.ip) {
      return {
        proxy: proxy,
        port: port,
        proxyip: true,
        ...parsedIpInfo,
      };
    } else {
      return { proxyip: false };
    }
  } catch (error: any) {
    return { proxyip: error.message };
  }
}
