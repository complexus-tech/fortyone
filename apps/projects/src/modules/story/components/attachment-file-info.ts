import type { StoryAttachment } from "../types";

export const formatAttachmentSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
};

export const getDocumentTypeLabel = (mimeType: string) => {
  if (mimeType === "application/pdf") return "PDF file";
  if (
    mimeType === "application/msword" ||
    mimeType.includes("wordprocessingml")
  )
    return "Word document";
  if (
    mimeType === "application/vnd.ms-excel" ||
    mimeType.includes("spreadsheetml")
  )
    return "Excel spreadsheet";
  if (
    mimeType === "application/vnd.ms-powerpoint" ||
    mimeType.includes("presentationml")
  )
    return "PowerPoint presentation";
  if (mimeType === "text/csv") return "CSV file";
  if (mimeType === "text/plain") return "Text file";
  return "Document";
};

export const getAttachmentTypeLabel = ({
  mimeType,
  filename,
}: Pick<StoryAttachment, "mimeType" | "filename">) => {
  const extension = filename.split(".").at(-1)?.toUpperCase();
  if (mimeType.startsWith("image/")) return `${extension || "Image"} image`;
  if (mimeType.startsWith("video/")) return `${extension || "Video"} video`;
  const documentType = getDocumentTypeLabel(mimeType);
  return documentType === "Document" && extension
    ? `${extension} file`
    : documentType;
};

export const isTextAttachment = ({
  mimeType,
  filename,
}: Pick<StoryAttachment, "mimeType" | "filename">) => {
  const type = mimeType.split(";")[0].trim().toLowerCase();
  if (type === "text/plain" || type === "text/csv") return true;
  // Only fall back to the filename when stored uploads lack a known MIME type.
  if (type && type !== "application/octet-stream") return false;
  return /\.(?:txt|csv)$/i.test(filename);
};
