/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, cleanup, renderHook } from "@testing-library/react";
import type {
  ImportAnalysisPollResponse,
  ImportAnalysisStartResponse,
  ImportDraft,
} from "../schema";
import { pollImportAnalysis, startImportAnalysis } from "../api";
import { useImportAnalysis } from "./use-import-analysis";

jest.mock("../api", () => ({
  pollImportAnalysis: jest.fn(),
  startImportAnalysis: jest.fn(),
}));
jest.mock("./use-import-terms", () => ({
  useImportTerms: () => ({ storyTerm: "story", storyTermCapitalized: "Story" }),
}));

const startAnalysis = jest.mocked(startImportAnalysis);
const pollAnalysis = jest.mocked(pollImportAnalysis);

const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
};

const preview = (fileName: string): ImportDraft => ({
  sourceType: "json",
  sourceNamespace: null,
  summary: `${fileName} preview`,
  warnings: [],
  mapping: null,
  teams: [],
  people: [],
  labels: [],
  strategicPillars: [],
  objectives: [],
  keyResults: [],
  sprints: [],
  tasks: [],
  columns: [],
  rows: [],
  fileHash: fileName,
  fileName,
});

const completedUpload = (draft: ImportDraft): ImportAnalysisStartResponse => ({
  analysis: draft,
  fileHash: draft.fileHash,
  responseId: null,
  status: "completed",
});

const options = () => ({
  workspaceSlug: "workspace",
  onNewFile: jest.fn(),
  onUploaded: jest.fn(),
  onCompleted: jest.fn(),
});

beforeEach(() => {
  jest.resetAllMocks();
  jest.useFakeTimers();
});

afterEach(() => {
  cleanup();
  jest.clearAllTimers();
  jest.useRealTimers();
});

describe("import analysis lifecycle", () => {
  it("ignores an earlier upload that finishes after its replacement", async () => {
    const earlier = deferred<ImportAnalysisStartResponse>();
    const latest = deferred<ImportAnalysisStartResponse>();
    startAnalysis
      .mockReturnValueOnce(earlier.promise)
      .mockReturnValueOnce(latest.promise);
    const callbacks = options();
    const { result } = renderHook(() => useImportAnalysis(callbacks));

    act(() => {
      result.current.handleFile(new File(["old"], "old.json"));
    });
    act(() => {
      result.current.handleFile(new File(["new"], "new.json"));
    });
    await act(async () => {
      latest.resolve(completedUpload(preview("new.json")));
    });
    await act(async () => {
      earlier.resolve(completedUpload(preview("old.json")));
    });

    expect(result.current.draft?.fileName).toBe("new.json");
    expect(result.current.fileName).toBe("new.json");
    expect(result.current.uploadPending).toBe(false);
    expect(callbacks.onUploaded).toHaveBeenCalledTimes(1);
    expect(callbacks.onUploaded).toHaveBeenCalledWith(
      expect.objectContaining({ fileName: "new.json" }),
      "new.json",
    );
  });

  it("ignores an in-flight poll after the user resets the import", async () => {
    const pendingPoll = deferred<ImportAnalysisPollResponse>();
    const uploadedDraft = preview("source.json");
    startAnalysis.mockResolvedValue({
      ...completedUpload(uploadedDraft),
      responseId: "resp_source",
      status: "queued",
    });
    pollAnalysis.mockReturnValue(pendingPoll.promise);
    const callbacks = options();
    const { result } = renderHook(() => useImportAnalysis(callbacks));
    await act(async () => {
      result.current.handleFile(new File(["source"], "source.json"));
    });
    await act(async () => {
      jest.advanceTimersByTime(700);
    });
    expect(pollAnalysis).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.reset();
    });
    await act(async () => {
      pendingPoll.resolve({ analysis: uploadedDraft, status: "completed" });
      jest.advanceTimersByTime(60_000);
    });

    expect(result.current.draft).toBeNull();
    expect(result.current.analysisPending).toBe(false);
    expect(callbacks.onCompleted).not.toHaveBeenCalled();
    expect(pollAnalysis).toHaveBeenCalledTimes(1);
  });

  it("keeps the deterministic preview when a poll hangs past the deadline", async () => {
    const pendingPoll = deferred<ImportAnalysisPollResponse>();
    const uploadedDraft = preview("source.json");
    startAnalysis.mockResolvedValue({
      ...completedUpload(uploadedDraft),
      responseId: "resp_source",
      status: "queued",
    });
    pollAnalysis.mockReturnValue(pendingPoll.promise);
    const callbacks = options();
    const { result } = renderHook(() => useImportAnalysis(callbacks));
    await act(async () => {
      result.current.handleFile(new File(["source"], "source.json"));
    });
    await act(async () => {
      jest.advanceTimersByTime(7 * 60 * 1000);
    });

    expect(result.current.analysisPending).toBe(false);
    expect(result.current.analysisError).toBe("");
    expect(result.current.analysisNotice).toContain("longer than expected");
    expect(result.current.draft).toMatchObject({
      fileName: "source.json",
      summary: uploadedDraft.summary,
      warnings: [expect.stringContaining("deterministic preview")],
    });

    await act(async () => {
      pendingPoll.resolve({
        analysis: { ...uploadedDraft, summary: "Late AI result" },
        status: "completed",
      });
    });
    expect(result.current.draft?.summary).toBe(uploadedDraft.summary);
    expect(callbacks.onCompleted).not.toHaveBeenCalled();
  });
});
