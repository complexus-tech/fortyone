import { openai } from "@ai-sdk/openai";
import { streamText } from "ai";
import type { NextRequest } from "next/server";
import { allTools } from "@/lib/ai/tools";
import { systemPrompt } from "./system";

export async function POST(req: NextRequest) {
  try {
    const { messages } = await req.json();

    const result = streamText({
      model: openai("gpt-4o-mini"),
      messages,
      tools: allTools,
      system: systemPrompt,
    });

    return new Response(result.toDataStream(), {
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
      },
    });
  } catch (error) {
    // Return a fallback response if AI service is unavailable
    const fallbackResponse = {
      id: Date.now().toString(),
      content:
        "I'm having trouble connecting to my AI service right now. You can try asking me to navigate to different parts of the app, manage stories, or get sprint information.",
      role: "assistant",
      createdAt: new Date(),
    };

    return new Response(JSON.stringify(fallbackResponse), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }
}
