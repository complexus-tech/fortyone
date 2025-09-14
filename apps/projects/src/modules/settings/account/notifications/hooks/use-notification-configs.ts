import { useMemo } from "react";
import { type NotificationType } from "@/modules/notifications/types";
import { useFeatures, useTerminology } from "@/hooks";

type NotificationConfig = {
  type: NotificationType;
  title: string;
  description: string;
};

export function useNotificationConfigs(): NotificationConfig[] {
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();

  return useMemo(() => {
    const configs: NotificationConfig[] = [
      {
        type: "story_update",
        title: `${getTermDisplay("storyTerm", { capitalize: true })} updates`,
        description: `Get notified when a ${getTermDisplay("storyTerm")} you're involved with is updated`,
      },
      {
        type: "comment_reply",
        title: "Comments",
        description: `Get notified when someone comments on your ${getTermDisplay("storyTerm", { variant: "plural" })}`,
      },
      {
        type: "mention",
        title: "Mentions",
        description: `Get notified when someone mentions you in a comment or ${getTermDisplay("storyTerm")}`,
      },
      {
        type: "story_comment",
        title: `${getTermDisplay("storyTerm", { capitalize: true })} comments`,
        description: `Get notified when comments are added to ${getTermDisplay("storyTerm", { variant: "plural" })}`,
      },
    ];

    if (features.objectiveEnabled) {
      configs.push({
        type: "objective_update",
        title: `${getTermDisplay("objectiveTerm", { capitalize: true })} updates`,
        description: `Get notified when ${getTermDisplay("objectiveTerm", { variant: "plural" })} are updated`,
      });
    }

    if (features.keyResultEnabled) {
      configs.push({
        type: "key_result_update",
        title: `${getTermDisplay("keyResultTerm", { capitalize: true })} updates`,
        description: `Get notified about updates to ${getTermDisplay("keyResultTerm", { variant: "plural" })}`,
      });
    }

    configs.push({
      type: "overdue_stories",
      title: `Overdue ${getTermDisplay("storyTerm", { variant: "plural", capitalize: true })}`,
      description: `Get notified when ${getTermDisplay("storyTerm", { variant: "plural" })} become overdue`,
    });

    return configs;
  }, [getTermDisplay, features]);
}
