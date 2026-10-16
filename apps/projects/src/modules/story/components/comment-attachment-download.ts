import type { ApiResponse } from "@/types/api-response";
import type { StoryAttachment } from "../types";

export const resolveCommentAttachmentUrl = async (
  file: StoryAttachment,
  signal?: AbortSignal,
) => {
  // Resolve with app cookies, then download the signed object without sending
  // those credentials to storage or persisting its expiring URL in the comment.
  const resolution = await fetch(file.url, {
    credentials: "include",
    headers: { Accept: "application/json" },
    cache: "no-store",
    ...(signal ? { signal } : {}),
  });
  const response = (await resolution.json()) as ApiResponse<{ url: string }>;
  if (!resolution.ok || !response.data?.url) {
    throw new Error(
      response.error?.message || "This attachment is unavailable.",
    );
  }
  return response.data.url;
};

export const downloadCommentAttachment = async (file: StoryAttachment) => {
  const url = await resolveCommentAttachmentUrl(file);
  const download = await fetch(url, { credentials: "omit" });
  if (!download.ok) throw new Error("The attachment could not be downloaded.");
  const objectURL = URL.createObjectURL(await download.blob());
  const link = document.createElement("a");
  link.href = objectURL;
  link.download = file.filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.setTimeout(() => {
    URL.revokeObjectURL(objectURL);
  }, 0);
};
