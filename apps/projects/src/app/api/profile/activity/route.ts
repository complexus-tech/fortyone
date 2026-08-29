import { ApiError } from "api-client";
import { NextResponse, type NextRequest } from "next/server";
import {
  getFeedbackProfileActivity,
  type FeedbackProfileActivityType,
} from "@/modules/public-portal/profile-activity";

const parsePositiveInteger = (value: string | null, fallback: number) => {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
};

export const GET = async (request: NextRequest) => {
  const requestedType = request.nextUrl.searchParams.get("type");
  const type: FeedbackProfileActivityType =
    requestedType === "comment" ? "comment" : "feedback";
  const page = parsePositiveInteger(
    request.nextUrl.searchParams.get("page"),
    1,
  );
  const pageSize = parsePositiveInteger(
    request.nextUrl.searchParams.get("pageSize"),
    20,
  );

  try {
    const activity = await getFeedbackProfileActivity(type, page, pageSize);
    return NextResponse.json({ data: activity });
  } catch (error) {
    if (error instanceof ApiError) {
      return NextResponse.json(error.data, { status: error.status });
    }
    throw error;
  }
};
