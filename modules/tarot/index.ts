import { tarotDeck } from "./card";

const baseImageUrl = "https://raw.githubusercontent.com/jeremytarling/python-tarot/refs/heads/master/webapp/static/";

export function getCard() {
  while (true) {
    const card = tarotDeck[Math.floor(Math.random() * tarotDeck.length) - 1];

    if (card.message) {
      return {
        ...card,
        image: baseImageUrl + card.image,
      };
    }
  }
}
