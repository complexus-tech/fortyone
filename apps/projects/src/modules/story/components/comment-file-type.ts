export type CommentFileType =
  | "word"
  | "excel"
  | "powerpoint"
  | "pdf"
  | "csv"
  | "text"
  | "video"
  | "image"
  | "audio"
  | "archive"
  | "code"
  | "file";

const MIME_TYPES: Readonly<Partial<Record<string, CommentFileType>>> = {
  "application/msword": "word",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
    "word",
  "application/vnd.ms-excel": "excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "excel",
  "application/vnd.ms-powerpoint": "powerpoint",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation":
    "powerpoint",
  "application/pdf": "pdf",
  "text/csv": "csv",
  "text/plain": "text",
  "video/mp4": "video",
  "image/jpeg": "image",
  "image/png": "image",
  "image/gif": "image",
  "image/webp": "image",
  "application/json": "code",
  "application/xml": "code",
  "text/xml": "code",
  "application/x-rar": "archive",
  "application/vnd.rar": "archive",
  "application/x-7z-compressed": "archive",
  "application/gzip": "archive",
  "application/x-tar": "archive",
};

const EXTENSIONS: Readonly<Partial<Record<string, CommentFileType>>> = {
  doc: "word",
  docx: "word",
  xls: "excel",
  xlsx: "excel",
  ppt: "powerpoint",
  pptx: "powerpoint",
  pdf: "pdf",
  csv: "csv",
  txt: "text",
  mp4: "video",
  jpg: "image",
  jpeg: "image",
  png: "image",
  gif: "image",
  webp: "image",
  mp3: "audio",
  m4a: "audio",
  wav: "audio",
  ogg: "audio",
  zip: "archive",
  rar: "archive",
  "7z": "archive",
  gz: "archive",
  json: "code",
  xml: "code",
};

export const getCommentFileType = (
  mimeType: string,
  filename?: string,
): CommentFileType => {
  const normalizedMimeType = mimeType.split(";")[0].trim().toLowerCase();
  const extension = filename?.includes(".")
    ? filename.split(".").at(-1)?.trim().toLowerCase() ?? ""
    : "";

  if (normalizedMimeType === "text/plain" && extension === "csv") return "csv";
  if (normalizedMimeType.startsWith("audio/")) return "audio";

  // Prefer the server's recognized content type over a misleading extension.
  const detectedType = MIME_TYPES[normalizedMimeType];
  if (detectedType) return detectedType;

  // Older uploads may have stored a generic MIME type for an Office container.
  if (
    !normalizedMimeType ||
    normalizedMimeType === "application/octet-stream" ||
    normalizedMimeType === "application/zip"
  ) {
    return (
      EXTENSIONS[extension] ??
      (normalizedMimeType === "application/zip" ? "archive" : "file")
    );
  }
  return "file";
};
