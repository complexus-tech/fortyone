import React, { useState } from "react";
import { Avatar, Row, Text } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { ProfileSheet } from "./profile-sheet";
import { Pressable } from "react-native";

export const Header = () => {
  const { data: user } = useProfile();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <>
      <Row align="end" justify="between" className="mb-4 " asContainer>
        <Text fontSize="3xl" fontWeight="semibold">
          Home
        </Text>
        <Pressable
          onPress={() => {
            setIsOpened(true);
          }}
          style={{
            zIndex: 1,
          }}
        >
          <Avatar
            name={user?.fullName || user?.username}
            className="size-[36px]"
            color="primary"
            src={user?.avatarUrl}
          />
        </Pressable>
      </Row>
      <ProfileSheet isOpened={isOpened} setIsOpened={setIsOpened} />
    </>
  );
};
