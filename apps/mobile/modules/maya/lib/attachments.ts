import type { FileUIPart } from "ai";

export const MAYA_ATTACHMENT_LIMITS = {
  count: 5,
  fileBytes: 5 * 1024 * 1024,
  totalBytes: 8 * 1024 * 1024,
} as const;

export const MAYA_ATTACHMENT_MEDIA_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
  "application/pdf",
] as const;

export type MayaAttachment = {
  id: string;
  uri: string;
  name: string;
  mediaType: (typeof MAYA_ATTACHMENT_MEDIA_TYPES)[number];
  size: number;
};

const mediaTypes = new Set<string>(MAYA_ATTACHMENT_MEDIA_TYPES);

export const validateMayaAttachmentSizes = (
  files: readonly { size: number }[],
) => {
  if (files.length > MAYA_ATTACHMENT_LIMITS.count)
    throw new Error("Attach up to 5 images or PDFs per message.");
  let total = 0;
  for (const file of files) {
    if (!Number.isSafeInteger(file.size) || file.size <= 0)
      throw new Error("An attachment is empty or its size could not be read.");
    if (file.size > MAYA_ATTACHMENT_LIMITS.fileBytes)
      throw new Error("Each attachment must be 5 MB or smaller.");
    total += file.size;
  }
  if (total > MAYA_ATTACHMENT_LIMITS.totalBytes)
    throw new Error("Attachments must total 8 MB or less per message.");
};

/** Checks encoded length before scanning the payload; never decodes a full file. */
export const validateMayaFileParts = (files: readonly FileUIPart[]): void => {
  if (files.length > MAYA_ATTACHMENT_LIMITS.count)
    throw new Error("Attach up to 5 images or PDFs per message.");
  const sizes = files.map((file) => {
    if (file.type !== "file" || !mediaTypes.has(file.mediaType))
      throw new Error("Attach JPEG, PNG, WebP, GIF images or PDF documents.");
    const prefix = `data:${file.mediaType};base64,`;
    if (typeof file.url !== "string" || !file.url.startsWith(prefix))
      throw new Error("An attachment could not be prepared. Select it again.");
    const payload = file.url.slice(prefix.length);
    if (payload.length > Math.ceil(MAYA_ATTACHMENT_LIMITS.fileBytes / 3) * 4)
      throw new Error("Each attachment must be 5 MB or smaller.");
    if (
      !payload.length ||
      payload.length % 4 !== 0 ||
      !/^[A-Za-z0-9+/]*={0,2}$/.test(payload)
    )
      throw new Error("An attachment could not be read. Select it again.");
    const padding = payload.endsWith("==") ? 2 : payload.endsWith("=") ? 1 : 0;
    return { size: (payload.length / 4) * 3 - padding };
  });
  validateMayaAttachmentSizes(sizes);
};

/** The filename/provider MIME is advisory; only supported file signatures pass. */
export const getMayaAttachmentMediaType = (
  bytes: Uint8Array,
): MayaAttachment["mediaType"] => {
  const startsWith = (signature: number[], offset = 0) =>
    signature.every((byte, index) => bytes[offset + index] === byte);
  if (startsWith([0xff, 0xd8, 0xff])) return "image/jpeg";
  if (startsWith([137, 80, 78, 71, 13, 10, 26, 10])) return "image/png";
  if (
    startsWith([71, 73, 70, 56]) &&
    (startsWith([55, 97], 4) || startsWith([57, 97], 4))
  )
    return "image/gif";
  if (startsWith([82, 73, 70, 70]) && startsWith([87, 69, 66, 80], 8))
    return "image/webp";
  if (startsWith([37, 80, 68, 70, 45])) return "application/pdf";
  throw new Error("Attach JPEG, PNG, WebP, GIF images or PDF documents.");
};

type AttachmentSnapshot = {
  attachments: MayaAttachment[];
  picking: boolean;
  preparing: boolean;
};

/** Owns asynchronous picker/read admission independently of React rerenders. */
export const createMayaAttachmentQueue = (options: {
  isCurrent: () => boolean;
  pick: () => Promise<MayaAttachment[]>;
  read: (
    attachment: MayaAttachment,
  ) => Promise<{ size: number; base64: string }>;
  release: (attachments: readonly MayaAttachment[]) => void;
}) => {
  let snapshot: AttachmentSnapshot = {
    attachments: [],
    picking: false,
    preparing: false,
  };
  const listeners = new Set<() => void>();
  let revision = 0;
  let focused = true;
  const publish = (change: Partial<AttachmentSnapshot>) => {
    snapshot = { ...snapshot, ...change };
    listeners.forEach((listener) => listener());
  };
  const assertCurrent = (version: number) => {
    if (!focused || !options.isCurrent() || version !== revision)
      throw new Error(
        "The conversation changed. Select the attachments again.",
      );
  };
  const clear = () => {
    revision++;
    options.release(snapshot.attachments);
    publish({ attachments: [] });
  };
  return {
    getSnapshot: () => snapshot,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    focus: () => {
      focused = true;
    },
    blur: () => {
      focused = false;
      clear();
    },
    background: () => {
      // System pickers can temporarily deactivate the app without leaving Maya.
      if (!snapshot.picking) clear();
    },
    clear,
    remove: (id: string) => {
      revision++;
      options.release(snapshot.attachments.filter((file) => file.id === id));
      publish({
        attachments: snapshot.attachments.filter((file) => file.id !== id),
      });
    },
    pick: async () => {
      const version = revision;
      assertCurrent(version);
      if (snapshot.picking || snapshot.preparing)
        throw new Error("Wait for the attachments to finish loading.");
      if (snapshot.attachments.length >= MAYA_ATTACHMENT_LIMITS.count)
        throw new Error("Attach up to 5 images or PDFs per message.");
      publish({ picking: true });
      let picked: MayaAttachment[] = [];
      try {
        picked = await options.pick();
        assertCurrent(version);
        const attachments = [...snapshot.attachments, ...picked];
        validateMayaAttachmentSizes(attachments);
        publish({ attachments });
      } catch (error) {
        options.release(picked);
        throw error;
      } finally {
        publish({ picking: false });
      }
    },
    prepare: async (): Promise<FileUIPart[]> => {
      const version = revision;
      assertCurrent(version);
      if (snapshot.picking || snapshot.preparing)
        throw new Error("Wait for the attachments to finish loading.");
      validateMayaAttachmentSizes(snapshot.attachments);
      const attachments = snapshot.attachments;
      publish({ preparing: true });
      try {
        const parts: FileUIPart[] = [];
        // Serial conversion bounds simultaneous native buffers and base64 copies.
        for (const attachment of attachments) {
          assertCurrent(version);
          const { size, base64 } = await options.read(attachment);
          assertCurrent(version);
          if (size !== attachment.size)
            throw new Error(
              "An attachment changed. Remove it and select it again.",
            );
          parts.push({
            type: "file",
            filename: attachment.name,
            mediaType: attachment.mediaType,
            url: `data:${attachment.mediaType};base64,${base64}`,
          });
        }
        validateMayaFileParts(parts);
        return parts;
      } finally {
        publish({ preparing: false });
      }
    },
  };
};
