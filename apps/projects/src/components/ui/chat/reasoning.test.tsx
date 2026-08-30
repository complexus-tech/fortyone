import { execFileSync } from "node:child_process";
import { fireEvent, render, screen } from "@testing-library/react";
import { REASONING_SANITIZE_SCHEMA } from "./reasoning-sanitize";
import { Reasoning } from "./reasoning";

type MarkdownProps = {
  children: string;
  rehypePlugins?: unknown[];
};

const mockMarkdown = jest.fn(({ children }: MarkdownProps) => <>{children}</>);

jest.mock("react-markdown", () => ({
  __esModule: true,
  default: (props: MarkdownProps) => mockMarkdown(props),
}));
jest.mock("rehype-raw", () => ({ __esModule: true, default: () => undefined }));
jest.mock("rehype-sanitize", () => ({
  __esModule: true,
  default: () => undefined,
}));
jest.mock("remark-gfm", () => ({ __esModule: true, default: () => undefined }));

const SANITIZE_MARKDOWN_SCRIPT = `
import React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import Markdown from "react-markdown";
import rehypeRaw from "rehype-raw";
import rehypeSanitize from "rehype-sanitize";
import remarkGfm from "remark-gfm";

const schema = JSON.parse(Buffer.from(process.argv[1], "base64").toString("utf8"));
const content = Buffer.from(process.argv[2], "base64").toString("utf8");
const markdown = React.createElement(
  Markdown,
  {
    rehypePlugins: [rehypeRaw, [rehypeSanitize, schema]],
    remarkPlugins: [remarkGfm],
  },
  content,
);

process.stdout.write(renderToStaticMarkup(markdown));
`;

const toBase64 = (value: string) => Buffer.from(value).toString("base64");

const sanitizeMarkdown = (content: string) =>
  execFileSync(
    process.execPath,
    [
      "--input-type=module",
      "--eval",
      SANITIZE_MARKDOWN_SCRIPT,
      toBase64(JSON.stringify(REASONING_SANITIZE_SCHEMA)),
      toBase64(content),
    ],
    { cwd: process.cwd(), encoding: "utf8" },
  );

const parseHtml = (html: string) => {
  const container = document.createElement("div");
  container.innerHTML = html;
  return container;
};

describe("Reasoning", () => {
  beforeEach(() => {
    mockMarkdown.mockClear();
  });

  it("runs the sanitizer after raw HTML parsing", () => {
    render(<Reasoning content="Safe reasoning" />);

    fireEvent.click(screen.getByRole("button", { name: "Show reasoning" }));

    expect(mockMarkdown).toHaveBeenCalledTimes(1);
    const markdownProps = mockMarkdown.mock.calls[0][0];
    expect(markdownProps.rehypePlugins).toHaveLength(2);
    expect(markdownProps.rehypePlugins?.[0]).toEqual(expect.any(Function));
    expect(markdownProps.rehypePlugins?.[1]).toEqual([
      expect.any(Function),
      REASONING_SANITIZE_SCHEMA,
    ]);
  });

  it("preserves Markdown, GFM, and the allowlisted reasoning HTML", () => {
    const content = `## Safe plan

- [x] Verify the **workspace**

| Step | Owner |
| --- | --- |
| Review | Maya |

<details open><summary>Context</summary><kbd>Command</kbd></details>

[Documentation](https://example.com/docs)

![Architecture](https://example.com/architecture.png "Diagram")`;

    const renderedContent = parseHtml(sanitizeMarkdown(content));

    expect(renderedContent.querySelector("h2")?.textContent).toBe("Safe plan");
    expect(
      renderedContent.querySelector("input[type='checkbox']"),
    ).toBeChecked();
    expect(
      renderedContent.querySelector("input[type='checkbox']"),
    ).toBeDisabled();
    expect(renderedContent.querySelector("table")).not.toBeNull();
    expect(renderedContent.querySelector("details")).toHaveAttribute("open");
    expect(renderedContent.querySelector("kbd")?.textContent).toBe("Command");
    expect(renderedContent.querySelector("a")).toHaveAttribute(
      "href",
      "https://example.com/docs",
    );
    expect(renderedContent.querySelector("img")).toHaveAttribute(
      "src",
      "https://example.com/architecture.png",
    );
  });

  it("removes executable HTML, dangerous URLs, and clobbering attributes", () => {
    const content = `<script>window.compromised = true</script>

<style>body { display: none; }</style>

<iframe srcdoc="<script>alert('iframe')</script>"></iframe>

<object data="javascript:alert('object')"></object>

<img alt="Unsafe image" src="javascript:alert('image')" onerror="alert('event')">

<a id="location" name="constructor" href="javascript:alert('link')" onclick="alert('event')">Unsafe link</a>

<a href="data:text/html,<script>alert('data')</script>">Data link</a>

<form id="document" name="window"><input name="location"></form>`;

    const renderedContent = parseHtml(sanitizeMarkdown(content));

    expect(renderedContent.querySelector("script")).not.toBeInTheDocument();
    expect(renderedContent.querySelector("style")).not.toBeInTheDocument();
    expect(renderedContent.querySelector("iframe")).not.toBeInTheDocument();
    expect(renderedContent.querySelector("object")).not.toBeInTheDocument();
    expect(renderedContent.querySelector("form")).not.toBeInTheDocument();

    const unsafeImage = renderedContent.querySelector("img");
    expect(unsafeImage).not.toHaveAttribute("src");
    expect(unsafeImage).not.toHaveAttribute("onerror");

    const links = renderedContent.querySelectorAll("a");
    expect(links).toHaveLength(2);
    expect(links[0]).not.toHaveAttribute("href");
    expect(links[0]).not.toHaveAttribute("onclick");
    expect(links[0]).not.toHaveAttribute("id");
    expect(links[0]).not.toHaveAttribute("name");
    expect(links[1]).not.toHaveAttribute("href");

    expect(
      renderedContent.querySelector("[id='document']"),
    ).not.toBeInTheDocument();
    expect(
      renderedContent.querySelector("[name='window']"),
    ).not.toBeInTheDocument();
    expect(
      renderedContent.querySelector("[name='location']"),
    ).not.toBeInTheDocument();
  });
});
