import type { NextRequest } from "next/server";

export function register(): void {
  // No-op for initialization
}

type RequestContext = Record<string, unknown>;

export const onRequestError = async (
  err: Error,
  request: NextRequest,
  _context: RequestContext,
): Promise<void> => {
  // eslint-disable-next-line turbo/no-undeclared-env-vars -- ok for this
  if (process.env.NEXT_RUNTIME === "nodejs") {
    // eslint-disable-next-line @typescript-eslint/no-var-requires -- Using require for conditional import in Node.js runtime
    const { getPostHogServer } = require("./app/posthog-server");
    const posthog = await getPostHogServer();

    let distinctId: string | null = null;
    if (request.headers.get("cookie")) {
      const cookieString = request.headers.get("cookie") || "";
      // eslint-disable-next-line prefer-named-capture-group -- ok for this
      const postHogCookieMatch = /ph_phc_.*?_posthog=([^;]+)/.exec(
        cookieString,
      );

      const cookieValue = postHogCookieMatch?.[1];
      if (cookieValue) {
        try {
          const decodedCookie = decodeURIComponent(cookieValue);
          const postHogData = JSON.parse(decodedCookie);
          distinctId = postHogData.distinct_id;
        } catch (e) {
          // ignore
        }
      }
    }

    await posthog.captureException(err, distinctId || undefined);
  }
};
