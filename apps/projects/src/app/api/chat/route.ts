import { openai } from "@ai-sdk/openai";
import { streamText } from "ai";
import type { NextRequest } from "next/server";
import { allTools } from "@/lib/ai/tools";

export async function POST(req: NextRequest) {
  try {
    const { messages } = await req.json();

    const result = streamText({
      model: openai("gpt-4o-mini"),
      messages,
      tools: allTools,
      system: `You are Maya, an AI assistant for Complexus, a project management platform. You help users navigate the application, manage stories, get sprint insights, and provide project management assistance.

Available capabilities:
- Navigation: Help users navigate to different sections of the app
- Stories: List, search, create, and manage stories/tasks
- Sprints: Get sprint summaries, burndown charts, velocity data, and planning recommendations

Be helpful, concise, and friendly. When users want to navigate somewhere, use the navigation tool. When they ask about stories or sprints, use the appropriate tools to provide detailed information.

Always provide context and explanations for your responses. If a user asks to navigate, explain why you're taking them there and what they can expect to find.`,
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
