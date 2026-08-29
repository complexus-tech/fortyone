/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { useQueryClient } from "@tanstack/react-query";
import type { Story } from "@/modules/stories/types";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { useAddAssociationMutation } from "./add-association-mutation";

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

const mockedUseQueryClient = jest.mocked(useQueryClient);

type AddAssociationVariables = {
  associatedStory: Story;
  fromStoryId: string;
  storyId: string;
  toStoryId: string;
  type: "related" | "blocking" | "duplicate";
};

type CapturedMutation = {
  onMutate: (variables: AddAssociationVariables) => Promise<unknown>;
};

describe("useAddAssociationMutation", () => {
  it("adds the association to the active story cache immediately", async () => {
    const currentStory = {
      associations: [],
      id: "story-1",
    } as unknown as DetailedStory;
    const associatedStory = {
      id: "story-2",
      title: "Related story",
    } as unknown as Story;
    const queryClient = {
      cancelQueries: jest.fn().mockResolvedValue(undefined),
      getQueryData: jest.fn(() => currentStory),
      invalidateQueries: jest.fn(),
      setQueryData: jest.fn(),
    };
    mockedUseQueryClient.mockReturnValue(
      queryClient as unknown as ReturnType<typeof useQueryClient>,
    );

    const mutation = useAddAssociationMutation() as unknown as CapturedMutation;
    await mutation.onMutate({
      associatedStory,
      fromStoryId: "story-1",
      storyId: "story-1",
      toStoryId: "story-2",
      type: "related",
    });

    const storyKey = storyKeys.detail("workspace", "story-1");
    expect(queryClient.cancelQueries).toHaveBeenCalledWith({
      queryKey: storyKey,
    });
    expect(queryClient.setQueryData).toHaveBeenCalledWith(storyKey, {
      ...currentStory,
      associations: [
        {
          fromStoryId: "story-1",
          id: "optimistic-association-story-1-story-2",
          story: associatedStory,
          toStoryId: "story-2",
          type: "related",
        },
      ],
    });
  });
});
