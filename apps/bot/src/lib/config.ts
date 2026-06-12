type StateDriver = "memory" | "redis";

export type BotConfig = {
  botUserName: string;
  fortyOneApiUrl: string;
  fortyOneAppUrl: string;
  fortyOneServiceToken: string;
  isProduction: boolean;
  openAIModel: string;
  redisUrl: string;
  slackBotToken: string;
  slackClientId: string;
  slackClientSecret: string;
  slackEncryptionKey: string;
  slackSigningSecret: string;
  stateDriver: StateDriver;
};

const readEnv = (key: string) => process.env[key]?.trim() ?? "";

const readStateDriver = (): StateDriver => {
  const value = readEnv("CHATSDK_STATE_DRIVER");
  if (value === "memory" || value === "redis") {
    return value;
  }
  return readEnv("REDIS_URL") ? "redis" : "memory";
};

export const getBotConfig = (): BotConfig => {
  const isProduction = process.env.NODE_ENV === "production";

  return {
    botUserName: readEnv("SLACK_BOT_USERNAME") || "maya",
    fortyOneApiUrl:
      readEnv("FORTYONE_API_URL") || "http://localhost:8000",
    fortyOneAppUrl:
      readEnv("FORTYONE_APP_URL") || "http://localhost:3000",
    fortyOneServiceToken: readEnv("FORTYONE_BOT_TOKEN"),
    isProduction,
    openAIModel: readEnv("BOT_OPENAI_MODEL") || "gpt-5.4-mini",
    redisUrl: readEnv("REDIS_URL"),
    slackBotToken: readEnv("SLACK_BOT_TOKEN"),
    slackClientId: readEnv("SLACK_CLIENT_ID"),
    slackClientSecret: readEnv("SLACK_CLIENT_SECRET"),
    slackEncryptionKey: readEnv("SLACK_INSTALLATION_ENCRYPTION_KEY"),
    slackSigningSecret: readEnv("SLACK_SIGNING_SECRET"),
    stateDriver: readStateDriver(),
  };
};

export const assertProductionConfig = (config: BotConfig) => {
  if (!config.isProduction) {
    return;
  }

  const missing = [
    ["FORTYONE_BOT_TOKEN", config.fortyOneServiceToken],
    ["REDIS_URL", config.redisUrl],
    ["SLACK_SIGNING_SECRET", config.slackSigningSecret],
  ].filter(([, value]) => !value);

  if (config.stateDriver !== "redis") {
    missing.push(["CHATSDK_STATE_DRIVER", ""]);
  }

  if (!config.slackBotToken && (!config.slackClientId || !config.slackClientSecret)) {
    missing.push(["SLACK_BOT_TOKEN or SLACK_CLIENT_ID/SLACK_CLIENT_SECRET", ""]);
  }

  if (missing.length > 0) {
    throw new Error(
      `Missing production bot configuration: ${missing
        .map(([key]) => key)
        .join(", ")}`,
    );
  }
};
