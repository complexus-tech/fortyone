/* eslint-disable turbo/no-undeclared-env-vars -- NEXT_RUNTIME is provided by Next.js. */

export async function register() {
  if (process.env.NEXT_RUNTIME !== "nodejs") return;

  const { registerAiTelemetry } = await import("./lib/ai/telemetry");
  registerAiTelemetry();
}
