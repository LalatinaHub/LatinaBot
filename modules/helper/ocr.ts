import { fetch } from "bun";
import vision from "@google-cloud/vision";

export async function scanOcrUrl(url: string) {
  const file = await fetch(url);
  const client = new vision.ImageAnnotatorClient({
    keyFilename: "./gcloud-cred.json",
  });

  const [result] = await client.textDetection(Buffer.from(await file.arrayBuffer()));

  if (!result.error) {
    return result.fullTextAnnotation?.text;
  }

  return;
}
