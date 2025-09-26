import {
  Button,
  ContextMenu,
  Form,
  Host,
  HStack,
  Image,
  Picker,
  Section,
  Spacer,
  Switch,
  Text,
} from "@expo/ui/swift-ui";
import { background, clipShape, frame } from "@expo/ui/swift-ui/modifiers";
import { Link } from "expo-router";
import { useState } from "react";

export default function Index() {
  const [isAirplaneMode, setIsAirplaneMode] = useState(true);
  const [selectedIndex, setSelectedIndex] = useState(0);
  return (
    <Host style={{ flex: 1 }}>
      <Form>
        <Section>
          <HStack spacing={8}>
            <Image
              systemName="airplane"
              color="white"
              size={18}
              modifiers={[
                frame({ width: 28, height: 28 }),
                background("#ffa500"),
                clipShape("roundedRectangle"),
              ]}
            />
            <Text>Airplane Mode</Text>
            <Text>Air</Text>
            <Spacer />
            <Switch value={isAirplaneMode} onValueChange={setIsAirplaneMode} />
          </HStack>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={8}>
                <Image
                  systemName="wifi"
                  color="white"
                  size={18}
                  modifiers={[
                    frame({ width: 28, height: 28 }),
                    background("#007aff"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="primary">Wi-Fi</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <ContextMenu>
            <ContextMenu.Items>
              <Button
                systemImage="person.crop.circle.badge.xmark"
                onPress={() => console.log("Pressed1")}
              >
                Hello
              </Button>
              <Button
                variant="bordered"
                systemImage="heart"
                onPress={() => console.log("Pressed2")}
              >
                Love it
              </Button>
              <Picker
                label="Doggos"
                options={["very", "veery", "veeery", "much"]}
                variant="menu"
                selectedIndex={selectedIndex}
                onOptionSelected={({ nativeEvent: { index } }) =>
                  setSelectedIndex(index)
                }
              />
            </ContextMenu.Items>
            <ContextMenu.Trigger>
              <Button variant="glass">Show Menu</Button>
            </ContextMenu.Trigger>
          </ContextMenu>
        </Section>
      </Form>
    </Host>
  );
}
