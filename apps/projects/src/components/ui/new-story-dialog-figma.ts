import type { Editor } from "@tiptap/react";
import { toast } from "sonner";

type FigmaArtifactReference = {
  canonicalUrl: string;
};

type FigmaDescriptionDraft = {
  overview: string;
  requirements: string[];
  acceptanceCriteria: string[];
  implementationNotes: string[];
};

type AttachFigmaDesignsInput<TArtifact extends FigmaArtifactReference> = {
  artifacts: TArtifact[];
  linkDesign: (input: { storyId: string; url: string }) => Promise<unknown>;
  storyId: string;
};

export const attachFigmaDesigns = async <
  TArtifact extends FigmaArtifactReference,
>({
  artifacts,
  linkDesign,
  storyId,
}: AttachFigmaDesignsInput<TArtifact>) => {
  const results = await Promise.allSettled(
    artifacts.map((artifact) =>
      linkDesign({ storyId, url: artifact.canonicalUrl }),
    ),
  );
  const failedArtifacts = artifacts.filter(
    (_, index) => results[index]?.status === "rejected",
  );
  const attachedCount = artifacts.length - failedArtifacts.length;

  if (failedArtifacts.length === 0) {
    toast.success(
      `${artifacts.length} Figma design${artifacts.length === 1 ? "" : "s"} attached`,
    );
    return;
  }

  toast.error(
    failedArtifacts.length === artifacts.length
      ? "Figma designs could not be attached"
      : `${attachedCount} of ${artifacts.length} Figma designs attached`,
    {
      description:
        "The story was created. Retry the remaining design attachments.",
      action: {
        label: "Retry",
        onClick: () => {
          void attachFigmaDesigns({
            artifacts: failedArtifacts,
            linkDesign,
            storyId,
          });
        },
      },
    },
  );
};

const createListSection = (heading: string, items: string[]) =>
  items.length > 0
    ? [
        {
          type: "heading",
          attrs: { level: 3 },
          content: [{ type: "text", text: heading }],
        },
        {
          type: "bulletList",
          content: items.map((text) => ({
            type: "listItem",
            content: [
              {
                type: "paragraph",
                content: [{ type: "text", text }],
              },
            ],
          })),
        },
      ]
    : [];

export const insertFigmaDescription = (
  editor: Editor | null,
  description: FigmaDescriptionDraft,
) => {
  if (!editor) return;

  editor
    .chain()
    .focus()
    .insertContent([
      {
        type: "heading",
        attrs: { level: 3 },
        content: [{ type: "text", text: "Overview" }],
      },
      {
        type: "paragraph",
        content: [{ type: "text", text: description.overview }],
      },
      ...createListSection("Requirements", description.requirements),
      ...createListSection(
        "Acceptance criteria",
        description.acceptanceCriteria,
      ),
      ...createListSection(
        "Implementation notes",
        description.implementationNotes,
      ),
    ])
    .run();
};
