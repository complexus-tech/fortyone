const GOOGLE_DRIVE_FILE_ID_PATTERN = /^[A-Za-z0-9_-]+$/;
const GOOGLE_DRIVE_RESOURCE_KEY_PATTERN = /^[A-Za-z0-9_-]+$/;

export type GoogleDriveURL = {
  fileId: string;
  kind: "document" | "file" | "presentation" | "spreadsheet";
  mimeType?: string;
  resourceKey?: string;
  url: string;
};

const documentKinds = {
  document: {
    kind: "document",
    mimeType: "application/vnd.google-apps.document",
  },
  presentation: {
    kind: "presentation",
    mimeType: "application/vnd.google-apps.presentation",
  },
  spreadsheets: {
    kind: "spreadsheet",
    mimeType: "application/vnd.google-apps.spreadsheet",
  },
} as const;

const isDocumentKind = (
  value: string | undefined,
): value is keyof typeof documentKinds =>
  Boolean(value && Object.prototype.hasOwnProperty.call(documentKinds, value));

const validIdentifier = (value: string | null) =>
  Boolean(value && GOOGLE_DRIVE_FILE_ID_PATTERN.test(value));

const getResourceKey = (url: URL) => {
  const resourceKey = url.searchParams.get("resourcekey")?.trim();
  return resourceKey && GOOGLE_DRIVE_RESOURCE_KEY_PATTERN.test(resourceKey)
    ? resourceKey
    : undefined;
};

const parseDocsURL = (url: URL): GoogleDriveURL | null => {
  const segments = url.pathname.split("/").filter(Boolean);
  const documentKindKey = segments[0];
  if (!isDocumentKind(documentKindKey)) return null;
  const documentKind = documentKinds[documentKindKey];

  let cursor = 1;
  if (segments[cursor] === "u" && /^\d+$/.test(segments[cursor + 1] ?? "")) {
    cursor += 2;
  }
  if (segments[cursor] !== "d") return null;

  const fileId = segments[cursor + 1] ?? "";
  // Published `/d/e/{publishedId}` documents are not Drive file identities.
  if (fileId === "e" || !validIdentifier(fileId)) return null;

  return {
    fileId,
    ...documentKind,
    ...(getResourceKey(url) ? { resourceKey: getResourceKey(url) } : {}),
    url: url.toString(),
  };
};

const parseDriveURL = (url: URL): GoogleDriveURL | null => {
  const segments = url.pathname.split("/").filter(Boolean);
  let fileId: string | null = null;

  if (segments[0] === "file" && segments[1] === "d") {
    fileId = segments[2] ?? null;
  } else if (segments.length === 1 && segments[0] === "open") {
    fileId = url.searchParams.get("id");
  }
  if (!validIdentifier(fileId)) return null;

  const resourceKey = getResourceKey(url);
  return {
    fileId: fileId!,
    kind: "file",
    ...(resourceKey ? { resourceKey } : {}),
    url: url.toString(),
  };
};

export const parseGoogleDriveURL = (value: string): GoogleDriveURL | null => {
  const trimmedValue = value.trim();
  let url: URL;
  try {
    url = new URL(trimmedValue);
  } catch {
    return null;
  }

  if (
    url.protocol !== "https:" ||
    url.username ||
    url.password ||
    url.port ||
    /%(?:0[0-9a-f]|1[0-9a-f]|2f|5c|7f)/i.test(url.pathname)
  ) {
    return null;
  }

  if (url.hostname === "docs.google.com") return parseDocsURL(url);
  if (url.hostname === "drive.google.com") return parseDriveURL(url);
  return null;
};

export const getStandaloneGoogleDriveURL = (value: string) => {
  const trimmedValue = value.trim();
  return parseGoogleDriveURL(trimmedValue) ? trimmedValue : null;
};
