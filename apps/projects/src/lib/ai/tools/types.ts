import type { InferUITools, UIDataTypes, UIMessage } from "ai";
import type { tools } from ".";

type MyTools = InferUITools<typeof tools>;

export type MayaUIMessage = UIMessage<unknown, UIDataTypes, MyTools>;
