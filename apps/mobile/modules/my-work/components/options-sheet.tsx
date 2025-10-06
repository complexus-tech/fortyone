import React from "react";
import { BottomSheetModal } from "@/components/ui";
import {
  Text,
  HStack,
  Spacer,
  ContextMenu,
  Button,
  Image,
  VStack,
} from "@expo/ui/swift-ui";
import { colors } from "@/constants";
import { opacity } from "@expo/ui/swift-ui/modifiers";

export const OptionsSheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  return (
    <BottomSheetModal
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={28}
    >
      <VStack spacing={20}>
        <HStack>
          <Text>Grouping</Text>
          <Spacer />
          <HStack spacing={2}>
            <ContextMenu>
              <ContextMenu.Trigger>
                <HStack spacing={3}>
                  <Text color={colors.dark.DEFAULT}>Status</Text>
                  <Image
                    systemName="chevron.up.chevron.down"
                    modifiers={[opacity(0.6)]}
                    color={colors.dark.DEFAULT}
                    size={14}
                  />
                </HStack>
              </ContextMenu.Trigger>
              <ContextMenu.Items>
                <Button>Status</Button>
                <Button>Priority</Button>
                <Button>Assignee</Button>
              </ContextMenu.Items>
            </ContextMenu>
          </HStack>
        </HStack>
        <HStack>
          <Text>Ordering</Text>
          <Spacer />
          <HStack spacing={2}>
            <ContextMenu>
              <ContextMenu.Trigger>
                <HStack spacing={3}>
                  <Text color={colors.dark.DEFAULT}>Created</Text>
                  <Image
                    systemName="chevron.up.chevron.down"
                    modifiers={[opacity(0.6)]}
                    color={colors.dark.DEFAULT}
                    size={14}
                  />
                </HStack>
              </ContextMenu.Trigger>
              <ContextMenu.Items>
                <Button>Created</Button>
                <Button>Updated</Button>
                <Button>Deadline</Button>
                <Button>Priority</Button>
              </ContextMenu.Items>
            </ContextMenu>
          </HStack>
        </HStack>
        <HStack>
          <Text>Order direction</Text>
          <Spacer />
          <HStack spacing={2}>
            <ContextMenu>
              <ContextMenu.Trigger>
                <HStack spacing={3}>
                  <Text color={colors.dark.DEFAULT}>Descending</Text>
                  <Image
                    systemName="chevron.up.chevron.down"
                    modifiers={[opacity(0.6)]}
                    color={colors.dark.DEFAULT}
                    size={14}
                  />
                </HStack>
              </ContextMenu.Trigger>
              <ContextMenu.Items>
                <Button>Descending</Button>
                <Button>Ascending</Button>
              </ContextMenu.Items>
            </ContextMenu>
          </HStack>
        </HStack>
      </VStack>
      <VStack spacing={16} alignment="leading">
        <Text modifiers={[opacity(0.75)]}>Display columns</Text>
        <HStack spacing={12}>
          <Button variant="bordered" color={colors.black}>
            ID
          </Button>
          <Button variant="bordered" color={colors.black}>
            Status
          </Button>
          <Button variant="plain">Assignee</Button>
          <Button variant="plain">Priority</Button>
        </HStack>
      </VStack>
    </BottomSheetModal>
  );
};
