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
    expect(cardSource).not.toContain("OKRIcon");
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

  it("uses the objective rich-text experience for strategy descriptions", () => {
    const editorSource = readSource(
      "src/modules/strategy/strategy-editor-dialog.tsx",
    );
    const descriptionEditorSource = readSource(
      "src/modules/strategy/strategy-description-editor.tsx",
    );
    const detailsSource = readSource(
      "src/modules/strategy/strategy-node-details.tsx",
    );
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );
    const pageSource = readSource("src/modules/strategy/index.tsx");

    expect(editorSource).toContain("StrategyDescriptionEditor");
    expect(editorSource).toContain('className="max-w-4xl"');
    expect(editorSource).toContain("h-auto border-0 bg-transparent");
    expect(editorSource).not.toContain("TextArea");
    expect(editorSource).not.toContain(">Description</Text>");
    expect(descriptionEditorSource).toContain("useEditor");
    expect(descriptionEditorSource).toContain("createRichTextStarterKit");
    expect(descriptionEditorSource).toContain("Underline");
    expect(descriptionEditorSource).toContain("autolink: true");
    expect(descriptionEditorSource).toContain("currentEditor.getHTML()");
    expect(descriptionEditorSource).toContain("currentEditor.isEmpty");
    expect(descriptionEditorSource).toContain("immediatelyRender: false");
    expect(descriptionEditorSource).toContain("TextEditor");
    expect(detailsSource).toContain("StrategyDescriptionEditor");
    expect(detailsSource).not.toContain("TextArea");
    expect(cardSource).toContain("getStrategyDescriptionPreview");
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
    expect(cardSource).toContain("dark:bg-surface-elevated/55");
    expect(canvasSource).toContain("dark:bg-surface-elevated/35");
    expect(pageSource).toContain("dark:bg-surface-elevated/20");
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
    expect(detailsSource).toContain("useDebouncedCallback");
    expect(detailsSource).toContain("AUTOSAVE_DELAY = 1000");
    expect(detailsSource).toContain("flushOnUnmount: true");
    expect(detailsSource).toContain("onBlur={flushSave}");
    expect(detailsSource).not.toContain("Save changes");
    expect(detailsSource).not.toContain("Saving...");
    expect(pageSource).not.toContain('title="Edit ultimate goal"');
    expect(pageSource).not.toContain('title="Edit strategic pillar"');
    expect(canvasSource).not.toContain("onEditGoal");
    expect(canvasSource).not.toContain("onEditPillar");
    expect(detailsSource.indexOf('className="mb-3 gap-5"')).toBeLessThan(
      detailsSource.indexOf("<Input"),
    );
  });

  it("uses a right-aligned chevron for key-result expansion", () => {
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );

    expect(cardSource).toContain("justify-between gap-1");
    expect(cardSource).toContain(
      "const [isExpanded, setIsExpanded] = useState(true)",
    );
    expect(cardSource).toContain(
      'className="text-foreground flex w-full items-center',
    );
    expect(cardSource).toContain('isExpanded && "rotate-90"');
    expect(cardSource).not.toContain("ArrowDownIcon");
    expect(cardSource).toContain("flex items-center gap-2.5 border-b");
    expect(cardSource).toContain("dark:border-border-strong/55");
    expect(cardSource).toContain("relative top-0.5 shrink-0");
    expect(cardSource).toContain("last:border-b-0");
    expect(cardSource).toContain("text-foreground line-clamp-2 text-base");
    expect(cardSource).toContain("<Tooltip");
    expect(cardSource).toContain("delayDuration={300}");
    expect(cardSource).toContain("size={14}");
    expect(cardSource).toContain("py-2.5");
    expect(cardSource).not.toContain("text-foreground/75");
  });

  it("keeps the goal label prominent and pillar content compact", () => {
    const cardSource = readSource(
      "src/modules/strategy/strategy-map-cards.tsx",
    );

    expect(cardSource).toContain(
      'className="text-text-primary text-[0.74rem] font-bold"',
    );
    expect(cardSource).not.toContain('aria-label="Edit ultimate goal"');
    expect(cardSource).toContain('className="mt-1 w-full text-left"');
    expect(cardSource).toContain("mt-1 line-clamp-3 leading-5");
    expect(cardSource).toContain('className="mt-2 gap-2"');
    expect(cardSource).toContain("text-[0.75rem]");
  });

  it("shares pillar alignment across objective property surfaces", () => {
    const hooksSource = readSource("src/modules/strategy/hooks.ts");
    const pillarPropertySource = readSource(
      "src/modules/strategy/objective-pillar-property.tsx",
    );
    const objectivePropertiesSource = readSource(
      "src/modules/objectives/stories/overview/properties.tsx",
    );
    const roadmapPropertiesSource = readSource(
      "src/modules/roadmap/components/objective-details-properties.tsx",
    );

    expect(pillarPropertySource).toContain("useStrategyMap");
    expect(pillarPropertySource).toContain("useAlignObjectiveMutation");
    expect(pillarPropertySource).toContain(
      "if (pillars.length === 0) return null",
    );
    expect(pillarPropertySource).toContain("Align to pillar");
    expect(pillarPropertySource).not.toContain("Add to pillar");
    expect(pillarPropertySource).toContain("Remove pillar alignment");
    expect(objectivePropertiesSource).toContain("ObjectivePillarProperty");
    expect(roadmapPropertiesSource).toContain("ObjectivePillarProperty");
    expect(hooksSource).toContain("onMutate: async");
    expect(hooksSource).toContain("cancelQueries");
    expect(hooksSource).toContain("previousStrategy");
    expect(hooksSource).toContain("setQueryData<StrategyMap>");
    expect(hooksSource).toContain("alignObjectiveInStrategy");
  });
});
