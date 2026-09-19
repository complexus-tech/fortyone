import "server-only";

import { OpenTelemetry } from "@ai-sdk/otel";
import { PostHogSpanProcessor } from "@posthog/ai/otel";
import { NodeSDK } from "@opentelemetry/sdk-node";
import type { Attributes } from "@opentelemetry/api";
import type { TelemetryOptions } from "ai";

const AI_TELEMETRY_STATE_KEY = Symbol.for("fortyone.ai-telemetry-state");

type AiTelemetryState = {
  processor?: PostHogSpanProcessor;
  registered: boolean;
  sdk?: NodeSDK;
};

type GlobalWithAiTelemetry = typeof globalThis & {
  [AI_TELEMETRY_STATE_KEY]?: AiTelemetryState;
};

const getState = () => {
  const globalState = globalThis as GlobalWithAiTelemetry;
  globalState[AI_TELEMETRY_STATE_KEY] ??= { registered: false };
  return globalState[AI_TELEMETRY_STATE_KEY];
};

export const registerAiTelemetry = () => {
  const state = getState();
  if (state.registered) return;

  state.registered = true;
  const projectToken = process.env.NEXT_PUBLIC_POSTHOG_KEY;
  if (!projectToken) return;

  const processor = new PostHogSpanProcessor({
    projectToken,
    host: process.env.NEXT_PUBLIC_POSTHOG_HOST,
  });
  const sdk = new NodeSDK({ spanProcessors: [processor] });
  sdk.start();

  state.processor = processor;
  state.sdk = sdk;
};

type PostHogAiTelemetryProperties = Record<
  string,
  boolean | number | string | undefined
>;

class FlushableOpenTelemetry extends OpenTelemetry {
  override async onAbort(
    event: Parameters<OpenTelemetry["onAbort"]>[0],
  ): Promise<void> {
    super.onAbort(event);
    await flushAiTelemetry();
  }

  override async onEnd(
    event: Parameters<OpenTelemetry["onEnd"]>[0],
  ): Promise<void> {
    super.onEnd(event);
    await flushAiTelemetry();
  }

  override async onError(error: unknown): Promise<void> {
    super.onError(error);
    await flushAiTelemetry();
  }
}

export const createPostHogAiTelemetry = ({
  distinctId,
  functionId,
  privacyMode,
  properties = {},
}: {
  distinctId: string;
  functionId: string;
  privacyMode: boolean;
  properties?: PostHogAiTelemetryProperties;
}): TelemetryOptions => {
  const attributes: Attributes = {
    "ai.settings.context.posthog_distinct_id": distinctId,
  };
  for (const [key, value] of Object.entries(properties)) {
    if (value !== undefined) {
      attributes[`ai.settings.context.${key}`] = value;
    }
  }

  return {
    functionId,
    integrations: new FlushableOpenTelemetry({
      enrichSpan: () => attributes,
    }),
    recordInputs: !privacyMode,
    recordOutputs: !privacyMode,
  };
};

/**
 * PostHog batches AI spans. Serverless request handlers must drain that batch
 * before their lifecycle ends or completed generations can be dropped.
 */
export const flushAiTelemetry = async () => {
  try {
    await getState().processor?.forceFlush();
  } catch (error) {
    // Telemetry delivery must never turn a successful product request into a
    // failure. Keep the diagnostic payload-free because spans may be private.
    // eslint-disable-next-line no-console -- Best-effort observability diagnostic.
    console.warn("[ai/telemetry] Failed to flush PostHog AI spans", error);
  }
};
