/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */

import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { PdfPreview } from "./pdf-preview";
import type { PreviewAttachment } from "./types";

jest.mock("./use-element-size", () => ({
  useElementSize: () => ({ width: 640, height: 480 }),
}));
jest.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  version: "6.4.299",
  GlobalWorkerOptions: {},
  AnnotationMode: { DISABLE: 0 },
  getDocument: (...args: unknown[]) => mockGetDocument(...args),
  TextLayer: class {
    render = mockTextRender;
    cancel = mockTextCancel;
  },
}));

const mockGetDocument = jest.fn();
const mockTextRender = jest.fn();
const mockTextCancel = jest.fn();
const mockDestroy = jest.fn();
const mockGetPage = jest.fn();
const file: PreviewAttachment = {
  id: "pdf",
  fileName: "Guide.pdf",
  mediaType: "application/pdf",
  sizeBytes: 3000,
  url: "/api/preview/guide.pdf",
};
const makePage = () => ({
  getViewport: ({ scale }: { scale: number }) => ({
    width: 600 * scale,
    height: 800 * scale,
    scale,
    userUnit: 1,
  }),
  getTextContent: jest.fn().mockResolvedValue({ items: [] }),
  cleanup: jest.fn(),
  render: jest.fn(() => ({ promise: Promise.resolve(), cancel: jest.fn() })),
});

beforeEach(() => {
  jest.clearAllMocks();
  mockDestroy.mockResolvedValue(undefined);
  mockTextRender.mockResolvedValue(undefined);
  mockGetPage.mockImplementation(async () => makePage());
  mockGetDocument.mockReturnValue({
    promise: Promise.resolve({ numPages: 3, getPage: mockGetPage }),
    destroy: mockDestroy,
  });
  jest
    .spyOn(HTMLCanvasElement.prototype, "getContext")
    .mockReturnValue({} as CanvasRenderingContext2D);
  HTMLElement.prototype.scrollTo = jest.fn();
});

afterEach(() => {
  cleanup();
  jest.restoreAllMocks();
});

describe("PDF preview resource lifecycle", () => {
  it("renders one page, keeps page navigation separate, and releases previous-page resources", async () => {
    const first = makePage();
    const second = makePage();
    mockGetPage.mockImplementation(async (number: number) =>
      number === 1 ? first : second,
    );
    const { container } = render(<PdfPreview attachment={file} />);
    await waitFor(() => {
      expect(first.render).toHaveBeenCalledTimes(1);
    });
    expect(container.querySelectorAll("canvas")).toHaveLength(1);
    expect(mockGetPage).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("button", { name: "Previous PDF page" }),
    ).toBeDisabled();
    const nextPage = screen.getByRole("button", { name: "Next PDF page" });
    act(() => {
      nextPage.focus();
    });
    fireEvent.click(nextPage);
    await waitFor(() => {
      expect(second.render).toHaveBeenCalledTimes(1);
    });
    expect(mockGetPage).toHaveBeenLastCalledWith(2);
    expect(first.cleanup).toHaveBeenCalledTimes(1);
    expect(nextPage).toHaveFocus();
    expect(
      screen.getByRole("textbox", { name: "PDF page number" }),
    ).toHaveValue("2");
  });

  it("retains same-page resources while a new zoom request is pending", async () => {
    const page = makePage();
    let finishNextPage: (value: ReturnType<typeof makePage>) => void = () =>
      undefined;
    const pendingPage = new Promise<ReturnType<typeof makePage>>((resolve) => {
      finishNextPage = resolve;
    });
    mockGetPage.mockResolvedValueOnce(page).mockReturnValueOnce(pendingPage);
    const view = render(<PdfPreview attachment={file} />);
    await waitFor(() => {
      expect(page.render).toHaveBeenCalledTimes(1);
    });
    fireEvent.click(screen.getByRole("button", { name: "Zoom PDF in" }));
    await waitFor(() => {
      expect(mockGetPage).toHaveBeenCalledTimes(2);
    });
    expect(page.cleanup).not.toHaveBeenCalled();
    await act(async () => {
      finishNextPage(page);
    });
    expect(page.render).toHaveBeenCalledTimes(2);
    expect(page.cleanup).not.toHaveBeenCalled();
    view.unmount();
    await waitFor(() => {
      expect(page.cleanup).toHaveBeenCalledTimes(1);
    });
  });

  it("cancels rendering during text extraction and destroys the worker on unmount", async () => {
    let rejectRender: (error: Error) => void = () => undefined;
    let finishText: (value: { items: never[] }) => void = () => undefined;
    const painting = new Promise<void>((_resolve, reject) => {
      rejectRender = reject;
    });
    const text = new Promise<{ items: never[] }>((resolve) => {
      finishText = resolve;
    });
    const cancel = jest.fn(() => {
      rejectRender(new Error("Rendering cancelled"));
    });
    const page = makePage();
    page.render.mockReturnValue({ promise: painting, cancel });
    page.getTextContent.mockReturnValue(text);
    mockGetPage.mockResolvedValue(page);
    const view = render(<PdfPreview attachment={file} />);
    await waitFor(() => {
      expect(page.render).toHaveBeenCalledTimes(1);
    });
    view.unmount();
    await act(async () => {
      finishText({ items: [] });
    });
    expect(cancel).toHaveBeenCalledTimes(1);
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(mockTextRender).not.toHaveBeenCalled();
    expect(page.cleanup).toHaveBeenCalledTimes(1);
  });

  it("exposes password-protected file fallback and retries with a fresh document task", async () => {
    const failure = new Error("Password required");
    failure.name = "PasswordException";
    mockGetDocument.mockReturnValueOnce({
      promise: Promise.reject(failure),
      destroy: mockDestroy,
    });
    render(<PdfPreview attachment={file} />);
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "password protected",
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry preview" }));
    await screen.findByRole("textbox", { name: "PDF page number" });
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(mockGetDocument).toHaveBeenCalledTimes(2);
  });
});
