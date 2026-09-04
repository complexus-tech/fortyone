export const MAX_FEEDBACK_ATTACHMENTS = 5;

export const FEEDBACK_ATTACHMENT_ACCEPT = [
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
  "video/mp4",
  "application/pdf",
  "application/msword",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-powerpoint",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  "text/plain",
  "text/csv",
].join(",");

export const addUniqueFeedbackAttachments = (
  current: File[],
  incoming: File[],
) => {
  const knownFiles = new Set(
    current.map((file) => `${file.name}:${file.size}:${file.lastModified}`),
  );
  const uniqueFiles = incoming.filter(
    (file) => !knownFiles.has(`${file.name}:${file.size}:${file.lastModified}`),
  );

  return [...current, ...uniqueFiles].slice(0, MAX_FEEDBACK_ATTACHMENTS);
};
