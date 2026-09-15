import { View } from "react-native";
import { Skeleton } from "@/components/ui";

export const OverviewSkeleton = () => (
  <View style={{ paddingHorizontal: 20, gap: 10 }}>
    {[0, 1].map((row) => (
      <View key={row} style={{ flexDirection: "row", gap: 10 }}>
        {[0, 1].map((column) => (
          <Skeleton
            key={column}
            style={{ flex: 1, height: 96, borderRadius: 12 }}
          />
        ))}
      </View>
    ))}
  </View>
);
