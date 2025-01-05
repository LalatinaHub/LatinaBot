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
      token: string;
      password: string;
      expired: Date;
      server_code: string;
      quota: number;
      relay: string;
      adblock: boolean;
      vpn: string;
    }>;
    timeBetweenRestart: (ctx: FoolishContext) => number;
    fetchsList: Promise<any>[];
  };
};

export type FoolishContext = FileFlavor<_FoolishContext & _Context>;
export type FoolishConversation = Conversation<FoolishContext>;
