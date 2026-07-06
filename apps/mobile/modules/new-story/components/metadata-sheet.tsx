import React from "react";
import { Pressable, ScrollView, View } from "react-native";
import { BottomSheetModal, Row, Text } from "@/components/ui";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";

export type MetadataOption = {
  id: string;
  label: string;
  description?: string;
  color?: string | null;
};

type MetadataSheetProps = {
  isOpen: boolean;
  title: string;
  options: MetadataOption[];
  selectedIds: string[];
  multiple?: boolean;
  emptyText?: string;
  onClose: () => void;
  onSelect: (option: MetadataOption) => void;
};

export const MetadataSheet = ({
  isOpen,
  title,
  options,
  selectedIds,
  multiple,
  emptyText = "Nothing available",
  onClose,
  onSelect,
}: MetadataSheetProps) => {
  const { resolvedTheme } = useTheme();
  const borderColor =
    resolvedTheme === "light" ? colors.gray[100] : colors.dark[100];

  return (
    <BottomSheetModal isOpen={isOpen} onClose={onClose} spacing={14}>
      <Text fontWeight="semibold" fontSize="lg">
        {title}
      </Text>
      <ScrollView style={{ maxHeight: 360 }}>
        {options.length === 0 ? (
          <Text color="muted">{emptyText}</Text>
        ) : (
          options.map((option) => {
            const selected = selectedIds.includes(option.id);
            return (
              <Pressable
                key={option.id}
                onPress={() => onSelect(option)}
                style={{
                  paddingVertical: 13,
                  borderBottomWidth: 1,
                  borderBottomColor: borderColor,
                }}
              >
                <Row align="center" justify="between">
                  <View style={{ flex: 1, paddingRight: 12 }}>
                    <Row align="center" gap={2}>
                      {option.color ? (
                        <View
                          style={{
                            width: 9,
                            height: 9,
                            borderRadius: 5,
                            backgroundColor: option.color,
                          }}
                        />
                      ) : null}
                      <Text>{option.label}</Text>
                    </Row>
                    {option.description ? (
                      <Text color="muted" fontSize="sm" className="mt-1">
                        {option.description}
                      </Text>
                    ) : null}
                  </View>
                  <Text color={selected ? "primary" : "muted"}>
                    {selected ? (multiple ? "Selected" : "Current") : "Select"}
                  </Text>
                </Row>
              </Pressable>
            );
          })
        )}
      </ScrollView>
    </BottomSheetModal>
  );
};
