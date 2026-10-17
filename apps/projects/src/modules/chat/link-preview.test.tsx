/* eslint-disable no-script-url -- Security fixtures verify rejected preview URL schemes. */
/* eslint-disable @next/next/no-img-element -- next/image is mocked to test native image load errors. */
import type { ImgHTMLAttributes } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import type { LinkMetadata } from "@/lib/queries/links/get-metadata";
import {
  ChatLinkPreview,
  LinkResourceRow,
  MessageLinkPreviews,
} from "./link-preview";
import type { ChatResource } from "./types";

jest.mock("@/lib/hooks/link-metadata", () => ({ useLinkMetadata: jest.fn() }));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "preview" }),
}));
jest.mock("next/image", () => ({
  __esModule: true,
  default: ({
    alt,
    src,
    className,
    onError,
  }: ImgHTMLAttributes<HTMLImageElement>) => (
    <img alt={alt} className={className} onError={onError} src={src} />
  ),
}));
const mockMetadata = jest.mocked(useLinkMetadata);
const setMetadata = (data?: LinkMetadata) => {
  mockMetadata.mockReturnValue({ data } as ReturnType<typeof useLinkMetadata>);
};
class PreviewObserver implements IntersectionObserver {
  static instances: PreviewObserver[] = [];
  readonly root = null;
  readonly rootMargin = "160px";
  readonly scrollMargin = "0px";
  readonly thresholds = [0];
  readonly observe = jest.fn();
  readonly unobserve = jest.fn();
  readonly disconnect = jest.fn();
  readonly takeRecords = jest.fn(() => []);
  constructor(private readonly callback: IntersectionObserverCallback) {
    PreviewObserver.instances.push(this);
  }
  intersect(isIntersecting: boolean) {
    this.callback([{ isIntersecting } as IntersectionObserverEntry], this);
  }
}
const originalObserver = globalThis.IntersectionObserver;
const favicon =
  "/api/metadata/image?url=https%3A%2F%2Fexample.com%2Ffavicon.ico";
const ogImage =
  "/api/metadata/image?url=https%3A%2F%2Fexample.com%2Fsocial.png";
const resource: ChatResource = {
  id: "resource-1",
  conversationId: "room-1",
  attachmentId: "",
  description: "",
  kind: "link",
  messageId: "message-1",
  createdAt: "2026-10-17T09:00:00Z",
  author: {
    id: "author-1",
    fullName: "Sarah",
    avatarUrl: null,
    isSystem: false,
  },
  fileName: "",
  title: "Design review",
  mediaType: "",
  sizeBytes: 0,
  url: "https://example.com/review",
};
beforeEach(() => {
  jest.clearAllMocks();
  PreviewObserver.instances = [];
  globalThis.IntersectionObserver = PreviewObserver;
  setMetadata();
});
afterAll(() => {
  globalThis.IntersectionObserver = originalObserver;
});

describe("message link previews", () => {
  it.each([
    "javascript:alert(1)",
    "data:text/html,hello",
    "file:///tmp/document",
    "ftp://example.com/file",
    "mailto:hello@example.com",
    "https://user:password@example.com/private",
    "https://user@example.com/private",
    "not a URL",
  ])("excludes unsafe or unsupported preview URL %s", (url) => {
    render(<ChatLinkPreview url={url} />);
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(
      mockMetadata.mock.calls.every(([, options]) => !options?.enabled),
    ).toBe(true);
    expect(PreviewObserver.instances).toHaveLength(0);
  });
  it("deduplicates canonical URLs and renders at most four cards", () => {
    render(
      <MessageLinkPreviews
        links={[
          {
            id: "1",
            url: "https://EXAMPLE.com:443/page#first",
            title: "First",
          },
          {
            id: "2",
            url: "https://example.com/page#second",
            title: "Duplicate",
          },
          { id: "3", url: "mailto:hello@example.com", title: "Email" },
          { id: "4", url: "https://two.example.com/", title: "Second" },
          { id: "5", url: "https://three.example.com/", title: "Third" },
          { id: "6", url: "https://four.example.com/", title: "Fourth" },
          { id: "7", url: "https://five.example.com/", title: "Fifth" },
        ]}
      />,
    );
    expect(
      screen.getAllByRole("link").map((link) => link.getAttribute("href")),
    ).toEqual([
      "https://example.com/page#first",
      "https://two.example.com/",
      "https://three.example.com/",
      "https://four.example.com/",
    ]);
    expect(screen.queryByText("Duplicate")).not.toBeInTheDocument();
    expect(screen.queryByText("Fifth")).not.toBeInTheDocument();
  });
  it("enables metadata only after the card intersects the viewport margin", () => {
    const { unmount } = render(
      <ChatLinkPreview url="https://example.com/page#section" />,
    );
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      "https://example.com/page#section",
    );
    expect(mockMetadata).toHaveBeenLastCalledWith("https://example.com/page", {
      enabled: false,
    });
    const observer = PreviewObserver.instances[0];
    expect(observer.observe).toHaveBeenCalledTimes(1);
    act(() => {
      observer.intersect(false);
    });
    expect(mockMetadata).toHaveBeenLastCalledWith("https://example.com/page", {
      enabled: false,
    });
    act(() => {
      observer.intersect(true);
    });
    expect(mockMetadata).toHaveBeenLastCalledWith("https://example.com/page", {
      enabled: true,
    });
    expect(observer.disconnect).toHaveBeenCalledTimes(1);
    unmount();
    expect(observer.disconnect).toHaveBeenCalledTimes(2);
  });
  it("renders favicon and OG image in their distinct presentations", () => {
    setMetadata({
      title: "Launch notes",
      description: "What changed this week",
      image: favicon,
      previewImage: ogImage,
    });
    const { container } = render(
      <ChatLinkPreview url="https://example.com/notes" />,
    );
    const images = Array.from(container.querySelectorAll("img"));
    expect(images).toHaveLength(2);
    expect(
      images.find((image) => image.getAttribute("src") === favicon),
    ).toHaveClass("object-contain");
    expect(
      images.find((image) => image.getAttribute("src") === ogImage),
    ).toHaveClass("object-cover");
    expect(
      screen.getByRole("link", { name: "Open Launch notes" }),
    ).toHaveAttribute("href", "https://example.com/notes");
    expect(screen.getByText("What changed this week")).toBeInTheDocument();
  });
  it("does not promote a favicon into a large OG image", () => {
    setMetadata({ image: favicon });
    const { container } = render(
      <ChatLinkPreview url="https://example.com/" />,
    );
    expect(container.querySelectorAll("img")).toHaveLength(1);
    expect(container.querySelector("img")).toHaveClass("object-contain");
    expect(
      screen.getByRole("link", { name: "Open example.com" }),
    ).toBeInTheDocument();
  });
  it("keeps the readable link and favicon when the OG image fails", () => {
    setMetadata({
      title: "Launch notes",
      image: favicon,
      previewImage: ogImage,
    });
    const { container } = render(
      <ChatLinkPreview url="https://example.com/notes" />,
    );
    const image = Array.from(container.querySelectorAll("img")).find(
      (item) => item.getAttribute("src") === ogImage,
    );
    expect(image).toBeDefined();
    fireEvent.error(image!);
    expect(container.querySelectorAll("img")).toHaveLength(1);
    expect(container.querySelector("img")).toHaveAttribute("src", favicon);
    expect(
      screen.getByRole("link", { name: "Open Launch notes" }),
    ).toHaveAttribute("href", "https://example.com/notes");
  });
  it("renders a readable fallback without metadata and refuses direct remote images", () => {
    setMetadata({
      image: "https://remote.example/favicon.ico",
      previewImage: "https://remote.example/og.png",
    });
    const { container } = render(
      <ChatLinkPreview url="https://www.example.com/notes" />,
    );
    expect(container.querySelector("img")).toBeNull();
    expect(
      screen.getByRole("link", { name: "Open example.com" }),
    ).toHaveAttribute("href", "https://www.example.com/notes");
  });
});
describe("link resource rows", () => {
  it("opens the supplied resource through its preview callback", () => {
    const onPreview = jest.fn();
    render(<LinkResourceRow onPreview={onPreview} resource={resource} />);
    fireEvent.click(
      screen.getByRole("button", { name: "Preview Design review" }),
    );
    expect(onPreview).toHaveBeenCalledTimes(1);
    expect(onPreview).toHaveBeenCalledWith(
      expect.objectContaining({ type: "click" }),
      resource,
    );
    expect(mockMetadata).toHaveBeenLastCalledWith(resource.url, {
      enabled: false,
    });
    expect(screen.getByText(/Shared by Sarah/)).toBeInTheDocument();
  });
});
