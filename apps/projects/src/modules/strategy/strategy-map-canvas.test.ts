/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("StrategyMapCanvas", () => {
  it("keeps objective alignment discoverable", () => {
    const source = readSource("src/modules/strategy/strategy-map-canvas.tsx");
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );

    expect(source).toContain("getPillarAtPointer");
    expect(source).toContain("onPointerDown");
    expect(source).toContain("Right-click for actions");
    expect(cardSource).toContain("Align to pillar");
    expect(cardSource).toContain("Remove pillar alignment");
  });

  it("keeps strategy controls in the main header", () => {
    const pageSource = readSource("src/modules/strategy/index.tsx");

    expect(pageSource).toContain('aria-label="Zoom out"');
    expect(pageSource).toContain('aria-label="Zoom in"');
    expect(pageSource).not.toContain("Strategy hierarchy");
    expect(pageSource).not.toContain("Show unaligned");
    expect(pageSource).not.toContain("All teams");
    expect(pageSource).not.toContain("getFullYear");
  });

  it("renders the hierarchy as cards with contextual objective actions", () => {
    const source = readSource("src/modules/strategy/strategy-map-canvas.tsx");
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const navigationSource = readSource(
      "src/components/shared/sidebar/navigation.tsx",
    );

    expect(source).toContain("CanvasConnections");
    expect(source).toContain("UltimateGoalNodeCard");
    expect(source).toContain("PillarNodeCard");
    expect(source).toContain("ObjectiveNodeCard");
    expect(cardSource).toContain("ObjectiveKeyResults");
    expect(cardSource).toContain("OKRIcon");
    expect(cardSource).toContain("CircleProgressBar");
    expect(cardSource).toContain("ObjectiveStatusesMenu");
    expect(cardSource).toContain("PrioritiesMenu");
    expect(cardSource).toContain("AssigneesMenu");
    expect(cardSource).toContain("line-clamp-2");
    expect(cardSource).toContain("Set status");
    expect(navigationSource).toContain("StrategyIcon");
  });

  it("uses the dotted freeform canvas and persistent node positions", () => {
    const source = readSource("src/modules/strategy/strategy-map-canvas.tsx");

    expect(source).toContain("radial-gradient");
    expect(source).toContain("strategy-map-layout:v");
    expect(source).toContain("window.localStorage.setItem");
    expect(source).toContain('vectorEffect="non-scaling-stroke"');
  });

  it("keeps strategy editor fields and destructive pillar actions polished", () => {
    const editorSource = readSource(
      "src/modules/strategy/strategy-editor-dialog.tsx",
    );
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const pageSource = readSource("src/modules/strategy/index.tsx");

    expect(editorSource).toContain("bg-surface-muted/80");
    expect(editorSource).toContain("bg-surface-muted/60");
    expect(cardSource).toContain("text-danger");
    expect(pageSource).toContain("Delete strategic pillar?");
    expect(pageSource).toContain('color="danger"');
  });

  it("keeps objective badges and resting canvas surfaces visually aligned", () => {
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const canvasSource = readSource(
      "src/modules/strategy/strategy-map-canvas.tsx",
    );
    const pageSource = readSource("src/modules/strategy/index.tsx");

    expect(cardSource).toContain("h-[1.85rem] gap-1.5 rounded-xl");
    expect(cardSource).toContain("[&_circle:first-child]:stroke-danger");
    expect(cardSource).not.toContain("min-h-[2.8rem]");
    expect(cardSource).not.toContain("hover:bg-danger/10");
    expect(canvasSource).toContain("dark:bg-surface-elevated/35");
    expect(pageSource).not.toContain("shadow-sm");
  });

  it("uses campaign-style objective references and the shared external-link icon", () => {
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const canvasSource = readSource(
      "src/modules/strategy/strategy-map-canvas.tsx",
    );
    const profileMenuSource = readSource(
      "src/components/shared/sidebar/profile-menu.tsx",
    );
    const externalLinkIconSource = readSource(
      "../../packages/icons/src/external-link.tsx",
    );

    expect(cardSource).toContain("objectiveReference");
    expect(cardSource).toMatch(/`\$\{teamCode\}-\$\{objective\.sequenceId\}`/);
    expect(cardSource).not.toMatch(
      /<NodeEyebrow[^>]*>\s*Objective\s*<\/NodeEyebrow>/,
    );
    expect(canvasSource).toContain("teamCodeById");
    expect(profileMenuSource).toContain("ExternalLinkIcon");
    expect(profileMenuSource).not.toContain("NewTabIcon");
    expect(externalLinkIconSource).toContain("strokeWidth = 2");
  });

  it("opens node details without treating controls or drags as card clicks", () => {
    const pageSource = readSource("src/modules/strategy/index.tsx");
    const canvasSource = readSource(
      "src/modules/strategy/strategy-map-canvas.tsx",
    );
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const detailsSource = readSource(
      "src/modules/strategy/strategy-node-details.tsx",
    );
    const selectedDetailsSource = readSource(
      "src/modules/strategy/strategy-selected-details.tsx",
    );

    expect(pageSource).toContain("StrategySelectedDetails");
    expect(selectedDetailsSource).toContain("RoadmapObjectiveDetails");
    expect(selectedDetailsSource).toContain("StrategyNodeDetails");
    expect(canvasSource).toContain("CLICK_MOVEMENT_THRESHOLD");
    expect(canvasSource).toContain("shouldCommit && !wasDragged");
    expect(canvasSource).toContain('target.closest("[data-card-select]")');
    expect(cardSource).toContain("onOpenDetails");
    expect(cardSource).toContain("data-card-select");
    expect(cardSource).not.toContain("hover:opacity-75");
    expect(detailsSource).toContain("border-0 bg-transparent");
    expect(detailsSource).toContain("Save changes");
  });

  it("uses a right-aligned chevron for key-result expansion", () => {
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );

    expect(cardSource).toContain("justify-between gap-1");
    expect(cardSource).toContain('isExpanded && "rotate-90"');
    expect(cardSource).not.toContain("ArrowDownIcon");
  });
});
