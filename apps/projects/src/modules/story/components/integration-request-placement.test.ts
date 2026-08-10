/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

describe("Story Slack thread placement", () => {
  it("renders the linked Slack thread immediately before the activity feed", () => {
    const source = readFileSync(
      join(process.cwd(), "src/modules/story/components/main-details.tsx"),
      "utf8",
    );

    const slackThreadIndex = source.indexOf(
      "<IntegrationRequestSection.Banner storyId={storyId} />",
    );
    const activityFeedIndex = source.indexOf(
      "<Activities isDialog={isDialog} storyId={storyId} teamId={teamId} />",
    );
    const titleIndex = source.indexOf("<TextEditor\n          asTitle");

    expect(slackThreadIndex).toBeGreaterThan(titleIndex);
    expect(activityFeedIndex).toBeGreaterThan(slackThreadIndex);
    expect(source.slice(slackThreadIndex, activityFeedIndex).trim()).toBe(
      "<IntegrationRequestSection.Banner storyId={storyId} />",
    );
  });
});
