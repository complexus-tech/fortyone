import type { ComponentType, ReactNode } from "react";
import { Box, Button, Checkbox, Flex, Text, Wrapper } from "ui";
import { AiIcon, InfoIcon } from "icons";

type SubstorySuggestion =
  | {
      title?: string;
    }
  | undefined;

type SuggestionRowProps = {
  children: ReactNode;
  className?: string;
};

type StoryTerms = {
  plural: string;
  pluralCapitalized: string;
  singular: string;
  singularCapitalized: string;
};

type SubstorySuggestionsProps = {
  canCreateSuggestedSubstories: boolean;
  isCreatingSuggestedSubstories: boolean;
  isShowingSuggestionError: boolean;
  onCancelSuggestions: () => void;
  onCreateSelectedSubstories: () => void;
  onRequestSuggestions: () => void;
  onToggleSelectedSubstory: (title: string) => void;
  selectedSubstories: ReadonlySet<string>;
  showSuggestions: boolean;
  SuggestionRow: ComponentType<SuggestionRowProps>;
  suggestedSubstories: SubstorySuggestion[] | undefined;
  terms: StoryTerms;
};

/**
 * Renders only the suggestion result states. Streaming, validation, selection,
 * and mutation transitions remain in useSubstorySuggestions so this component
 * can stay a pure view over an explicit flow contract.
 */
export const SubstorySuggestions = ({
  canCreateSuggestedSubstories,
  isCreatingSuggestedSubstories,
  isShowingSuggestionError,
  onCancelSuggestions,
  onCreateSelectedSubstories,
  onRequestSuggestions,
  onToggleSelectedSubstory,
  selectedSubstories,
  showSuggestions,
  SuggestionRow,
  suggestedSubstories,
  terms,
}: SubstorySuggestionsProps) => {
  if (isShowingSuggestionError) {
    return (
      <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 my-2.5 flex items-center justify-between gap-2 border p-4">
        <Flex align="center" gap={2}>
          <InfoIcon className="text-warning dark:text-warning" />
          <Text>
            Maya could not finish generating sub {terms.plural}. Try again to
            create a complete set of suggestions.
          </Text>
        </Flex>
        <Button color="warning" onClick={onRequestSuggestions}>
          Try again
        </Button>
      </Wrapper>
    );
  }

  if (!showSuggestions || !suggestedSubstories) return null;

  if (suggestedSubstories.length === 0) {
    return (
      <Box className="my-2.5">
        <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 flex items-center justify-between gap-2 border p-4">
          <Flex align="center" gap={2}>
            <InfoIcon className="text-warning dark:text-warning" />
            <Text>
              Could not generate sub {terms.plural}, make sure your parent{" "}
              {terms.singular} is actionable
            </Text>
          </Flex>
          <Button color="warning" onClick={onRequestSuggestions}>
            Try again
          </Button>
        </Wrapper>
      </Box>
    );
  }

  return (
    <Box className="my-2.5">
      <Box className="border-border rounded-2xl border-[0.5px]">
        {suggestedSubstories.map((substory) => {
          if (!substory?.title) return null;

          return (
            <SuggestionRow
              className="gap-6 px-2 last-of-type:border-b-0 md:px-4"
              key={substory.title}
            >
              <Flex align="center" className="flex-1" gap={2}>
                <AiIcon className="shrink-0" />
                <Text
                  className="line-clamp-1"
                  color={
                    selectedSubstories.has(substory.title) ? undefined : "muted"
                  }
                >
                  {substory.title}
                </Text>
              </Flex>
              <Checkbox
                checked={selectedSubstories.has(substory.title)}
                className="shrink-0"
                onCheckedChange={() => {
                  onToggleSelectedSubstory(substory.title!);
                }}
              />
            </SuggestionRow>
          );
        })}
      </Box>
      <Flex className="mt-2" gap={2} justify="end">
        <Button color="tertiary" onClick={onCancelSuggestions} variant="naked">
          Cancel
        </Button>
        <Button
          disabled={
            selectedSubstories.size === 0 ||
            !canCreateSuggestedSubstories ||
            isCreatingSuggestedSubstories
          }
          onClick={onCreateSelectedSubstories}
        >
          Create {selectedSubstories.size} Sub{" "}
          {selectedSubstories.size === 1
            ? terms.singularCapitalized
            : terms.pluralCapitalized}
        </Button>
      </Flex>
    </Box>
  );
};
