import {
  getStandaloneGoogleDriveURL,
  parseGoogleDriveURL,
} from "./google-drive-url";

describe("parseGoogleDriveURL", () => {
  it.each([
    ["https://docs.google.com/document/d/doc_123/edit", "doc_123", "document"],
    [
      "https://docs.google.com/spreadsheets/u/0/d/sheet-123/edit#gid=7",
      "sheet-123",
      "spreadsheet",
    ],
    [
      "https://docs.google.com/presentation/d/slides_123/edit?resourcekey=key_123",
      "slides_123",
      "presentation",
    ],
    ["https://drive.google.com/file/d/file_123/view", "file_123", "file"],
    [
      "https://drive.google.com/open?id=file-123&resourcekey=key-123",
      "file-123",
      "file",
    ],
  ])("parses a supported Google file URL", (url, fileId, kind) => {
    expect(parseGoogleDriveURL(url)).toEqual(
      expect.objectContaining({ fileId, kind }),
    );
  });

  it.each([
    "http://docs.google.com/document/d/file/edit",
    "https://docs.google.com.evil.test/document/d/file/edit",
    "https://user@docs.google.com/document/d/file/edit",
    "https://docs.google.com:444/document/d/file/edit",
    "https://docs.google.com/document/d/e/published-id/pub",
    "https://drive.google.com/drive/folders/folder-id",
    "https://docs.google.com/forms/d/form-id/edit",
    "https://docs.google.com/document/d/file%2Fother/edit",
  ])("rejects unsupported or untrusted URLs", (url) => {
    expect(parseGoogleDriveURL(url)).toBeNull();
  });

  it("keeps standalone detection strict", () => {
    const url = "https://docs.google.com/spreadsheets/d/sheet_123/edit";
    expect(getStandaloneGoogleDriveURL(`  ${url}\n`)).toBe(url);
    expect(getStandaloneGoogleDriveURL(`Review ${url}`)).toBeNull();
  });
});
