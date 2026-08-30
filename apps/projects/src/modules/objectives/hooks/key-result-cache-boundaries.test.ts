/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useQueryClient } from "@tanstack/react-query";
import { keyResultKeys } from "@/constants/keys";
import { objectiveKeys } from "../constants";
import type { KeyResult, KeyResultUpdate } from "../types";
import { useDeleteKeyResultMutation } from "./use-delete-key-result-mutation";
import { useUpdateKeyResultMutation } from "./use-update-key-result-mutation";

jest.mock("@tanstack/react-query", () => ({
  useMutation: jest.fn((options) => options),
  useQueryClient: jest.fn(),
}));

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(() => ({ analytics: { track: jest.fn() } })),
  useWorkspacePath: jest.fn(() => ({ workspaceSlug: "workspace" })),
}));

jest.mock("../actions/delete-key-result", () => ({
  deleteKeyResult: jest.fn(),
}));

jest.mock("../actions/update-key-result", () => ({
  updateKeyResult: jest.fn(),
}));

const mockedUseQueryClient = jest.mocked(useQueryClient);

const keyResult: KeyResult = {
  contributors: [],
  createdAt: "2026-08-30T00:00:00.000Z",
  createdBy: "member-1",
  currentValue: 20,
  endDate: "2026-12-31",
  id: "key-result-1",
  lead: null,
  measurementType: "percentage",
  name: "Launch a reliable roadmap",
  objectiveId: "objective-1",
  sequenceId: 1,
  startDate: "2026-01-01",
  startValue: 0,
  targetValue: 100,
  updatedAt: "2026-08-30T00:00:00.000Z",
};

const createQueryClient = () => ({
  cancelQueries: jest.fn().mockResolvedValue(undefined),
  getQueryData: jest.fn(() => [keyResult]),
  invalidateQueries: jest.fn(),
  setQueryData: jest.fn(),
  setQueriesData: jest.fn(),
});

type UpdateMutation = {
  onMutate: (variables: {
    data: KeyResultUpdate;
    keyResultId: string;
    objectiveId: string;
  }) => Promise<unknown>;
  onSuccess: (
    response: unknown,
    variables: {
      data: KeyResultUpdate;
      keyResultId: string;
      objectiveId: string;
    },
  ) => void;
};

type DeleteMutation = {
  onMutate: (variables: {
    keyResultId: string;
    objectiveId: string;
  }) => Promise<unknown>;
};

describe("key result cache boundaries", () => {
  it("updates only the objective-owned cache optimistically", async () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useUpdateKeyResultMutation() as unknown as UpdateMutation;
    await mutation.onMutate({
      data: { name: "Keep strategy current" },
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
    });

    const objectiveKey = objectiveKeys.keyResults(
      "workspace",
      keyResult.objectiveId,
    );
    expect(queryClient.cancelQueries).toHaveBeenCalledWith({
      queryKey: objectiveKey,
    });
    expect(queryClient.setQueryData).toHaveBeenCalledWith(
      objectiveKey,
      expect.any(Function),
    );
    expect(queryClient.setQueriesData).not.toHaveBeenCalled();
  });

  it("removes only the objective-owned cache optimistically", async () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useDeleteKeyResultMutation() as unknown as DeleteMutation;
    await mutation.onMutate({
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
    });

    const objectiveKey = objectiveKeys.keyResults(
      "workspace",
      keyResult.objectiveId,
    );
    expect(queryClient.cancelQueries).toHaveBeenCalledWith({
      queryKey: objectiveKey,
    });
    expect(queryClient.setQueryData).toHaveBeenCalledWith(
      objectiveKey,
      expect.any(Function),
    );
    expect(queryClient.setQueriesData).not.toHaveBeenCalled();
  });

  it("reconciles the workspace view through its shared query identity", () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useUpdateKeyResultMutation() as unknown as UpdateMutation;
    mutation.onSuccess(undefined, {
      data: { name: "Keep strategy current" },
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
    });

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: keyResultKeys.all("workspace"),
    });
  });
});
