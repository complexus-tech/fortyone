/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { experimental_useObject as useObject } from "@ai-sdk/react";
import { act, renderHook } from "@testing-library/react";
import { substoryGenerationSchema } from "@/modules/stories/public/substory-generation";
import { useCreateStoryMutation } from "../hooks/create-mutation";
import { useSubstorySuggestions } from "./use-substory-suggestions";

jest.mock("@ai-sdk/react", () => ({
  experimental_useObject: jest.fn(),
}));

jest.mock("../hooks/create-mutation", () => ({
  useCreateStoryMutation: jest.fn(),
}));

type SuggestionStreamState = {
  error?: Error;
  isLoading: boolean;
  object?: {
    substories?: { title?: string }[];
  };
};

const suggestions = {
  substories: [{ title: "Plan onboarding" }, { title: "Invite teammates" }],
};

const submit = jest.fn();
const createStory = jest.fn();
const mockedUseObject = jest.mocked(useObject);
const mockedUseCreateStoryMutation = jest.mocked(useCreateStoryMutation);
let streamState: SuggestionStreamState;

const renderSuggestions = () =>
  renderHook(() =>
    useSubstorySuggestions({
      defaultStatusId: "status-default",
      storyId: "story-1",
      teamId: "team-1",
      workspaceSlug: "acme",
    }),
  );

describe("useSubstorySuggestions", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    streamState = {
      isLoading: false,
      object: suggestions,
    };
    mockedUseObject.mockImplementation(
      () =>
        ({
          clear: jest.fn(),
          error: streamState.error,
          isLoading: streamState.isLoading,
          object: streamState.object,
          stop: jest.fn(),
          submit,
        }) as ReturnType<typeof useObject>,
    );
    mockedUseCreateStoryMutation.mockReturnValue({
      isPending: false,
      mutate: createStory,
    } as unknown as ReturnType<typeof useCreateStoryMutation>);
  });

  it("submits the scoped locator and restores the default selection on retry", () => {
    const { result } = renderSuggestions();

    act(() => {
      result.current.toggleSelectedSubstory("Plan onboarding");
    });
    expect([...result.current.selectedSubstories]).toEqual([
      "Invite teammates",
    ]);

    act(() => {
      result.current.cancelSuggestions();
    });
    expect(result.current.showSuggestions).toBe(false);
    expect(result.current.selectedSubstories).toEqual(new Set());

    act(() => {
      result.current.requestSuggestions();
    });

    expect(submit).toHaveBeenCalledWith({
      storyId: "story-1",
      workspaceSlug: "acme",
    });
    expect([...result.current.selectedSubstories]).toEqual([
      "Plan onboarding",
      "Invite teammates",
    ]);
    expect(mockedUseObject).toHaveBeenLastCalledWith({
      api: "/api/suggest-substories",
      schema: substoryGenerationSchema,
    });
  });

  it.each<[string, SuggestionStreamState]>([
    ["a partial stream", { isLoading: true, object: suggestions }],
    [
      "a schema-invalid completed stream",
      { isLoading: false, object: { substories: [{ title: " " }] } },
    ],
    [
      "a failed stream",
      {
        error: new Error("provider disconnected"),
        isLoading: false,
        object: suggestions,
      },
    ],
  ])("does not create suggestions from %s", (_, state) => {
    streamState = state;
    const { result } = renderSuggestions();

    expect(result.current.canCreateSuggestedSubstories).toBe(false);

    act(() => {
      result.current.createSelectedSubstories();
    });

    expect(createStory).not.toHaveBeenCalled();
  });

  it("creates only the selected, complete suggestions with the parent defaults", () => {
    const { result } = renderSuggestions();

    act(() => {
      result.current.toggleSelectedSubstory("Plan onboarding");
    });
    act(() => {
      result.current.createSelectedSubstories();
    });

    expect(createStory).toHaveBeenCalledTimes(1);
    expect(createStory).toHaveBeenCalledWith({
      parentId: "story-1",
      priority: "No Priority",
      statusId: "status-default",
      teamId: "team-1",
      title: "Invite teammates",
    });
    expect(result.current.showSuggestions).toBe(false);
    expect(result.current.selectedSubstories).toEqual(new Set());
  });
});
