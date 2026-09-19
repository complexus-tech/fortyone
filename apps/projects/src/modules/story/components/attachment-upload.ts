import type { FileRejection } from "react-dropzone";

export const MAX_ATTACHMENT_BATCH_FILES = 10;
export const FREE_ATTACHMENT_SIZE_LIMIT = 10 * 1024 * 1024;
export const PAID_ATTACHMENT_SIZE_LIMIT = 25 * 1024 * 1024;

export const STORY_ATTACHMENT_ACCEPT = {
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/gif": [".gif"],
  "image/webp": [".webp"],
  "video/mp4": [".mp4"],
  "application/pdf": [".pdf"],
  "application/msword": [".doc"],
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [
    ".docx",
  ],
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
    ".xlsx",
  ],
  "application/vnd.ms-powerpoint": [".ppt"],
  "application/vnd.openxmlformats-officedocument.presentationml.presentation": [
    ".pptx",
  ],
  "text/plain": [".txt"],
  "text/csv": [".csv"],
};

const formatFileSize = (bytes: number) =>
  `${(bytes / (1024 * 1024)).toFixed(bytes >= 10 * 1024 * 1024 ? 0 : 1)} MB`;

export const getAttachmentRejectionMessage = (
  rejection: FileRejection,
  maxFileSize: number,
) => {
  const codes = new Set(rejection.errors.map((error) => error.code));
  if (codes.has("file-too-large")) {
    return `${rejection.file.name} is ${formatFileSize(rejection.file.size)}. The maximum file size is ${formatFileSize(maxFileSize)}.`;
  }
  if (codes.has("file-invalid-type")) {
    return `${rejection.file.name} is not supported. Upload an image, MP4 video, PDF, Word, Excel, PowerPoint, text, or CSV file.`;
  }
  return rejection.errors.map((error) => error.message).join(" ");
};

export const uploadAttachmentsConcurrently = (
  files: File[],
  upload: (file: File) => Promise<unknown>,
) => Promise.allSettled(files.map((file) => upload(file)));
