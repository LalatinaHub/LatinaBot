import { Context, type CommandContext, type SessionFlavor } from "grammy";
import { type Conversation, type ConversationFlavor } from "@grammyjs/conversations";
import { type FileFlavor } from "@grammyjs/files";
import { type SessionData } from "./session";

type _Context = Context & ConversationFlavor & SessionFlavor<SessionData>;
type _FoolishContext = {
  foolish: {
    isAdmin: (ctx: CommandContext<FoolishContext>) => boolean;
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
    timeBetweenRestart: (ctx: FoolishContext) => number;
    fetchsList: Promise<any>[];
  };
};

export type FoolishContext = FileFlavor<_FoolishContext & _Context>;
export type FoolishConversation = Conversation<FoolishContext>;
