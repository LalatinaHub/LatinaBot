import { fetch } from "bun";

const transactionUrl = "https://api.xendit.co/qr_codes";

export class Payment {
  private apiKey = process.env.GATEWAY_PAYMENT_API_KEY || "";

  async makePayment(amount: number) {
    let returnValue = {
      error: false,
      message: "",
    };

    try {
      const res = await fetch(transactionUrl, {
        method: "POST",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          "api-version": "2022-07-31",
          authorization: `Basic ${btoa(this.apiKey)}`,
        },
        body: JSON.stringify({
          reference_id: `order-id-${new Date().getTime()}`,
          type: "DYNAMIC",
          currency: "IDR",
          amount: amount,
        }),
      });

      if (res.status == 201) {
        returnValue = {
          error: false,
          message: (await res.json()).qr_string,
        };
      } else {
        throw new Error(res.statusText);
      }
    } catch (e: any) {
      returnValue = {
        error: true,
        message: e.message,
      };
    }

    return returnValue;
  }

  async checkPayment(token: string) {
    let returnValue = {
      error: false,
      message: "",
    };
    const checkUrl = `${transactionUrl}/${token}`;

    try {
      const res = await fetch(checkUrl, {
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          "api-version": "2022-07-31",
          authorization: `Basic ${btoa(this.apiKey)}`,
        },
      });

      if (res.status == 200) {
        if ((await res.json()).status == "SUCCEEDED") {
          returnValue = {
            error: false,
            message: "success",
          };
        }
      } else {
        throw new Error(res.statusText);
      }
    } catch (e: any) {
      returnValue = {
        error: true,
        message: e.message,
      };
    }

    return returnValue;
  }
}
