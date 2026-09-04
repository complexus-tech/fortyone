/* global beforeEach, afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, renderHook } from "@testing-library/react";
import { createWorkspaceAction } from "@/lib/actions/create-workspace";
import { updateProfile } from "@/lib/actions/update-profile";
import { checkWorkspaceAvailability } from "@/lib/queries/check-workspace-availability";
import type { User } from "@/types/user";
import {
  getWorkspaceDraftKey,
  readWorkspaceDraft,
  saveWorkspaceDraft,
} from "./workspace-onboarding-model";
import { useWorkspaceOnboarding } from "./use-workspace-onboarding";

jest.mock("@/lib/actions/create-workspace", () => ({
  createWorkspaceAction: jest.fn(),
}));
jest.mock("@/lib/actions/update-profile", () => ({ updateProfile: jest.fn() }));
jest.mock("@/lib/queries/check-workspace-availability", () => ({
  checkWorkspaceAvailability: jest.fn(),
}));

const profile = { id: "user-1", fullName: "Ada Lovelace" } as User;
const createMock = jest.mocked(createWorkspaceAction);
const updateMock = jest.mocked(updateProfile);
const availabilityMock = jest.mocked(checkWorkspaceAvailability);
const success = { data: { slug: "acme" }, error: { message: "" } } as Awaited<
  ReturnType<typeof createWorkspaceAction>
>;
const available = {
  data: { available: true, slug: "acme" },
  error: { message: "" },
};
const deferred = <T>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
};

const prepareDraft = (user = profile) => {
  saveWorkspaceDraft(user.id, {
    ...readWorkspaceDraft(user.id, user.fullName),
    name: "Acme",
    slug: "acme",
    workType: "product",
    teamSize: "11-50",
    start: "examples",
    step: 2,
    furthestStep: 2,
  });
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.resetAllMocks();
  sessionStorage.clear();
  availabilityMock.mockResolvedValue(available);
  updateMock.mockResolvedValue({ data: profile, error: { message: "" } });
  createMock.mockResolvedValue(success);
});
afterEach(() => {
  jest.useRealTimers();
});

describe("workspace onboarding", () => {
  it("restores the user's draft before persistence and preserves choices through back navigation", async () => {
    prepareDraft();
    const { result, unmount } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated: jest.fn() }),
    );
    expect(result.current.draft.step).toBe(2);
    expect(
      JSON.parse(sessionStorage.getItem(getWorkspaceDraftKey(profile.id))!),
    ).toMatchObject({
      name: "Acme",
      teamSize: "11-50",
      workType: "product",
      start: "examples",
    });
    await act(async () => {
      jest.advanceTimersByTime(400);
    });
    act(() => {
      result.current.changeStep(0);
    });
    act(() => {
      result.current.changeName("Acme Studio");
    });
    await act(async () => {
      jest.advanceTimersByTime(400);
    });
    act(() => {
      result.current.changeStep(2);
    });
    expect(result.current.draft).toMatchObject({
      step: 2,
      furthestStep: 2,
      name: "Acme Studio",
      workType: "product",
      start: "examples",
      teamSize: "11-50",
    });
    unmount();
    const other = renderHook(() =>
      useWorkspaceOnboarding({
        profile: { ...profile, id: "user-2" },
        onCreated: jest.fn(),
      }),
    );
    expect(other.result.current.draft).toMatchObject({
      name: "",
      step: 0,
      workType: null,
    });
  });

  it("preserves an explicitly edited URL when the workspace name changes", () => {
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated: jest.fn() }),
    );
    act(() => {
      result.current.changeName("Acme Studio");
    });
    expect(result.current.draft.slug).toBe("acme-studio");
    act(() => {
      result.current.updateDraft({ slug: "my-team", slugEdited: true });
    });
    act(() => {
      result.current.changeName("A different name");
    });
    expect(result.current.draft.slug).toBe("my-team");
  });

  it("ignores availability results for a replaced URL", async () => {
    const first =
      deferred<Awaited<ReturnType<typeof checkWorkspaceAvailability>>>();
    const second =
      deferred<Awaited<ReturnType<typeof checkWorkspaceAvailability>>>();
    availabilityMock
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated: jest.fn() }),
    );
    act(() => {
      result.current.changeName("Taken");
    });
    act(() => {
      jest.advanceTimersByTime(400);
    });
    act(() => {
      result.current.changeName("Available");
    });
    act(() => {
      jest.advanceTimersByTime(400);
    });
    await act(async () => {
      second.resolve(available);
    });
    expect(result.current.availability).toBe("available");
    await act(async () => {
      first.resolve({
        data: { available: false, slug: "taken" },
        error: { message: "" },
      });
    });
    expect(result.current.availability).toBe("available");
  });

  it("shows a failed availability check as unknown instead of claiming the URL is taken", async () => {
    availabilityMock.mockRejectedValue(new Error("Network unavailable"));
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated: jest.fn() }),
    );
    act(() => {
      result.current.changeName("Acme");
    });
    await act(async () => {
      jest.advanceTimersByTime(400);
    });
    expect(result.current.availability).toBe("unknown");
  });

  it("does not create a workspace when profile save returns an error", async () => {
    prepareDraft();
    updateMock.mockResolvedValue({
      data: null,
      error: { message: "Could not save your name" },
    } as Awaited<ReturnType<typeof updateProfile>>);
    const onCreated = jest.fn();
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated }),
    );
    await act(async () => {
      await result.current.createWorkspace();
    });
    expect(createMock).not.toHaveBeenCalled();
    expect(onCreated).not.toHaveBeenCalled();
    expect(result.current.error).toBe("Could not save your name");
    expect(result.current.draft.step).toBe(2);
  });

  it("keeps the draft after a failed workspace request and reuses a successfully saved profile on retry", async () => {
    prepareDraft();
    createMock.mockResolvedValueOnce({
      data: null,
      error: { message: "Please try again" },
    } as Awaited<ReturnType<typeof createWorkspaceAction>>);
    const onCreated = jest.fn();
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated }),
    );
    await act(async () => {
      await result.current.createWorkspace();
    });
    expect(result.current.error).toBe("Please try again");
    expect(
      sessionStorage.getItem(getWorkspaceDraftKey(profile.id)),
    ).not.toBeNull();
    expect(onCreated).not.toHaveBeenCalled();
    await act(async () => {
      await result.current.createWorkspace();
    });
    expect(updateMock).toHaveBeenCalledTimes(1);
    expect(createMock).toHaveBeenCalledTimes(2);
    expect(createMock).toHaveBeenLastCalledWith({
      name: "Acme",
      slug: "acme",
      teamSize: "11-50",
      workType: "product",
      includeExamples: true,
    });
    expect(onCreated).toHaveBeenCalledWith("acme", "examples");
    expect(sessionStorage.getItem(getWorkspaceDraftKey(profile.id))).toBeNull();
  });

  it("does not submit twice while a workspace is being created", async () => {
    prepareDraft();
    const pending =
      deferred<Awaited<ReturnType<typeof createWorkspaceAction>>>();
    createMock.mockReturnValue(pending.promise);
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated: jest.fn() }),
    );
    let request!: Promise<void>;
    await act(async () => {
      request = result.current.createWorkspace();
      await result.current.createWorkspace();
    });
    expect(createMock).toHaveBeenCalledTimes(1);
    expect(result.current.isLoading).toBe(true);
    await act(async () => {
      pending.resolve(success);
      await request;
    });
  });

  it("retries a failed handoff without creating the already saved workspace again", async () => {
    prepareDraft();
    const onCreated = jest.fn().mockImplementationOnce(() => {
      throw new Error("Navigation failed");
    });
    const { result } = renderHook(() =>
      useWorkspaceOnboarding({ profile, onCreated }),
    );
    await act(async () => {
      await result.current.createWorkspace("empty");
    });
    expect(createMock).toHaveBeenCalledWith(
      expect.objectContaining({ includeExamples: false }),
    );
    await act(async () => {
      await result.current.createWorkspace();
    });
    expect(createMock).toHaveBeenCalledTimes(1);
    expect(onCreated).toHaveBeenLastCalledWith("acme", "empty");
  });
});
