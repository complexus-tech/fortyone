import type { StateAdapter } from "chat";
import { createMemoryState } from "@chat-adapter/state-memory";
import { createRedisState } from "@chat-adapter/state-redis";
import type { BotConfig } from "@/lib/config";

export const createBotState = (config: BotConfig): StateAdapter => {
  if (config.redisUrl) {
    return createRedisState({
      keyPrefix: "fortyone:chat-sdk",
      url: config.redisUrl,
    });
  }

  return createMemoryState();
};
