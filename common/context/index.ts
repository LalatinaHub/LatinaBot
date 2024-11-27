import { Context } from "grammy";
import { type Conversation, type ConversationFlavor } from "@grammyjs/conversations";
import { type FileFlavor } from "@grammyjs/files";

type _Context = Context & ConversationFlavor;
type _FoolishContext = {
  foolish: {
    user: () => Promise<{
      id: number;
      expired: Date;
      password: string;
      premium?: {
        id: number;
        password: string;
        type: string;
        domain: string;
        quota: number;
        cc: string;
        adblock: boolean;
      };
    }>;
  };
};

export type FoolishContext = FileFlavor<_FoolishContext & _Context>;
export type FoolishConversation = Conversation<FoolishContext>;