/* eslint-disable no-undef -- Jest globals are provided by the Projects test environment. */
import { isFigmaURL } from "./url";

describe("isFigmaURL", () => {
  it.each<string>([
    "https://www.figma.com/design/key/file?node-id=1-2",
    "https://figma.com/file/key/file",
  ])("recognizes Figma URLs", (url) => {
    expect(isFigmaURL(url)).toBe(true);
  });

  it.each<string>([
    "https://figma.com.example.org/design/key/file",
    "https://example.org/?next=https://figma.com/design/key/file",
    "not-a-url",
  ])("rejects non-Figma URLs", (url) => {
    expect(isFigmaURL(url)).toBe(false);
  });
});
