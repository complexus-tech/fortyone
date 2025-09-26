import {
  Button,
  ContextMenu,
  Form,
  Host,
  HStack,
  Image,
  Section,
  Spacer,
  Switch,
  Text,
} from "@expo/ui/swift-ui";
import { background, clipShape, frame } from "@expo/ui/swift-ui/modifiers";
import { Link } from "expo-router";
import { useState } from "react";

export default function Settings() {
  const [isDarkMode, setIsDarkMode] = useState(false);

  return (
    <Host style={{ flex: 1 }}>
      <Form>
        {/* Account Section */}
        <Section title="Account">
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="person.circle"
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#007aff"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Account Details</Text>
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
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#34c759"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Switch Workspace</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>
        </Section>

        {/* Appearance Section */}
        <Section>
          <HStack spacing={12}>
            <Image
              systemName="paintbrush"
              color="white"
              size={20}
              modifiers={[
                frame({ width: 32, height: 32 }),
                background("#ff9500"),
                clipShape("roundedRectangle"),
              ]}
            />
            <Text>Dark Mode</Text>
            <Spacer />
            <Switch value={isDarkMode} onValueChange={setIsDarkMode} />
          </HStack>
        </Section>

        {/* Account Management Section */}
        <Section>
          <Button onPress={() => console.log("Log out")}>
            <HStack spacing={12}>
              <Image
                systemName="rectangle.portrait.and.arrow.right"
                color="white"
                size={20}
                modifiers={[
                  frame({ width: 32, height: 32 }),
                  background("#ff3b30"),
                  clipShape("roundedRectangle"),
                ]}
              />
              <Text>Log Out</Text>
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
                    color="white"
                    size={20}
                    modifiers={[
                      frame({ width: 32, height: 32 }),
                      background("#8e8e93"),
                      clipShape("roundedRectangle"),
                    ]}
                  />
                  <Text>Manage Account</Text>
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

        {/* Support & Feedback Section */}
        <Section>
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="questionmark.circle"
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#007aff"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Support</Text>
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
                  background("#ff3b30"),
                  clipShape("roundedRectangle"),
                ]}
              />
              <Text color="red">Send Feedback</Text>
            </HStack>
          </Button>
        </Section>

        {/* Legal & Social Section */}
        <Section>
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="hand.raised"
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#8e8e93"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Privacy Policy</Text>
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
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#1da1f2"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Follow on Twitter</Text>
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
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#ffcc00"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Rate the App</Text>
                <Spacer />
                <Image systemName="chevron.right" size={14} color="secondary" />
              </HStack>
            </Button>
          </Link>
        </Section>

        {/* Help & Docs Section */}
        <Section>
          <Link href="/inbox" asChild>
            <Button>
              <HStack spacing={12}>
                <Image
                  systemName="book"
                  color="white"
                  size={20}
                  modifiers={[
                    frame({ width: 32, height: 32 }),
                    background("#007aff"),
                    clipShape("roundedRectangle"),
                  ]}
                />
                <Text>Help Center</Text>
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
