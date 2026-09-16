import type { ReactNode, ComponentProps } from "react";
import { Col, Row, Text } from "@/components/ui";
import { GlassSurface } from "@/components/ui/glass-surface";
import { PropertyBottomSheet } from "@/modules/story/components/properties/property-bottom-sheet";
import { PropertyChip } from "@/modules/story/components/properties/property-chip";

export function DetailProperties({ children }: { children: ReactNode }) {
  return (
    <Col asContainer align="stretch" className="my-[8px]">
      <GlassSurface cornerRadius={20} style={{ padding: 8 }}>
        <Row
          wrap
          align="center"
          style={{ columnGap: 4, rowGap: 0, minWidth: 0, maxWidth: "100%" }}
        >
          {children}
        </Row>
      </GlassSurface>
    </Col>
  );
}
export function SelectProperty({
  label,
  icon,
  ...props
}: Omit<ComponentProps<typeof PropertyBottomSheet>, "trigger"> & {
  label: string;
  icon?: ReactNode;
}) {
  return (
    <PropertyBottomSheet
      {...props}
      trigger={
        <PropertyChip>
          {icon}
          <Text
            fontSize="sm"
            numberOfLines={1}
            style={{ maxWidth: 150, flexShrink: 1 }}
          >
            {label}
          </Text>
        </PropertyChip>
      }
    />
  );
}
