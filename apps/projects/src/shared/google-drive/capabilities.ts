const MAYA_READABLE_GOOGLE_DRIVE_MIME_TYPES = new Set([
  "application/vnd.google-apps.document",
  "application/vnd.google-apps.presentation",
  "application/vnd.google-apps.spreadsheet",
  "text/csv",
  "text/plain",
]);

const GOOGLE_FILE_HOSTS = new Set(["docs.google.com", "drive.google.com"]);

export const canMayaReadGoogleDriveFile = (mimeType: string) =>
  MAYA_READABLE_GOOGLE_DRIVE_MIME_TYPES.has(mimeType);

export const isTrustedGoogleDriveWebViewLink = (value: string) => {
  try {
    const url = new URL(value);
    return (
      url.protocol === "https:" &&
      !url.username &&
      !url.password &&
      !url.port &&
      GOOGLE_FILE_HOSTS.has(url.hostname)
    );
  } catch {
    return false;
  }
};
