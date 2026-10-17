import { render, screen } from "@testing-library/react";
import { MessagePreview } from "./message-preview";

describe("MessagePreview", () => {
  it("preserves mention labels and separates rich blocks without nested links", () => {
    render(
      <button type="button">
        <MessagePreview
          content={{
            type: "doc",
            content: [
              {
                type: "paragraph",
                content: [
                  { type: "mention", attrs: { label: "Sarah" } },
                  { type: "text", text: ", please review" },
                ],
              },
              {
                type: "paragraph",
                content: [
                  {
                    type: "text",
                    text: "the launch plan",
                    marks: [
                      {
                        type: "link",
                        attrs: { href: "https://example.com/plan" },
                      },
                    ],
                  },
                  { type: "hardBreak" },
                  { type: "text", text: "today." },
                ],
              },
            ],
          }}
          text="@Sarah, please review the launch plan today."
        />
      </button>,
    );

    expect(screen.getByRole("button")).toHaveTextContent(
      "@Sarah, please review the launch plan today.",
    );
    expect(screen.getByText("@Sarah")).toBeInTheDocument();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it("keeps plain messages and attachment fallbacks when rich content is absent", () => {
    const { rerender } = render(
      <MessagePreview content={null} text="A plain message" />,
    );
    expect(screen.getByText("A plain message")).toBeInTheDocument();
    rerender(
      <MessagePreview
        content={{ type: "doc", content: [] }}
        text="Shared an attachment"
      />,
    );
    expect(screen.getByText("Shared an attachment")).toBeInTheDocument();
  });

  it("keeps the attachment fallback when the editor contains only empty paragraphs", () => {
    render(
      <MessagePreview
        content={{
          type: "doc",
          content: [
            { type: "paragraph" },
            {
              type: "paragraph",
              content: [{ type: "text", text: " " }, { type: "hardBreak" }],
            },
          ],
        }}
        text="Shared an attachment"
      />,
    );
    expect(screen.getByText("Shared an attachment")).toBeInTheDocument();
  });
});
