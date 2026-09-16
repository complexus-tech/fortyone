import type { DetailedStory } from "@/modules/stories/types";
import { differenceInCalendarDays, addDays } from "date-fns";
import { Text, Badge } from "@/components/ui";
import { DetailTitle } from "@/modules/entity-details/title";
import { getStory } from "@/modules/stories/queries/get-story";
import { useUpdateStoryMutation } from "../hooks/use-update-story-mutation";

export const Title = ({ story }: { story: DetailedStory }) => {
  const mutation = useUpdateStoryMutation();
  const isDeleted = story.deletedAt !== null;
  const daysLeft = story.deletedAt
    ? Math.max(
        0,
        differenceInCalendarDays(
          addDays(new Date(story.deletedAt), 30),
          new Date(),
        ),
      )
    : 0;
  return (
    <DetailTitle
      title={story.title}
      draftKey={`story:${story.id}:title`}
      onSave={
        isDeleted
          ? undefined
          : async (title, baseline) => {
              const latest = await getStory(story.id);
              if (latest.title !== baseline && latest.title !== title)
                throw new Error(
                  "This title changed elsewhere. Your draft is preserved; review the latest title before replacing it.",
                );
              if (latest.title !== title)
                await mutation.mutateAsync({
                  storyId: story.id,
                  payload: { title },
                });
            }
      }
    >
      {isDeleted ? (
        <Badge color="tertiary" className="mb-3">
          <Text>{daysLeft} days left in bin</Text>
        </Badge>
      ) : story.archivedAt ? (
        <Badge color="tertiary" className="mb-3">
          <Text>Archived</Text>
        </Badge>
      ) : null}
    </DetailTitle>
  );
};
