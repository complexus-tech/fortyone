"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";
import { Flex, Text } from "ui";
import { useUserRole } from "@/hooks/role";
import { useMembers } from "@/lib/hooks/members";
import type {
  BoardPropertyProviderProps,
  BoardPropertyStory,
  StoryPropertyBadgesProps,
} from "@/shared/story/board-property-slots";
import type { CustomField } from "./types";
import { useCustomFieldStoryValues, useVisibleTeamCustomFields } from "./hooks";
import { FieldValuePicker } from "./field-value-picker";
import { isCustomFieldValueSet } from "./value-utils";

type DisplayContext = {
  fields: CustomField[];
  values: Map<string, Map<string, string | null>>;
  versions: Map<string, number>;
  people: { id: string; fullName: string }[];
  onReload: (
    storyId: string,
    fieldId: string,
  ) => Promise<{ value: string | null; version: number }>;
};
const CustomFieldsDisplayContext = createContext<DisplayContext | null>(null);

const collectStories = (stories: BoardPropertyStory[]): BoardPropertyStory[] =>
  stories.flatMap((story) => [
    story,
    ...collectStories(story.subStories ?? []),
  ]);

const LoadedDisplayProvider = ({
  children,
  stories,
  selectedIds,
}: {
  children: ReactNode;
  stories: BoardPropertyStory[];
  selectedIds: string[];
}) => {
  const visibleStories = collectStories(stories);
  const queries = useVisibleTeamCustomFields(
    visibleStories.map((story) => story.teamId),
  );
  const snapshot = useCustomFieldStoryValues(
    visibleStories.map((story) => story.id),
  );
  const { data: people = [] } = useMembers();
  const fields = queries
    .flatMap((query) => query.data ?? [])
    .filter((field) => selectedIds.includes(field.id));
  return (
    <CustomFieldsDisplayContext.Provider
      value={{
        fields,
        people,
        values: new Map(
          Array.from(snapshot.items.values()).map((item) => [
            item.storyId,
            new Map(item.values.map((value) => [value.fieldId, value.value])),
          ]),
        ),
        versions: new Map(
          Array.from(snapshot.items.values()).map((item) => [
            item.storyId,
            item.version,
          ]),
        ),
        onReload: async (storyId, fieldId) => {
          const latest = await snapshot.refetchStory(storyId);
          return {
            value:
              latest.values.find((item) => item.fieldId === fieldId)?.value ??
              null,
            version: latest.version,
          };
        },
      }}
    >
      {snapshot.isError || queries.some((query) => query.isError) ? (
        <Text className="px-2 py-3" color="danger" role="alert">
          Custom fields could not be loaded.
        </Text>
      ) : null}
      {children}
    </CustomFieldsDisplayContext.Provider>
  );
};

export const CustomFieldsBoardProvider = ({
  children,
  stories,
  selectedIds = [],
}: BoardPropertyProviderProps) => {
  if (!selectedIds.length) return children;
  return (
    <LoadedDisplayProvider
      selectedIds={selectedIds.slice(0, 3)}
      stories={stories}
    >
      {children}
    </LoadedDisplayProvider>
  );
};

export const StoryCustomFieldBadges = ({
  storyId,
  teamId,
  asList = false,
  disabled = false,
}: StoryPropertyBadgesProps) => {
  const context = useContext(CustomFieldsDisplayContext);
  const { userRole } = useUserRole();
  const [editingFields, setEditingFields] = useState(new Set<string>());
  if (!context) return null;
  const values = context.values.get(storyId);
  const fields = context.fields.filter((field) => {
    const value = values?.get(field.id);
    return (
      field.teamId === teamId &&
      (isCustomFieldValueSet(value) || editingFields.has(field.id))
    );
  });
  if (!fields.length) return null;
  const version = context.versions.get(storyId);
  const readOnly =
    disabled ||
    (userRole !== "admin" && userRole !== "member") ||
    version === undefined;
  return (
    <Flex
      className={
        asList
          ? "mt-2 max-w-[36rem] flex-wrap gap-2 md:mt-0"
          : "mt-3 flex-wrap gap-2"
      }
    >
      {fields.map((field) => (
        <FieldValuePicker
          disabled={readOnly}
          field={field}
          key={field.id}
          onOpenChange={(open) => {
            setEditingFields((current) => {
              const next = new Set(current);
              if (open) next.add(field.id);
              else next.delete(field.id);
              return next;
            });
          }}
          onReload={() => context.onReload(storyId, field.id)}
          people={context.people}
          size="xs"
          storyId={storyId}
          value={values?.get(field.id) ?? null}
          variant="outline"
          version={version}
        />
      ))}
    </Flex>
  );
};
