import { createOpenAI } from "@ai-sdk/openai";
import { streamObject } from "ai";
import { withTracing } from "@posthog/ai";
import { auth } from "@/auth";
import posthogServer from "@/app/posthog-server";
import { keyResultGenerationSchema } from "@/modules/objectives/schemas/key-result-generation";
import type { Objective, KeyResult } from "@/modules/objectives/types";

export const maxDuration = 30;

type RequestContext = {
  objective: Objective;
  keyResults: KeyResult[];
};

export async function POST(req: Request) {
  const context = await req.json();
  const { objective, keyResults } = context as RequestContext;
  const session = await auth();
  if (!session?.user) {
    return new Response("Unauthorized", { status: 401 });
  }

  const phClient = posthogServer();

  const openaiClient = createOpenAI({
    // eslint-disable-next-line turbo/no-undeclared-env-vars -- this is ok
    apiKey: process.env.OPENAI_API_KEY
  });

  const model = withTracing(openaiClient("gpt-4.1-nano"), phClient, {
    posthogDistinctId: session.user.email ?? "",
    posthogProperties: {
      action: "generate_key_results",
    },
  });

  try {
    const improvedPrompt = `You are an expert in OKR (Objectives and Key Results) methodology. Analyze this objective and suggest 1-5 well-structured key results:

    Context - Objective:
     - Name: ${objective.name}
     - Description: ${objective.description}
     - Current Key Results: ${keyResults.map((kr) => kr.name).join("\n ") || "None"} don't suggest key results that are already in the list or similar to existing key results

    ## Guidelines
    - Each key result should be specific, measurable, and directly contribute to the objective
    - Follow SMART criteria: Specific, Measurable, Achievable, Relevant, Time-bound
    - Only suggest key results if the objective is actionable and has a clear goal

    ## Key Result Requirements
    - **Name**: Clear, specific description of the measurable outcome
    - **Measurement Type**: Choose appropriate type:
      - "number": For countable metrics (users, revenue, etc.)
      - "percentage": For rates and ratios (0-100%)
      - "boolean": For binary outcomes (complete/incomplete)
    - **Start Value**: Realistic baseline value
    - **Target Value**: Ambitious but achievable target

    ## Examples
    - Number: "Increase monthly active users from 10,000 to 15,000"
    - Percentage: "Achieve customer satisfaction score from 85% to 95%"
    - Boolean: "Launch new mobile app successfully"

    ## Quality Criteria
    - Specificity: Avoid vague terms like "improve" or "better"
    - Measurability: Clear what success looks like
    - Relevance: Directly supports the objective
    - Achievability: Realistic given resources and timeframe

    ## Avoid
    - Key results too broad or vague
    - Duplicating existing key results this this very important, also don't suggest key results similar to existing key results
    - Unrealistic targets
    - Non-measurable outcomes

    Generate 1-5 key results with appropriate measurement types and realistic start/target values that follow these principles.
`;

    const result = streamObject({
      model,
      schema: keyResultGenerationSchema,
      prompt: improvedPrompt,
    });

    return result.toTextStreamResponse();
  } catch (error) {
    throw new Error("Failed to generate key results");
  }
}
