import { type NextRequest, NextResponse } from "next/server";
import {
  getAgentMarkdown,
  getAgentNotFoundMarkdown,
} from "@/lib/agent-content";
import { preferredRepresentation } from "@/lib/content-negotiation";

const appendVaryAccept = (headers: Headers) => {
  const values = new Set(
    (headers.get("Vary") ?? "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
  );
  values.add("Accept");
  headers.set("Vary", Array.from(values).join(", "));
};

const markdownResponse = (body: string, status = 200) =>
  new NextResponse(body, {
    status,
    headers: {
      "Cache-Control": "public, max-age=300",
      "Content-Type": "text/markdown; charset=utf-8",
      Vary: "Accept, Accept-Encoding",
    },
  });

const getCanonicalMarkdownPath = (pathname: string) => {
  if (pathname === "/index.md") return "/";
  const withoutExtension = pathname.slice(0, -3);
  return withoutExtension || "/";
};

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (request.headers.has("rsc")) {
    return NextResponse.next();
  }

  const explicitMarkdown = pathname.endsWith(".md");
  const canonicalPath = explicitMarkdown
    ? getCanonicalMarkdownPath(pathname)
    : pathname;
  const preferred = explicitMarkdown
    ? "text/markdown"
    : preferredRepresentation(request.headers.get("Accept"));

  if (preferred === "text/markdown") {
    const markdown = getAgentMarkdown(canonicalPath);
    return markdown
      ? markdownResponse(markdown)
      : markdownResponse(getAgentNotFoundMarkdown(canonicalPath), 404);
  }

  if (preferred === null) {
    return new NextResponse(
      "Not Acceptable\n\nAvailable: text/html, text/markdown\n",
      {
        status: 406,
        headers: {
          "Content-Type": "text/plain; charset=utf-8",
          Vary: "Accept",
        },
      },
    );
  }

  const response = NextResponse.next();
  appendVaryAccept(response.headers);
  const markdown = getAgentMarkdown(pathname);
  if (markdown) {
    response.headers.set(
      "Link",
      `<${pathname === "/" ? "/index.md" : `${pathname}.md`}>; rel="alternate"; type="text/markdown"`,
    );
  }
  return response;
}

export const config = {
  matcher: ["/((?!api/|_next/|ingest/|.*\\..*).*)", "/:path*.md"],
};
