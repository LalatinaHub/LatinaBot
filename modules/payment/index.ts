import { fetch } from "bun";

const transactionUrl = "https://app.midtrans.com/snap/v1/transactions";

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
          authorization: `Basic ${this.apiKey}`,
        },
        body: JSON.stringify({
          transaction_details: { order_id: new Date().getTime(), gross_amount: amount },
          credit_card: { secure: false },
          page_expiry: {
            duration: 5,
            unit: "minutes",
          },
        }),
      });

      if (res.status == 201) {
        returnValue = {
          error: false,
          message: (await res.json()).token,
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
    const checkUrl = `${transactionUrl}/${token}/status`;

    try {
      const res = await fetch(checkUrl);

      if (res.status == 200) {
        if ((await res.json()).transaction_status == "success") {
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
