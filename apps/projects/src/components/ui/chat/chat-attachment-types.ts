export const WORD_DOCUMENT_MIME_TYPES = {
  doc: "application/msword",
  docx: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
} as const;

export const CHAT_ATTACHMENT_ACCEPT = {
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/webp": [".webp"],
  "image/gif": [".gif"],
  "application/pdf": [".pdf"],
  [WORD_DOCUMENT_MIME_TYPES.doc]: [".doc"],
  [WORD_DOCUMENT_MIME_TYPES.docx]: [".docx"],
};

export const getChatAttachmentMediaType = (file: File) => {
  const extension = file.name.toLowerCase().split(".").pop();
  if (extension === "doc") return WORD_DOCUMENT_MIME_TYPES.doc;
  if (extension === "docx") return WORD_DOCUMENT_MIME_TYPES.docx;
  return file.type;
};
