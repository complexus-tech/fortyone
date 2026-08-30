import { substoryGenerationSchema } from "@/modules/stories/public/substory-generation";

type SubstorySuggestionState = {
  error: unknown;
  isLoading: boolean;
  object: unknown;
};

export const isSubstorySuggestionReadyToCreate = ({
  error,
  isLoading,
  object,
}: SubstorySuggestionState) =>
  !isLoading && !error && substoryGenerationSchema.safeParse(object).success;
