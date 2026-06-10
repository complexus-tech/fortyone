import { Chat } from "chat";
import { createSlackAdapter } from "@chat-adapter/slack";
import { createMemoryState } from "@chat-adapter/state-memory";
import { createAgentStream } from "@/lib/agent";

export const bot = new Chat({
  userName: "photon",
  adapters: {
    slack: createSlackAdapter(),
  },
  state: createMemoryState(),
});

bot.onNewMention(async (thread, message) => {
  await thread.subscribe();
  await thread.post(createAgentStream(message.text));
});

bot.onDirectMessage(async (thread, message) => {
  await thread.subscribe();
  await thread.post(createAgentStream(message.text));
});

bot.onSubscribedMessage(async (thread, message) => {
  if (message.author?.isBot) {
    return;
  }
  await thread.post(createAgentStream(message.text));
});

bot.onSlashCommand("/fortyone", async (event) => {
  const prompt = event.text.trim() || "What can you help me with?";
  await event.channel.post(createAgentStream(prompt));
});
