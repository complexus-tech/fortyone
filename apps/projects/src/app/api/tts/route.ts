/* eslint-disable turbo/no-undeclared-env-vars -- ok for now */
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import OpenAI from "openai";
import { auth } from "@/auth";

export const maxDuration = 30;

type TTSRequest = {
  text: string;
  voice?: "alloy" | "echo" | "fable" | "onyx" | "nova" | "shimmer";
  speed?: number;
};

export async function POST(req: NextRequest) {
  try {
    const session = await auth();
    if (!session?.user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const { text, voice = "alloy", speed = 1.0 }: TTSRequest = await req.json();

    if (!text || text.trim().length === 0) {
      return NextResponse.json({ error: "No text provided" }, { status: 400 });
    }

    if (text.length > 4096) {
      return NextResponse.json(
        { error: "Text too long. Maximum length is 4096 characters." },
        { status: 400 },
      );
    }

    const openai = new OpenAI({
      apiKey: process.env.OPENAI_API_KEY,
    });

    const mp3 = await openai.audio.speech.create({
      model: "tts-1",
      voice,
      input: text,
      speed,
    });

    const buffer = Buffer.from(await mp3.arrayBuffer());

    return new NextResponse(buffer, {
      headers: {
        "Content-Type": "audio/mpeg",
        "Content-Length": buffer.length.toString(),
      },
    });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to generate speech" },
      { status: 500 },
    );
  }
}
