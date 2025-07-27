import { createOpenAI } from "@ai-sdk/openai";
import { streamObject } from "ai";
import { withTracing } from "@posthog/ai";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { substoryGenerationSchema } from "@/modules/stories/schemas";
import type { DetailedStory } from "@/modules/story/types";

export const maxDuration = 30;

export async function POST(req: Request) {
  const context = await req.json();
  const parentStory = context as DetailedStory;
  const session = await auth();
  if (!session?.user) {
    return new Response("Unauthorized", { status: 401 });
  }

  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    // eslint-disable-next-line turbo/no-undeclared-env-vars -- this is ok
    apiKey: process.env.OPENAI_API_KEY,
    compatibility: "strict",
  });

  const model = withTracing(openaiClient("gpt-4.1-nano"), phClient, {
    posthogDistinctId: session.user.email ?? "",
    posthogProperties: {
      action: "generate_substories",
    },
  });

  try {
    const improvedPrompt = `You are an expert in agile project management. Analyze this story and suggest 0-5 well-structured substories:

    Context - Parent Story:
     - Title: ${parentStory.title}
     - Description: ${parentStory.description}
     - Current Sub Stories: ${parentStory.subStories.map((subStory) => subStory.title).join("\n ")}

    ## Guidelines
    - Break down by user value - each substory delivers tangible value
    - Follow INVEST criteria: Independent, Negotiable, Valuable, Estimable, Small, Testable
    - Only suggest substories if the parent story is actionable and has a clear goal, if not return an empty array

    ## Title Requirements
    - Clear and actionable (e.g., "Implement user authentication flow" not "User auth")
    - Specific scope focused on single feature
    - User-focused emphasizing what user gets
    - Measurable with concrete outcomes

    ## Quality Criteria
    - Specificity: Avoid vague terms like "improve"
    - Measurability: Clear what success looks like
    - Actionability: Developer can start working immediately
    - Value-focused: Contributes to parent story's goal

    ## Avoid
    - Substories too large
    - Technical implementation details alone
    - No user value
    - Duplicating existing substories this this very important, also don't suggest substories similar to existing substories

    Generate 0-5 substories with clear, actionable titles that follow these principles.
`;

    const result = streamObject({
      model,
      schema: substoryGenerationSchema,
      prompt: improvedPrompt,
    });

    return result.toTextStreamResponse();
  } catch (error) {
    throw new Error("Failed to generate substories");
  }
}
