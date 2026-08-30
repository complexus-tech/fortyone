/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { objectiveKeys } from "@/shared/objectives/keys";
import { alignObjective } from "./api";
import { strategyKeys, useAlignObjectiveMutation } from "./hooks";
import type { StrategyMap } from "./types";

jest.mock("@tanstack/react-query", () => ({
  useMutation: jest.fn((options) => options),
  useQueryClient: jest.fn(),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: jest.fn(),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: jest.fn(),
}));

jest.mock("./api", () => ({
  alignObjective: jest.fn(),
  getStrategyMap: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: { error: jest.fn() },
}));

const strategy: StrategyMap = {
  description: null,
  pillars: [
    {
      description: null,
      id: "pillar-a",
      name: "Pillar A",
      objectiveIds: ["objective-1"],
      orderIndex: 0,
    },
    {
      description: null,
      id: "pillar-b",
      name: "Pillar B",
      objectiveIds: [],
      orderIndex: 1,
    },
  ],
  ultimateGoal: "Build a durable business",
};

type AlignMutation = {
  mutationFn: (variables: {
    objectiveId: string;
    pillarId: string | null;
  }) => Promise<unknown>;
  onError: (
    error: Error,
    variables: { objectiveId: string; pillarId: string | null },
    context?: { previousStrategy?: StrategyMap },
  ) => void;
  onMutate: (variables: {
    objectiveId: string;
    pillarId: string | null;
  }) => Promise<{ previousStrategy?: StrategyMap }>;
  onSettled: () => void;
};

const createQueryClient = () => ({
  cancelQueries: jest.fn().mockResolvedValue(undefined),
  getQueryData: jest.fn(() => strategy),
  invalidateQueries: jest.fn(),
  setQueryData: jest.fn(),
});

const mutationVariables = {
  objectiveId: "objective-1",
  pillarId: "pillar-b",
};

describe("useAlignObjectiveMutation", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useWorkspacePath).mockReturnValue({
      workspaceSlug: "north-star",
    } as ReturnType<typeof useWorkspacePath>);
    jest.mocked(useSession).mockReturnValue({
      data: { session: { token: "token" } },
    } as never);
  });

  it("optimistically reconciles the strategy cache through the shared contract", async () => {
    const queryClient = createQueryClient();
    jest
      .mocked(useQueryClient)
      .mockReturnValue(
        queryClient as unknown as ReturnType<typeof useQueryClient>,
      );
    const mutation = useAlignObjectiveMutation() as unknown as AlignMutation;
    let optimisticStrategy: StrategyMap | undefined;
    queryClient.setQueryData.mockImplementation(
      (
        _queryKey: ReturnType<typeof strategyKeys.map>,
        updater: (value: StrategyMap | undefined) => StrategyMap | undefined,
      ) => {
        optimisticStrategy = updater(strategy);
      },
    );

    const context = await mutation.onMutate(mutationVariables);

    expect(queryClient.cancelQueries).toHaveBeenCalledWith({
      queryKey: strategyKeys.map("north-star"),
    });
    expect(context).toEqual({ previousStrategy: strategy });
    expect(optimisticStrategy?.pillars[0]?.objectiveIds).toEqual([]);
    expect(optimisticStrategy?.pillars[1]?.objectiveIds).toEqual([
      "objective-1",
    ]);
  });

  it("uses the same cache identities for rollback and cross-surface reconciliation", async () => {
    const queryClient = createQueryClient();
    const onAlignmentSettled = jest.fn();
    jest
      .mocked(useQueryClient)
      .mockReturnValue(
        queryClient as unknown as ReturnType<typeof useQueryClient>,
      );
    jest.mocked(alignObjective).mockResolvedValue({ data: null } as never);
    const mutation = useAlignObjectiveMutation({
      onAlignmentSettled,
    }) as unknown as AlignMutation;

    await mutation.mutationFn(mutationVariables);
    mutation.onError(new Error("Update failed"), mutationVariables, {
      previousStrategy: strategy,
    });
    mutation.onSettled();

    expect(alignObjective).toHaveBeenCalledWith(
      "objective-1",
      "pillar-b",
      expect.objectContaining({ workspaceSlug: "north-star" }),
    );
    expect(queryClient.setQueryData).toHaveBeenCalledWith(
      strategyKeys.map("north-star"),
      strategy,
    );
    expect(toast.error).toHaveBeenCalledWith("Strategy could not be updated", {
      description: "Update failed",
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: strategyKeys.map("north-star"),
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: objectiveKeys.list("north-star"),
    });
    expect(onAlignmentSettled).toHaveBeenCalledTimes(1);
  });
});
