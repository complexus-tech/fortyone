/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { publicPortalFixture } from "./fixtures";
import {
  buildPublicPortalMetadata,
  getPublicPortalCanonicalUrl,
} from "./metadata";

describe("public portal metadata", () => {
  it("builds organization-specific search and social metadata", () => {
    const canonicalUrl = new URL("https://city-roads.fortyone.app/feedback");
    const metadata = buildPublicPortalMetadata(
      publicPortalFixture,
      canonicalUrl,
    );

    expect(metadata.title).toBe(
      "Send feedback to City Roads Program | FortyOne",
    );
    expect(metadata.description).toContain(
      "Send requests and feedback directly to City Roads Program",
    );
    expect(metadata.alternates?.canonical).toEqual(canonicalUrl);
    expect(metadata.openGraph).toEqual(
      expect.objectContaining({
        title: "Send feedback to City Roads Program | FortyOne",
        type: "website",
        url: canonicalUrl,
      }),
    );
    expect(metadata.robots).toEqual({ follow: true, index: true });
    expect(metadata.icons).toEqual({
      apple: "/complexus.png",
      icon: "/complexus.png",
      shortcut: "/complexus.png",
    });
  });

  it("uses the organization logo for portal icons when one is available", () => {
    const logoUrl = "https://cdn.example.com/city-roads-logo.png";
    const metadata = buildPublicPortalMetadata(
      {
        ...publicPortalFixture,
        workspace: {
          ...publicPortalFixture.workspace,
          avatarUrl: logoUrl,
        },
      },
      new URL("https://city-roads.fortyone.app/feedback"),
    );

    expect(metadata.icons).toEqual({
      apple: logoUrl,
      icon: logoUrl,
      shortcut: logoUrl,
    });
  });

  it("falls back to the peach portal icon when no logo is available", () => {
    const fallbackUrl = new URL(
      "https://city-roads.fortyone.app/api/public-portal/favicon",
    );
    const metadata = buildPublicPortalMetadata(
      {
        ...publicPortalFixture,
        workspace: {
          ...publicPortalFixture.workspace,
          avatarUrl: null,
        },
      },
      new URL("https://city-roads.fortyone.app/feedback"),
    );

    expect(metadata.icons).toEqual({
      apple: fallbackUrl,
      icon: fallbackUrl,
      shortcut: fallbackUrl,
    });
  });

  it("uses the clean feedback path for workspace subdomains", () => {
    expect(
      getPublicPortalCanonicalUrl({
        forwardedHost: "art-circles.fortyone.app",
        forwardedProtocol: "https",
        portalSlug: "art-circles",
      }).toString(),
    ).toBe("https://art-circles.fortyone.app/feedback");
  });

  it("keeps the portal path for local and shared hosts", () => {
    expect(
      getPublicPortalCanonicalUrl({
        host: "localhost:3000",
        portalSlug: "art-circles",
      }).toString(),
    ).toBe("http://localhost:3000/portal/art-circles/feedback");
  });
});
