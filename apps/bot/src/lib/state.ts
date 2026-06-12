import type { StateAdapter } from "chat";
import { createMemoryState } from "@chat-adapter/state-memory";
import { createRedisState } from "@chat-adapter/state-redis";

import type { BotConfig } from "@/lib/config";

export const createBotState = (config: BotConfig): StateAdapter => {
  if (config.stateDriver === "redis") {
    return createRedisState({
      keyPrefix: "fortyone:chat-sdk",
      url: config.redisUrl,
    });
  }

  if (config.isProduction) {
    throw new Error("Production bot runtime requires Redis state.");
  }

  return createMemoryState();
};
