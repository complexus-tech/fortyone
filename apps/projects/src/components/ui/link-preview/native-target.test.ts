/* global describe, expect, it -- Provided by the projects test runner. */
import { parseNativeLink } from "./native-target";

const ID = "8c86611a-a44b-4be1-9bea-b537d52280b4";
const TEAM = "661e24d6-bffb-4c75-9048-d946fe2c43b7";
const context = { workspaceSlug: "acme", origin: "https://cloud.fortyone.app" };

describe("native link routing", () => {
  it.each([
    [`https://cloud.fortyone.app/acme/work/ENG-42`, "story", "ENG-42", "acme"],
    [`https://other.fortyone.app/work/${ID}`, "story", ID, "other"],
    [
      `https://acme.fortyone.app/teams/${TEAM}/objectives/${ID}?view=details#progress`,
      "objective",
      ID,
      "acme",
    ],
    [`https://cloud.fortyone.app/acme/docs/${ID}`, "document", ID, "acme"],
    [`/work/eng-42`, "story", "ENG-42", "acme"],
  ])(
    "recognizes %s without guessing resource IDs",
    (url, kind, id, workspaceSlug) => {
      expect(parseNativeLink(url, context)).toMatchObject({
        kind,
        id,
        workspaceSlug,
      });
    },
  );

  it("accepts only the current loopback origin and supports prefixed workspace routes", () => {
    const local = { workspaceSlug: "acme", origin: "http://127.0.0.1:4177" };
    expect(
      parseNativeLink(`http://127.0.0.1:4177/acme/docs/${ID}`, local),
    ).toMatchObject({ kind: "document", id: ID, workspaceSlug: "acme" });
    expect(
      parseNativeLink("http://127.0.0.1:4177/work/ENG-42", local),
    ).toMatchObject({ kind: "story" });
    expect(
      parseNativeLink(`http://127.0.0.1:3000/acme/docs/${ID}`, local),
    ).toBeNull();
  });

  it.each([
    "https://fortyone.app/",
    "https://www.fortyone.app/pricing",
    "https://docs.fortyone.app/",
    "https://acme.fortyone.app.evil.test/work/ENG-42",
    "https://fortyone-app.test/acme/work/ENG-42",
    // eslint-disable-next-line no-script-url -- Explicitly exercises an unsafe input, never navigation.
    "javascript:alert(1)",
    "https://user:password@acme.fortyone.app/work/ENG-42",
  ])("does not trust %s", (url) => {
    expect(parseNativeLink(url, context)).toBeNull();
  });

  it.each([
    "https://cloud.fortyone.app/acme/settings/private",
    "https://acme.fortyone.app/work/not-a-reference",
    "https://acme.fortyone.app/work/ENG-42/other",
    "https://acme.fortyone.app/work/ENG%2f42",
    "https://acme.fortyone.app/work/ENG%252f42",
    "https://nested.acme.fortyone.app/work/ENG-42",
    "https://api.fortyone.app/private",
    "https://acme.fortyone.app:444/work/ENG-42",
  ])(
    "classifies internal unsupported URL without an OG fallback: %s",
    (url) => {
      expect(parseNativeLink(url, context)).toMatchObject({
        kind: "unsupported",
      });
    },
  );
});
