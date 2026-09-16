import type { UIMessage } from "ai";
import { record, resultStories } from "../components/message-model";

export type MayaResponseSnapshot = {
  chatId: string;
  requestId: string;
  mode: "send" | "approval";
  messageIds: readonly string[];
  assistant: UIMessage | null;
};

export type MayaResponseRow =
  | { id: string; kind: "message"; message: UIMessage }
  | { id: string; kind: "thinking" };

/** Capture values at admission: approval continuations reuse the assistant ID. */
export const captureMayaResponse = ({
  chatId,
  requestId,
  mode,
  messages,
}: {
  chatId: string;
  requestId: string;
  mode: MayaResponseSnapshot["mode"];
  messages: readonly UIMessage[];
}): MayaResponseSnapshot => {
  const last = messages.at(-1);
  return {
    chatId,
    requestId,
    mode,
    messageIds: messages.map((message) => message.id),
    // UI messages contain JSON data. Copy the continuation baseline so stream
    // mutations cannot turn its existing text into a newly received response.
    assistant:
      mode === "approval" && last?.role === "assistant"
        ? (JSON.parse(JSON.stringify(last)) as UIMessage)
        : null,
  };
};

const hasNaturalLanguage = (message: UIMessage) =>
  message.parts.some((part) => part.type === "text" && part.text.trim());

/** Matches content rendered by MayaMessage, excluding invisible protocol parts. */
const hasVisibleContent = (message: UIMessage) =>
  message.parts.some((part) => {
    if (part.type === "text") return Boolean(part.text.trim());
    if (part.type === "file") return message.role === "user";
    if (!part.type.startsWith("tool-") && part.type !== "dynamic-tool")
      return false;
    const tool = record(part);
    if (!tool) return false;
    if (
      tool.state === "approval-requested" &&
      typeof record(tool.approval)?.id === "string"
    )
      return true;
    if (tool.state === "output-denied" || tool.state === "output-error")
      return true;
    const output = record(tool.output);
    if (output)
      return Boolean(
        resultStories(output).length ||
          (typeof output.message === "string" && output.message) ||
          (typeof output.error === "string" && output.error),
      );
    return false;
  });

const messageRow = (message: UIMessage, id = message.id): MayaResponseRow[] =>
  hasVisibleContent(message) ? [{ id, kind: "message", message }] : [];

const splitContinuation = (message: UIMessage, baseline: UIMessage) => {
  const history: UIMessage = { ...message, parts: [] };
  const response: UIMessage = { ...message, parts: [] };
  message.parts.forEach((part, index) => {
    const previous = baseline.parts[index];
    if (!previous) {
      response.parts.push(part);
      return;
    }
    if (part.type === "text" && previous.type === "text") {
      if (part.text.startsWith(previous.text)) {
        history.parts.push({ ...part, text: previous.text });
        response.parts.push({
          ...part,
          text: part.text.slice(previous.text.length),
        });
      } else response.parts.push(part);
      return;
    }
    history.parts.push(part);
  });
  return { history, response };
};

/** One response slot survives empty starts/tool deltas and becomes the reply. */
export const selectMayaResponseRows = ({
  chatId,
  messages,
  response,
  active,
}: {
  chatId: string;
  messages: readonly UIMessage[];
  response: MayaResponseSnapshot | null;
  active: boolean;
}): MayaResponseRow[] => {
  const ordinaryRows = () => messages.flatMap((message) => messageRow(message));
  if (!response || response.chatId !== chatId) return ordinaryRows();

  const slotId = `maya-response:${response.requestId}`;
  const thinking: MayaResponseRow = { id: slotId, kind: "thinking" };
  if (response.mode === "approval") {
    const baseline = response.assistant;
    if (!baseline) return ordinaryRows();
    return messages.flatMap((message) => {
      if (message.id !== baseline.id) return messageRow(message);
      const parts = splitContinuation(message, baseline);
      if (active && !hasNaturalLanguage(parts.response))
        return [...messageRow(baseline), thinking];
      return [
        ...messageRow(parts.history),
        ...messageRow(parts.response, slotId),
      ];
    });
  }

  // Preparing files/context has not appended the new user turn yet. Inserting
  // progress earlier would make it jump when the SDK eventually adds that row.
  const knownIds = new Set(response.messageIds);
  let userIndex = -1;
  messages.forEach((message, index) => {
    if (message.role === "user" && !knownIds.has(message.id)) userIndex = index;
  });
  if (userIndex < 0) return ordinaryRows();
  let assistantIndex = -1;
  messages.forEach((message, index) => {
    if (index > userIndex && message.role === "assistant")
      assistantIndex = index;
  });
  if (assistantIndex < 0)
    return [...ordinaryRows(), ...(active ? [thinking] : [])];
  return messages.flatMap((message, index) => {
    if (index !== assistantIndex) return messageRow(message);
    if (active && !hasNaturalLanguage(message)) return [thinking];
    return messageRow(message, slotId);
  });
};
