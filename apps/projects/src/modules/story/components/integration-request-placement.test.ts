/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

describe("Story Slack thread placement", () => {
  it("includes linked Slack threads in the banner stack above the title", () => {
    const source = readFileSync(
      join(process.cwd(), "src/modules/story/components/main-details.tsx"),
      "utf8",
    );

    const bannerStackIndex = source.indexOf("<StoryBanners story={data!} />");
    const titleIndex = source.indexOf("<TextEditor\n          asTitle");

    expect(bannerStackIndex).toBeGreaterThan(-1);
    expect(titleIndex).toBeGreaterThan(bannerStackIndex);
    expect(source).not.toContain("<IntegrationRequestSection.Banner");
  });
});
