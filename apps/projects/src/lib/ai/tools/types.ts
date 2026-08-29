import type { InferUITools, UIDataTypes, UIMessage } from "ai";
import type { tools } from ".";

type MyTools = InferUITools<typeof tools>;

export type MayaMessageMetadata = {
  source?: "text" | "voice";
  voiceAnchorMessageId?: string | null;
  voiceOrder?: number;
};

export type MayaUIMessage = UIMessage<
  MayaMessageMetadata,
  UIDataTypes,
  MyTools
>;
