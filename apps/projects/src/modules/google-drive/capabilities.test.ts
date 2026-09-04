/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  canMayaReadGoogleDriveFile,
  isTrustedGoogleDriveWebViewLink,
} from "./capabilities";

describe("Google Drive capabilities", () => {
  it.each([
    "application/vnd.google-apps.document",
    "application/vnd.google-apps.presentation",
    "application/vnd.google-apps.spreadsheet",
    "text/csv",
    "text/plain",
  ])("allows Maya to read %s", (mimeType) => {
    expect(canMayaReadGoogleDriveFile(mimeType)).toBe(true);
  });

  it.each(["application/pdf", "image/jpeg", "image/png", "image/webp"])(
    "keeps Ask Maya unavailable for %s",
    (mimeType) => {
      expect(canMayaReadGoogleDriveFile(mimeType)).toBe(false);
    },
  );

  it.each([
    "https://docs.google.com/document/d/file/edit",
    "https://drive.google.com/file/d/file/view",
  ])("accepts a trusted Google file URL: %s", (url) => {
    expect(isTrustedGoogleDriveWebViewLink(url)).toBe(true);
  });

  it.each([
    "http://docs.google.com/document/d/file/edit",
    "https://docs.google.com.evil.example/document/d/file/edit",
    "https://user@drive.google.com/file/d/file/view",
    "not-a-url",
  ])("rejects an untrusted Google file URL: %s", (url) => {
    expect(isTrustedGoogleDriveWebViewLink(url)).toBe(false);
  });
});
