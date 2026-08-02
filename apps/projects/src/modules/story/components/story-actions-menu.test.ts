/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Story actions menu", () => {
  it("uses a naked horizontal actions trigger", () => {
    const source = readSource(
      "src/modules/story/components/story-actions-menu.tsx",
    );

    expect(source).toContain("MoreHorizontalIcon");
    expect(source).not.toContain("MoreVerticalIcon");
    expect(source).toContain('variant="naked"');
    expect(source).not.toContain("buttonClassName");
  });

  it("keeps subscription and lifecycle actions together before delete", () => {
    const source = readSource(
      "src/modules/story/components/story-actions-menu.tsx",
    );
    const menu = source.slice(
      source.indexOf("<Menu.Items"),
      source.indexOf("<ConfirmationDialog"),
    );
    const utilitySeparator = menu.indexOf("<Menu.Separator");
    const subscription = menu.indexOf("subscriptionMutation.mutate");
    const lifecycle = menu.indexOf("{lifecycleMenuItem}");
    const dangerSeparator = menu.indexOf(
      "<Menu.Separator",
      utilitySeparator + 1,
    );
    const deletion = menu.indexOf("openDialogAfterMenuClose(setIsDeleteOpen)");

    expect(utilitySeparator).toBeLessThan(subscription);
    expect(subscription).toBeLessThan(lifecycle);
    expect(lifecycle).toBeLessThan(dangerSeparator);
    expect(dangerSeparator).toBeLessThan(deletion);
  });

  it("offers focused property and utility actions", () => {
    const source = readSource(
      "src/modules/story/components/story-actions-menu.tsx",
    );
    const objectiveMenu = readSource(
      "src/components/ui/story/objective-key-result-menu.tsx",
    );

    expect(source).toContain("StoryStatusSubMenu");
    expect(source).toContain("StoryPrioritySubMenu");
    expect(source).not.toContain("Change status");
    expect(source).not.toContain("Change priority");
    expect(objectiveMenu).not.toContain(
      '{getTermDisplay("objectiveTerm", { capitalize: true })} and',
    );
    expect(source).toContain("Duplicate");
    expect(source).toContain("Open in new tab");
    expect(source).toContain("ExternalLinkIcon");
    expect(source).not.toContain("NewTabIcon");
    expect(source).toContain("Copy link");
  });

  it("exposes the same story properties from the list context menu", () => {
    const contextMenu = readSource("src/components/ui/story/context-menu.tsx");
    const properties = readSource(
      "src/components/ui/story/story-property-context-menu.tsx",
    );

    expect(contextMenu).toContain("StoryPropertyContextMenu");
    expect(properties).toContain('label="Status"');
    expect(properties).toContain('label="Priority"');
    expect(properties).toContain("ObjectiveKeyResultContextSubMenu");
    expect(properties).toContain("STORY_PRIORITIES.map");
    expect(contextMenu).toContain("ExternalLinkIcon");
    expect(contextMenu).not.toContain("NewTabIcon");
  });

  it("shows the menu once per surface and last on the full page", () => {
    const header = readSource(
      "src/modules/story/components/options-header.tsx",
    );
    const dialog = readSource("src/components/ui/story-dialog/index.tsx");

    expect(header).toContain("{!isDialog ? (");
    expect(header.indexOf("<GitIcon")).toBeLessThan(
      header.indexOf("<StoryActionsMenu"),
    );
    expect(dialog).toContain(
      '<StoryActionsMenu align="start" storyId={storyId} />',
    );
    expect(dialog).not.toContain("buttonClassName");
  });
});
