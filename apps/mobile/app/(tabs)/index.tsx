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
import { Link, Stack } from "expo-router";
import { useState } from "react";

export default function Settings() {
  const [appearanceMode, setAppearanceMode] = useState(0);

  return (
    <Host style={{ flex: 1 }}>
      <Stack.Screen options={{ title: "Home", headerShown: true }} />
      <Form>
        {/* Account & Settings Section */}
        <Section title="Account & Settings">
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="person.circle"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Account Details</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="arrow.triangle.2.circlepath"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Switch Workspace</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <HStack spacing={12}>
            <Image
              systemName="paintbrush"
              color="white"
              size={20}
              modifiers={[
                frame({ width: 32, height: 32 }),
                background("#333333"),
                clipShape("roundedRectangle"),
              ]}
            />
            <Text color="black">Appearance</Text>
            <Spacer />
            <Picker
              options={["Day mode", "Night mode", "System"]}
              selectedIndex={appearanceMode}
              onOptionSelected={({ nativeEvent: { index } }) =>
                setAppearanceMode(index)
              }
              variant="menu"
            />
          </HStack>

          <Button onPress={() => console.log("Log out")}>
            <HStack spacing={12}>
              <Image
                systemName="rectangle.portrait.and.arrow.right"
                color="white"
                size={20}
                modifiers={[
                  frame({ width: 32, height: 32 }),
                  background("#333333"),
                  clipShape("roundedRectangle"),
                ]}
              />
              <Text color="black">Log Out</Text>
            </HStack>
          </Button>

          <ContextMenu>
            <ContextMenu.Items>
              <Button
                systemImage="trash"
                onPress={() => console.log("Delete Workspace")}
              >
                Delete Workspace
              </Button>
            </ContextMenu.Items>
            <ContextMenu.Trigger>
              <Button>
                <HStack spacing={12}>
                  <Image
                    systemName="gear"
                    color="#E5E5E7"
                    size={20}
                    modifiers={[
                      frame({ width: 32, height: 32 }),
                      background("#333333"),
                      clipShape("roundedRectangle"),
                    ]}
                  />
                  <Text color="black">Manage Account</Text>
                  <Spacer />
                  <Image
                    systemName="chevron.right"
                    size={14}
                    color="secondary"
                  />
                </HStack>
              </Button>
            </ContextMenu.Trigger>
          </ContextMenu>
        </Section>

        {/* Support & Info Section */}
        <Section title="Support & Info">
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="questionmark.circle"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Support</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <Button onPress={() => console.log("Send Feedback")}>
            <HStack spacing={12}>
              <Image
                systemName="envelope"
                color="white"
                size={20}
                modifiers={[
                  frame({ width: 32, height: 32 }),
                  background("#333333"),
                  clipShape("roundedRectangle"),
                ]}
              />
              <Text color="black">Send Feedback</Text>
            </HStack>
          </Button>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="book"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Help Center</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="hand.raised"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Privacy Policy</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="bird"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Follow on Twitter</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>

          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="star"
                  color="#E5E5E7"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#333333"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text color="black">Rate the App</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>
        </Section>
      </Form>
    </Host>
  );
}
