import type { InferUITools, UIDataTypes, UIMessage } from "ai";
import type { MayaUIDataTypes } from "../google-drive-context";
import type { MayaActionLease } from "../action-lease";
import type { tools } from ".";

type MyTools = InferUITools<typeof tools>;

export type MayaMessageMetadata = {
  actionLease?: MayaActionLease;
  source?: "text" | "voice";
  voiceAnchorMessageId?: string | null;
  voiceOrder?: number;
};

export type MayaUIMessage = UIMessage<
  MayaMessageMetadata,
  MayaUIDataTypes & UIDataTypes,
  MyTools
>;
