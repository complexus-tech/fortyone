/* eslint-disable no-script-url -- Security regression fixtures must include rejected schemes. */
import { render, screen } from "@testing-library/react";
import {
  extractMessageLinks,
  findChatLinks,
  getChatLinkHref,
} from "./chat-links";
import { ChatPlainText, ChatRichText } from "./chat-rich-text";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

describe("chat links", () => {
  it("uses story references for internal URLs while keeping the full destination", () => {
    render(
      <ChatPlainText text="Review https://acme.fortyone.app/work/TM-276" />,
    );
    expect(screen.getByRole("link", { name: "TM-276" })).toHaveAttribute(
      "href",
      "https://acme.fortyone.app/work/TM-276",
    );
    expect(
      screen.queryByText("https://acme.fortyone.app/work/TM-276"),
    ).not.toBeInTheDocument();
  });
  it("preserves a deliberately named internal link", () => {
    render(
      <ChatRichText
        content={{
          type: "doc",
          content: [
            {
              type: "paragraph",
              content: [
                {
                  type: "text",
                  text: "Launch plan",
                  marks: [
                    {
                      type: "link",
                      attrs: { href: "https://acme.fortyone.app/work/TM-276" },
                    },
                  ],
                },
              ],
            },
          ],
        }}
      />,
    );
    expect(screen.getByRole("link", { name: "Launch plan" })).toBeVisible();
  });

  it("derives a bounded, deduplicated preview list from legacy plain messages", () => {
    const text = Array.from(
      { length: 60 },
      (_, index) => `https://example.com/${index}`,
    ).join(" ");
    const links = extractMessageLinks(`https://example.com/0 ${text}`, null);
    expect(links).toHaveLength(50);
    expect(links[0]).toEqual({
      id: "https://example.com/0",
      url: "https://example.com/0",
      title: "https://example.com/0",
    });
    expect(links.at(-1)?.url).toBe("https://example.com/49");
  });
  it("derives rich previews from visible destinations, excluding code and URL-like labels", () => {
    const links = extractMessageLinks("https://mirror.example/", {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "text",
              text: "https://label.example",
              marks: [
                {
                  type: "link",
                  attrs: { href: "https://destination.example/page#part" },
                },
              ],
            },
            { type: "text", text: " www.visible.example.com " },
            {
              type: "text",
              text: "https://inline-code.example",
              marks: [{ type: "code" }],
            },
            {
              type: "text",
              text: "https://unsafe-label.example",
              marks: [{ type: "link", attrs: { href: "javascript:alert(1)" } }],
            },
          ],
        },
        {
          type: "codeBlock",
          content: [{ type: "text", text: "https://block-code.example" }],
        },
      ],
    });
    expect(links.map((link) => link.url)).toEqual([
      "https://destination.example/page#part",
      "https://www.visible.example.com/",
    ]);
    expect(
      extractMessageLinks("https://mirror.example", {
        type: "doc",
        content: [],
      }),
    ).toEqual([]);
  });
  it("recognizes pasted URLs, domains, and emails without trailing punctuation", () => {
    const links = findChatLinks(
      "See https://example.com/a_(b). Also (www.example.org), example.net and hi@example.com!",
    );
    expect(links.map((link) => link.href)).toEqual([
      "https://example.com/a_(b)",
      "https://www.example.org/",
      "https://example.net/",
      "mailto:hi@example.com",
    ]);
  });
  it("excludes code spans and complete or unfinished fenced code", () => {
    expect(
      findChatLinks(
        "`https://inline.example` https://outside.example\n```js\nhttps://block.example\n```\nhttps://after.example",
      ).map((link) => link.href),
    ).toEqual(["https://outside.example/", "https://after.example/"]);
    expect(findChatLinks("~~~\nhttps://unfinished.example")).toEqual([]);
    expect(findChatLinks("``code ` https://nested.example``")).toEqual([]);
  });
  it("keeps links inside code until a matching backtick run closes it", () => {
    expect(
      findChatLinks(
        "`https://hidden.example `` https://still-hidden.example` https://visible.example",
      ).map((link) => link.href),
    ).toEqual(["https://visible.example/"]);
  });
  it("allows only safe explicit protocols", () => {
    for (const value of [
      "javascript:alert(1)",
      "data:text/html,hi",
      "file:///tmp/a",
      "ftp://example.com",
      "//example.com",
    ]) {
      expect(getChatLinkHref(value)).toBeNull();
    }
    expect(findChatLinks("ftp://example.com")).toEqual([]);
  });
  it("links legacy plain messages", () => {
    render(<ChatPlainText text="Go to https://example.com." />);
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "https://example.com/",
    );
  });
  it("keeps explicit rich links authoritative and code literal", () => {
    render(
      <ChatRichText
        content={{
          type: "doc",
          content: [
            {
              type: "paragraph",
              content: [
                {
                  type: "text",
                  text: "https://label.example",
                  marks: [
                    {
                      type: "link",
                      attrs: { href: "https://destination.example" },
                    },
                  ],
                },
              ],
            },
            {
              type: "paragraph",
              content: [
                {
                  type: "text",
                  text: "https://inline.example",
                  marks: [{ type: "code" }],
                },
              ],
            },
            {
              type: "codeBlock",
              content: [{ type: "text", text: "https://block.example" }],
            },
            {
              type: "paragraph",
              content: [
                {
                  type: "text",
                  text: "Click me",
                  marks: [
                    { type: "link", attrs: { href: "javascript:alert(1)" } },
                  ],
                },
              ],
            },
          ],
        }}
      />,
    );
    expect(screen.getAllByRole("link")).toHaveLength(1);
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "https://destination.example/",
    );
  });
});
