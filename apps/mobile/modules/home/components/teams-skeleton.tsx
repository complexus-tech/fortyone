import { View } from "react-native";
import { Skeleton } from "@/components/ui";

export const TeamsSkeleton = () => (
  <View style={{ paddingHorizontal: 20, paddingTop: 24 }}>
    <Skeleton style={{ width: 62, height: 16, marginBottom: 6 }} />
    {[0, 1, 2].map((row) => (
      <View
        key={row}
        style={{
          minHeight: 56,
          flexDirection: "row",
          alignItems: "center",
          gap: 12,
        }}
      >
        <Skeleton style={{ width: 10, height: 10, borderRadius: 3 }} />
        <Skeleton style={{ width: 100, height: 18 }} />
      </View>
    ))}
  </View>
);
