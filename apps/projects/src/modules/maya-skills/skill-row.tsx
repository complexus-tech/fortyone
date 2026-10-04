import { DeleteIcon, EditIcon, MoreHorizontalIcon, WorkflowIcon } from "icons";
import { Box, Button, Flex, Menu, Text } from "ui";
import type { MayaSkill } from "./types";

export const MayaSkillRow = ({
  skill,
  onEdit,
  onDelete,
  editDisabled = false,
}: {
  skill: MayaSkill;
  onEdit: (skill: MayaSkill) => void;
  onDelete: (skill: MayaSkill) => void;
  editDisabled?: boolean;
}) => (
  <Flex align="center" className="gap-3 py-3" justify="between">
    <WorkflowIcon aria-hidden className="text-icon h-5 w-auto shrink-0" />
    <Box className="min-w-0 flex-1">
      <Text className="break-words" fontWeight="medium">
        {skill.name}
      </Text>
      {skill.description ? (
        <Text className="mt-1 break-words" color="muted">
          {skill.description}
        </Text>
      ) : null}
    </Box>
    <Menu>
      <Menu.Button asChild>
        <Button
          aria-label={`Actions for ${skill.name}`}
          asIcon
          color="tertiary"
          size="sm"
          type="button"
          variant="naked"
        >
          <MoreHorizontalIcon />
        </Button>
      </Menu.Button>
      <Menu.Items align="end" className="w-40">
        <Menu.Item
          disabled={editDisabled}
          onSelect={() => {
            onEdit(skill);
          }}
        >
          <EditIcon />
          Edit skill
        </Menu.Item>
        <Menu.Separator />
        <Menu.Item
          className="text-danger"
          onSelect={() => {
            onDelete(skill);
          }}
        >
          <DeleteIcon />
          Delete skill
        </Menu.Item>
      </Menu.Items>
    </Menu>
  </Flex>
);
