/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import { PdfThumbnail } from "./pdf-thumbnail";

jest.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  version: "6.4.299",
  GlobalWorkerOptions: {},
  AnnotationMode: { DISABLE: 0 },
  getDocument: (...args: unknown[]) => mockGetDocument(...args),
}));

const mockGetDocument = jest.fn();
const originalIntersectionObserver = Object.getOwnPropertyDescriptor(
  globalThis,
  "IntersectionObserver",
);
const originalPixelRatio = Object.getOwnPropertyDescriptor(
  window,
  "devicePixelRatio",
);
type Observer = {
  callback: IntersectionObserverCallback;
  disconnect: jest.Mock;
};
let observers: Observer[];

const deferred = <T,>() => {
  let resolve: (value: T) => void = () => undefined;
  let reject: (reason: Error) => void = () => undefined;
  const promise = new Promise<T>((accept, decline) => {
    resolve = accept;
    reject = decline;
  });
  return { promise, resolve, reject };
};

const makeDocument = ({ pendingRender = false } = {}) => {
  const painting = deferred<undefined>();
  if (!pendingRender) painting.resolve(undefined);
  const cancel = jest.fn(() => {
    painting.reject(new Error("Rendering canceled"));
  });
  const page = {
    getViewport: jest.fn(({ scale }: { scale: number }) => ({
      width: 600 * scale,
      height: 800 * scale,
    })),
    cleanup: jest.fn(),
    render: jest.fn(() => ({ promise: painting.promise, cancel })),
  };
  const document = {
    numPages: 100,
    getPage: jest.fn().mockResolvedValue(page),
  };
  const task = {
    promise: Promise.resolve(document),
    destroy: jest.fn().mockResolvedValue(undefined),
  };
  return { painting, cancel, page, document, task };
};

beforeEach(() => {
  observers = [];
  mockGetDocument.mockReset().mockImplementation(() => makeDocument().task);
  Object.defineProperty(globalThis, "IntersectionObserver", {
    configurable: true,
    value: class {
      callback: IntersectionObserverCallback;
      observe = jest.fn();
      disconnect = jest.fn();
      constructor(callback: IntersectionObserverCallback) {
        this.callback = callback;
        observers.push(this);
      }
    },
  });
  Object.defineProperty(window, "devicePixelRatio", {
    configurable: true,
    value: 3,
  });
  jest
    .spyOn(HTMLCanvasElement.prototype, "getContext")
    .mockReturnValue({} as CanvasRenderingContext2D);
});

afterEach(async () => {
  cleanup();
  await act(async () => undefined);
  jest.restoreAllMocks();
  if (originalIntersectionObserver) {
    Object.defineProperty(
      globalThis,
      "IntersectionObserver",
      originalIntersectionObserver,
    );
  } else Reflect.deleteProperty(globalThis, "IntersectionObserver");
  if (originalPixelRatio) {
    Object.defineProperty(window, "devicePixelRatio", originalPixelRatio);
  }
});

const intersect = (index = 0, isIntersecting = true) => {
  act(() => {
    const observer = observers[index];
    observer.callback(
      [{ isIntersecting } as IntersectionObserverEntry],
      observer as unknown as IntersectionObserver,
    );
  });
};

describe("inline PDF thumbnails", () => {
  it("waits until near the viewport, renders only page one, and releases its worker", async () => {
    const pdf = makeDocument();
    mockGetDocument.mockReturnValue(pdf.task);
    const { container } = render(
      <button type="button">
        <PdfThumbnail fileName="Guide.pdf" url="/api/guide/preview" />
      </button>,
    );
    expect(screen.getByText("Loading PDF preview…")).toBeVisible();
    expect(screen.getAllByRole("button")).toHaveLength(1);
    intersect(0, false);
    await act(async () => undefined);
    expect(mockGetDocument).not.toHaveBeenCalled();

    intersect();
    await waitFor(() => {
      expect(screen.getByRole("img")).toHaveAttribute("aria-busy", "false");
    });
    expect(pdf.document.getPage).toHaveBeenCalledTimes(1);
    expect(pdf.document.getPage).toHaveBeenCalledWith(1);
    expect(pdf.page.cleanup).toHaveBeenCalledTimes(1);
    expect(pdf.task.destroy).toHaveBeenCalledTimes(1);
    expect(mockGetDocument).toHaveBeenCalledWith(
      expect.objectContaining({
        url: "/api/guide/preview",
        cMapUrl: "/pdfjs/6.4.299/cmaps/",
        disableAutoFetch: true,
        disableStream: true,
      }),
    );
    const canvas = container.querySelector("canvas")!;
    expect(canvas).toBeVisible();
    expect(canvas.width).toBe(840);
    expect(canvas.height).toBe(525);
    expect(canvas.width * canvas.height).toBeLessThanOrEqual(1_000_000);
    expect(observers[0].disconnect).toHaveBeenCalled();
  });

  it("limits active render jobs to two and drops a queued thumbnail after unmount", async () => {
    const first = makeDocument({ pendingRender: true });
    const second = makeDocument({ pendingRender: true });
    const fourth = makeDocument();
    mockGetDocument
      .mockReturnValueOnce(first.task)
      .mockReturnValueOnce(second.task)
      .mockReturnValueOnce(fourth.task);
    render(<PdfThumbnail fileName="One.pdf" url="/one" />);
    render(<PdfThumbnail fileName="Two.pdf" url="/two" />);
    const queued = render(<PdfThumbnail fileName="Three.pdf" url="/three" />);
    render(<PdfThumbnail fileName="Four.pdf" url="/four" />);
    observers.forEach((_observer, index) => {
      intersect(index);
    });
    await waitFor(() => {
      expect(second.page.render).toHaveBeenCalled();
    });
    expect(mockGetDocument).toHaveBeenCalledTimes(2);

    queued.unmount();
    await act(async () => {
      first.painting.resolve(undefined);
    });
    await waitFor(() => {
      expect(mockGetDocument).toHaveBeenCalledTimes(3);
    });
    expect(mockGetDocument.mock.calls.map(([options]) => options.url)).toEqual([
      "/one",
      "/two",
      "/four",
    ]);
    await act(async () => {
      second.painting.resolve(undefined);
    });
  });

  it("cancels active painting and releases canvas memory on unmount", async () => {
    const pdf = makeDocument({ pendingRender: true });
    mockGetDocument.mockReturnValue(pdf.task);
    const view = render(<PdfThumbnail fileName="Guide.pdf" url="/guide" />);
    intersect();
    await waitFor(() => {
      expect(pdf.page.render).toHaveBeenCalled();
    });
    const canvas = view.container.querySelector("canvas")!;
    view.unmount();
    await act(async () => undefined);
    expect(pdf.cancel).toHaveBeenCalledTimes(1);
    expect(pdf.task.destroy).toHaveBeenCalledTimes(1);
    expect(canvas.width).toBe(0);
    expect(canvas.height).toBe(0);
  });

  it("aborts an unfinished document load when unmounted", async () => {
    const pending = deferred<unknown>();
    const destroy = jest.fn(async () => {
      pending.reject(new Error("Loading canceled"));
    });
    mockGetDocument.mockReturnValue({ promise: pending.promise, destroy });
    const view = render(<PdfThumbnail fileName="Guide.pdf" url="/guide" />);
    intersect();
    await waitFor(() => {
      expect(mockGetDocument).toHaveBeenCalled();
    });
    view.unmount();
    await act(async () => undefined);
    expect(destroy).toHaveBeenCalledTimes(1);
  });

  it("preserves the placeholder size and exposes a usable fallback on failure", async () => {
    const destroy = jest.fn().mockResolvedValue(undefined);
    mockGetDocument.mockImplementation(() => ({
      promise: Promise.reject(new Error("Invalid PDF")),
      destroy,
    }));
    render(<PdfThumbnail fileName="Guide.pdf" url="/guide" />);
    const region = screen.getByRole("img");
    const initialClassName = region.className;
    intersect();
    expect(
      await screen.findByText("PDF preview unavailable · Open to view"),
    ).toBeVisible();
    expect(region).toHaveAccessibleName(
      "PDF preview unavailable for Guide.pdf",
    );
    expect(region.className).toBe(initialClassName);
    expect(destroy).toHaveBeenCalledTimes(1);
  });

  it("loads a fresh thumbnail when the file URL changes", async () => {
    const first = makeDocument();
    const next = makeDocument();
    mockGetDocument
      .mockReturnValueOnce(first.task)
      .mockReturnValueOnce(next.task);
    const view = render(<PdfThumbnail fileName="Guide.pdf" url="/guide" />);
    intersect();
    await waitFor(() => {
      expect(screen.getByRole("img")).toHaveAttribute("aria-busy", "false");
    });
    view.rerender(<PdfThumbnail fileName="Next.pdf" url="/next" />);
    expect(screen.getByText("Loading PDF preview…")).toBeVisible();
    expect(mockGetDocument).toHaveBeenCalledTimes(1);
    intersect(1);
    await waitFor(() => {
      expect(next.page.render).toHaveBeenCalled();
    });
    expect(screen.getByRole("img")).toHaveAccessibleName(
      "First page of Next.pdf",
    );
  });
});
