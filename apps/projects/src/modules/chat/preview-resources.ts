type PreviewResourceDetails = {
  id: string;
  conversationId: string;
  messageId: string;
  title: string;
  description: string;
  author: string;
  date: string;
};

export type PreviewResource = PreviewResourceDetails &
  (
    | {
        kind: "document" | "image";
        fileName: string;
        mediaType: string;
        sizeLabel: string;
      }
    | { kind: "link"; url: string }
  );

// Local samples only. File names and sizes describe the interface, not uploads.
export const PREVIEW_RESOURCES: PreviewResource[] = [
  {
    id: "product-review-outline",
    conversationId: "product",
    messageId: "product-20261012-03",
    kind: "document",
    title: "First-visit review outline",
    description:
      "A sample walkthrough outline covering the first screen, moments of hesitation, and a clear next step.",
    author: "You",
    date: "Monday, October 12 at 09:18",
    fileName: "first-visit-review.md",
    mediaType: "Markdown",
    sizeLabel: "6 KB",
  },
  {
    id: "product-empty-state-comparison",
    conversationId: "product",
    messageId: "product-20261013-06",
    kind: "image",
    title: "Empty-state comparison",
    description:
      "A sample comparison of an empty team and a populated team with no work assigned to a new teammate.",
    author: "Sarah Chen",
    date: "Tuesday, October 13 at 11:45",
    fileName: "empty-state-comparison.png",
    mediaType: "PNG",
    sizeLabel: "312 KB",
  },
  {
    id: "product-language-reference",
    conversationId: "product",
    messageId: "product-20261015-03",
    kind: "link",
    title: "FortyOne · Product language reference",
    description:
      "The FortyOne homepage, shared as a reference for the tone of the welcome experience.",
    author: "Sarah Chen",
    date: "Yesterday at 09:22",
    url: "https://fortyone.app",
  },
  {
    id: "welcome-brief",
    conversationId: "product",
    messageId: "product-1",
    kind: "document",
    title: "Welcome flow brief",
    description:
      "A short brief for making the first few minutes in a workspace feel clear and useful.",
    author: "Sarah Chen",
    date: "Today at 09:41",
    fileName: "welcome-flow-brief.pdf",
    mediaType: "PDF",
    sizeLabel: "184 KB",
  },
  {
    id: "welcome-checklist",
    conversationId: "product",
    messageId: "product-3",
    kind: "image",
    title: "Welcome checklist concept",
    description:
      "An early direction for a welcome screen with three clear steps and room for team priorities.",
    author: "Sarah Chen",
    date: "Today at 09:45",
    fileName: "welcome-checklist.png",
    mediaType: "PNG",
    sizeLabel: "420 KB",
  },
  {
    id: "product-workspace",
    conversationId: "product",
    messageId: "product-3",
    kind: "link",
    title: "FortyOne · Where teams move work forward",
    description:
      "A shared reference for the language and experience we want the welcome flow to reflect.",
    author: "Sarah Chen",
    date: "Today at 09:45",
    url: "https://fortyone.app",
  },
  {
    id: "walkthrough-notes",
    conversationId: "engineering",
    messageId: "engineering-1",
    kind: "document",
    title: "Build walkthrough notes",
    description:
      "A place to collect small-screen feedback, unclear steps, and follow-ups from the latest build.",
    author: "Daniel Williams",
    date: "Today at 10:12",
    fileName: "walkthrough-notes.md",
    mediaType: "Markdown",
    sizeLabel: "8 KB",
  },
  {
    id: "welcome-direction",
    conversationId: "design",
    messageId: "design-1",
    kind: "image",
    title: "Welcome screen direction",
    description:
      "Fewer choices up front, a clear next step, and more room for the actual work.",
    author: "Sarah Chen",
    date: "Today at 08:52",
    fileName: "welcome-direction.png",
    mediaType: "PNG",
    sizeLabel: "356 KB",
  },
];
