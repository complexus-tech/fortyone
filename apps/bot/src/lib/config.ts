export interface BotConfig {
  botUserName: string;
  fortyOneApiUrl: string;
  fortyOneServiceToken: string;
  isProduction: boolean;
  openAIKey: string;
  openAIModel: string;
  redisUrl: string;
  slackBotToken: string;
  slackSigningSecret: string;
}

const readEnv = (key: string) => process.env[key]?.trim() ?? "";

export const getBotConfig = (): BotConfig => {
  const isProduction = process.env.NODE_ENV === "production";

  return {
    botUserName: readEnv("SLACK_BOT_USERNAME") || "maya",
    fortyOneApiUrl: readEnv("FORTYONE_API_URL"),
    fortyOneServiceToken: readEnv("FORTYONE_BOT_TOKEN"),
    isProduction,
    openAIKey: readEnv("OPENAI_API_KEY"),
    openAIModel: readEnv("BOT_OPENAI_MODEL") || "gpt-5.4-mini",
    redisUrl: readEnv("REDIS_URL"),
    slackBotToken: readEnv("SLACK_BOT_TOKEN"),
    slackSigningSecret: readEnv("SLACK_SIGNING_SECRET"),
  };
};

export const assertProductionConfig = (config: BotConfig) => {
  if (!config.isProduction) {
    return;
  }

  const missing = [
    ["OPENAI_API_KEY", config.openAIKey],
    ["FORTYONE_API_URL", config.fortyOneApiUrl],
    ["FORTYONE_BOT_TOKEN", config.fortyOneServiceToken],
    ["REDIS_URL", config.redisUrl],
    ["SLACK_SIGNING_SECRET", config.slackSigningSecret],
  ].filter(([, value]) => !value);

  if (missing.length > 0) {
    throw new Error(
      `Missing production bot configuration: ${missing
        .map(([key]) => key)
        .join(", ")}`,
    );
  }
};
