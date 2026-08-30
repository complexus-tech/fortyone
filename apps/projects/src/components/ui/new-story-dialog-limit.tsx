import { CrownIcon } from "icons";
import { Button, Dialog, Divider, Flex, Text, Wrapper } from "ui";

export const NewStoryDialogLimit = ({
  billingHref,
  canUpgrade,
  isOpen,
  maxStories,
  onClose,
  planName,
  storyTerm,
  storyTermPlural,
  storyTermPluralCapitalized,
  totalStories,
}: {
  billingHref: string;
  canUpgrade: boolean;
  isOpen: boolean;
  maxStories: number;
  onClose: () => void;
  planName: string;
  storyTerm: string;
  storyTermPlural: string;
  storyTermPluralCapitalized: string;
  totalStories: number;
}) => (
  <Dialog open={isOpen}>
    <Dialog.Content hideClose>
      <Dialog.Header className="flex items-center gap-2 px-6 pt-6 text-xl">
        <CrownIcon className="text-warning relative -top-px h-6" />
        <Dialog.Title>{storyTerm} Limit Reached</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Text className="mb-4 dark:font-normal" color="muted">
          You&apos;ve reached the limit of {maxStories} {storyTermPlural} on
          your {planName} plan. {canUpgrade ? "Upgrade" : "Ask your admin"} to
          create unlimited {storyTermPlural} and unlock premium features.
        </Text>
        <Wrapper className="bg-surface-elevated/60">
          <Flex align="center" gap={3} justify="between">
            <Text color="muted">Current plan:</Text>
            <Text transform="capitalize">{planName}</Text>
          </Flex>
          <Divider className="my-3" />
          <Flex align="center" gap={3} justify="between">
            <Text color="muted">{storyTermPluralCapitalized}:</Text>
            <Text color="primary">
              {totalStories}/{maxStories}
            </Text>
          </Flex>
        </Wrapper>
        {canUpgrade ? (
          <Button
            align="center"
            className="mt-4 border-0"
            fullWidth
            href={billingHref}
            rounded="lg"
            size="lg"
          >
            Upgrade now
          </Button>
        ) : null}
        <Button
          align="center"
          className="mt-3 mb-2 border-[0.5px]"
          color="tertiary"
          fullWidth
          onClick={onClose}
          rounded="lg"
          size="lg"
        >
          Maybe later
        </Button>
      </Dialog.Body>
    </Dialog.Content>
  </Dialog>
);
