/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useQueryClient } from "@tanstack/react-query";
import { storyKeys } from "@/modules/stories/constants";
import { useAddAssociationMutation } from "./add-association-mutation";
import { useRemoveAssociationMutation } from "./remove-association-mutation";
import { useUpdateAssociationMutation } from "./update-association-mutation";
import { useUpdateLabelsMutation } from "./update-labels-mutation";

jest.mock("@tanstack/react-query", () => ({
  useMutation: jest.fn((options) => options),
  useQueryClient: jest.fn(),
}));

jest.mock("@/hooks", () => ({
  useWorkspacePath: jest.fn(() => ({ workspaceSlug: "workspace" })),
}));

jest.mock("../actions/add-association", () => ({
  addAssociationAction: jest.fn(),
}));

jest.mock("../actions/remove-association", () => ({
  removeAssociationAction: jest.fn(),
}));

jest.mock("../actions/update-association", () => ({
  updateAssociationAction: jest.fn(),
}));

jest.mock("../actions/update-labels", () => ({
  updateLabelsAction: jest.fn(),
}));

const mockedUseQueryClient = jest.mocked(useQueryClient);

type CapturedMutation = {
  onSuccess?: (
    data: { data: null },
    variables: unknown,
    context: unknown,
  ) => void;
};

const createQueryClient = () => ({
  cancelQueries: jest.fn(),
  getQueryCache: jest.fn(() => ({ findAll: jest.fn(() => []) })),
  getQueryData: jest.fn(),
  invalidateQueries: jest.fn(),
  setQueryData: jest.fn(),
  setQueriesData: jest.fn(),
});

describe("story mutation activity invalidation", () => {
  it("invalidates story activities after labels update", () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useUpdateLabelsMutation() as unknown as CapturedMutation;

    mutation.onSuccess?.(
      { data: null },
      { labels: ["label-1"], storyId: "story-1" },
      undefined,
    );

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("workspace", "story-1"),
      refetchType: "all",
    });
  });

  it("invalidates both story activity feeds after association add", () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useAddAssociationMutation() as unknown as CapturedMutation;

    mutation.onSuccess?.(
      { data: null },
      { fromStoryId: "story-1", toStoryId: "story-2", type: "blocking" },
      undefined,
    );

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("workspace", "story-1"),
      refetchType: "all",
    });
    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("workspace", "story-2"),
      refetchType: "all",
    });
  });

  it("invalidates story activities after association update", () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation =
      useUpdateAssociationMutation() as unknown as CapturedMutation;

    mutation.onSuccess?.(
      { data: null },
      {
        associationId: "association-1",
        fromStoryId: "story-1",
        storyId: "story-1",
        toStoryId: "story-2",
        type: "blocking",
      },
      undefined,
    );

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("workspace", "story-1"),
      refetchType: "all",
    });
  });

  it("invalidates story activities after association removal", () => {
    const queryClient = createQueryClient();
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation =
      useRemoveAssociationMutation() as unknown as CapturedMutation;

    mutation.onSuccess?.(
      { data: null },
      { associationId: "association-1", storyId: "story-1" },
      undefined,
    );

    expect(queryClient.invalidateQueries).toHaveBeenCalledWith({
      queryKey: storyKeys.activitiesInfinite("workspace", "story-1"),
      refetchType: "all",
    });
  });
});
