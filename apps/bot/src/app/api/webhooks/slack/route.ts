import type { NextRequest } from "next/server";
import { bot } from "@/lib/bot";
import { errorMessage, logBotError } from "@/lib/logger";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export const POST = async (request: NextRequest) => {
  try {
    return await bot.webhooks.slack(request);
  } catch (error) {
    logBotError("Unhandled Slack webhook error", {
      error: errorMessage(error),
    });
    return Response.json({ error: "Slack webhook failed" }, { status: 500 });
  }
};
