import React from "react";
import { FlatList, View } from "react-native";
import { Link } from "@/types/link";
import { Row, Text } from "@/components/ui";
import { EmptyState } from "./empty-state";
import { Card } from "./card";

export const List = ({ links }: { links: Link[] }) => {
  return (
    <View>
      {links.length > 0 && (
        <Row asContainer className="mb-1">
          <Text color="muted">Links</Text>
        </Row>
      )}
      <FlatList
        data={links}
        renderItem={({ item }) => <Card link={item} />}
        ListEmptyComponent={<EmptyState />}
        showsVerticalScrollIndicator={false}
        keyExtractor={(item) => item.id}
      />
    </View>
  );
};
