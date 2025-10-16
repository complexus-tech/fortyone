import React, { useCallback, useRef, useMemo } from "react";
import { Pressable } from "react-native";
import { BottomSheetModal, BottomSheetView } from "@gorhom/bottom-sheet";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";

interface PropertyBottomSheetProps {
  children: React.ReactNode;
  trigger: React.ReactNode;
  snapPoints?: string[];
}

export const PropertyBottomSheet = ({
  children,
  trigger,
  snapPoints = ["25%", "50%"],
}: PropertyBottomSheetProps) => {
  const bottomSheetModalRef = useRef<BottomSheetModal>(null);
  const { colorScheme } = useColorScheme();

  const memoizedSnapPoints = useMemo(() => snapPoints, [snapPoints]);

  const handlePresentModalPress = useCallback(() => {
    bottomSheetModalRef.current?.present();
  }, []);

  return (
    <>
      <Pressable onPress={handlePresentModalPress}>{trigger}</Pressable>

      <BottomSheetModal
        ref={bottomSheetModalRef}
        snapPoints={memoizedSnapPoints}
        index={1}
        enablePanDownToClose
        enableDismissOnClose
        backdropComponent={({ style }) => (
          <Pressable
            style={[
              style,
              {
                backgroundColor:
                  colorScheme === "light"
                    ? "rgba(0, 0, 0, 0.2)"
                    : "rgba(0, 0, 0, 0.3)",
              },
            ]}
            onPress={() => bottomSheetModalRef.current?.dismiss()}
          />
        )}
        backgroundStyle={{
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark[300],
          borderTopLeftRadius: 24,
          borderTopRightRadius: 24,
        }}
        handleIndicatorStyle={{
          backgroundColor:
            colorScheme === "light" ? colors.gray[300] : colors.gray[300],
        }}
      >
        <BottomSheetView className="pb-6">{children}</BottomSheetView>
      </BottomSheetModal>
    </>
  );
};
