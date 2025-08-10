import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { experimental_transcribe as transcribe } from "ai";
import { openai } from "@ai-sdk/openai";
import { auth } from "@/auth";

export const maxDuration = 30;

export async function POST(req: NextRequest) {
  try {
    const session = await auth();
    if (!session?.user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const formData = await req.formData();
    const audioFile = formData.get("audio") as File | undefined;

    if (!audioFile) {
      return NextResponse.json(
        { error: "No audio file provided" },
        { status: 400 },
      );
    }

    if (!audioFile.type.startsWith("audio/")) {
      return NextResponse.json(
        { error: "Invalid file type. Please provide an audio file." },
        { status: 400 },
      );
    }

    const maxSize = 5 * 1024 * 1024;
    if (audioFile.size > maxSize) {
      return NextResponse.json(
        { error: "File too large. Maximum size is 5MB." },
        { status: 400 },
      );
    }

    const arrayBuffer = await audioFile.arrayBuffer();

    const transcript = await transcribe({
      model: openai.transcription("gpt-4o-mini-transcribe"),
      audio: arrayBuffer,
    });

    return NextResponse.json({
      text: transcript.text,
    });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to transcribe audio" },
      { status: 500 },
    );
  }
}
