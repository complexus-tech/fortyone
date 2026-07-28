import type { FileRejection } from "react-dropzone";

export const MAX_ATTACHMENT_BATCH_FILES = 10;
export const FREE_ATTACHMENT_SIZE_LIMIT = 10 * 1024 * 1024;
export const PAID_ATTACHMENT_SIZE_LIMIT = 25 * 1024 * 1024;

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
    return `${rejection.file.name} is not supported. Upload an image, MP4 video, or PDF.`;
  }
  return rejection.errors.map((error) => error.message).join(" ");
};

export const uploadAttachmentsConcurrently = (
  files: File[],
  upload: (file: File) => Promise<unknown>,
) => Promise.allSettled(files.map((file) => upload(file)));
