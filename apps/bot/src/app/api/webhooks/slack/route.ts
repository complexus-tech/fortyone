import type { NextRequest } from "next/server";
import { bot } from "@/lib/bot";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export const POST = (request: NextRequest) => bot.webhooks.slack(request);
