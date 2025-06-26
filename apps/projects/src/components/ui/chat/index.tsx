"use client";

import { AiIcon, ArrowRightIcon } from "icons";
import { useState, useRef, useEffect } from "react";
import { Avatar, Box, Button, Dialog, Input, Text, Flex } from "ui";
import { cn } from "lib";

type Message = {
  id: string;
  content: string;
  sender: "user" | "ai";
  timestamp: Date;
};

const SUGGESTED_PROMPTS = [
  "Summarize the current sprint progress",
  "What are the highest priority tasks?",
  "Show me blocked stories and their blockers",
  "Generate a status report for the team",
  "What's the team velocity this sprint?",
  "Help me plan the next sprint",
];

export const Chat = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content:
        "Hi! I'm Maya, your AI assistant. I can help you with project insights, sprint planning, story management, and more. How can I assist you today?",
      sender: "ai",
      timestamp: new Date(),
    },
  ]);
  const [inputValue, setInputValue] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = (content: string) => {
    if (!content.trim()) return;

    const newMessage: Message = {
      id: Date.now().toString(),
      content,
      sender: "user",
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, newMessage]);
    setInputValue("");
    setIsLoading(true);

    // Simulate AI response (replace with actual API call)
    setTimeout(() => {
      const aiResponse: Message = {
        id: (Date.now() + 1).toString(),
        content:
          "I understand your request. While I'm currently in development, I'll soon be able to provide detailed insights about your projects, analyze sprint data, and help you make data-driven decisions. Is there anything specific you'd like to know about your current sprint?",
        sender: "ai",
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, aiResponse]);
      setIsLoading(false);
    }, 1000);
  };

  const handleSuggestedPrompt = (prompt: string) => {
    handleSendMessage(prompt);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage(inputValue);
    }
  };

  return (
    <>
      <Box className="fixed bottom-8 right-8">
        <Button
          className="bg-gray-50/70 backdrop-blur dark:bg-dark-200/70"
          color="tertiary"
          leftIcon={<AiIcon />}
          onClick={() => {
            setIsOpen(true);
          }}
          rounded="full"
          size="lg"
        >
          Ask Maya
        </Button>
      </Box>

      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="max-w-[42rem] rounded-[2rem] md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Header className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
            <Dialog.Title className="text-lg">
              <Flex align="center" gap={3}>
                <Avatar
                  className="bg-gradient-to-br from-primary to-secondary"
                  color="primary"
                  size="sm"
                >
                  <AiIcon />
                </Avatar>
                <Text fontSize="lg" fontWeight="medium">
                  Maya AI Assistant
                </Text>
              </Flex>
            </Dialog.Title>
          </Dialog.Header>

          <Dialog.Body className="h-[80dvh] max-h-[80dvh] p-0">
            <Flex className="h-full" direction="column">
              {/* Messages Area */}
              <Box className="flex-1 overflow-y-auto px-6 py-6">
                <Flex direction="column" gap={6}>
                  {messages.map((message) => (
                    <Flex
                      className={cn({
                        "flex-row-reverse": message.sender === "user",
                      })}
                      gap={3}
                      key={message.id}
                    >
                      <Avatar
                        className={cn({
                          "bg-gradient-to-br from-primary to-secondary":
                            message.sender === "ai",
                        })}
                        color={
                          message.sender === "ai" ? "primary" : "secondary"
                        }
                        name={message.sender === "user" ? "You" : "Maya"}
                        size="md"
                      >
                        {message.sender === "ai" && <AiIcon />}
                      </Avatar>
                      <Flex
                        className={cn("max-w-[75%] flex-1", {
                          "items-end": message.sender === "user",
                        })}
                        direction="column"
                      >
                        <Box
                          className={cn("rounded-2xl rounded-tl-none p-4", {
                            "bg-primary text-white": message.sender === "user",
                            "bg-gray-50 dark:bg-dark-50":
                              message.sender === "ai",
                          })}
                        >
                          <Text
                            className={cn({
                              "text-white": message.sender === "user",
                            })}
                            fontSize="md"
                          >
                            {message.content}
                          </Text>
                        </Box>
                        <Text className="mt-2 px-1" color="muted" fontSize="sm">
                          {message.timestamp.toLocaleTimeString([], {
                            hour: "2-digit",
                            minute: "2-digit",
                          })}
                        </Text>
                      </Flex>
                    </Flex>
                  ))}

                  {isLoading ? (
                    <Flex gap={3}>
                      <Avatar
                        className="bg-gradient-to-br from-primary to-secondary"
                        color="primary"
                        size="md"
                      >
                        <AiIcon />
                      </Avatar>
                      <Box className="rounded-2xl bg-gray-50 p-4 dark:bg-dark-100">
                        <Flex gap={1}>
                          <Box className="size-2 animate-bounce rounded-full bg-gray" />
                          <Box
                            className="size-2 animate-bounce rounded-full bg-gray"
                            style={{ animationDelay: "0.1s" }}
                          />
                          <Box
                            className="size-2 animate-bounce rounded-full bg-gray"
                            style={{ animationDelay: "0.2s" }}
                          />
                        </Flex>
                      </Box>
                    </Flex>
                  ) : null}
                  <div ref={messagesEndRef} />
                </Flex>
              </Box>

              {/* Suggested Prompts */}
              {messages.length === 1 && (
                <Box className="border-t border-gray-100 px-6 py-6 dark:border-dark-100">
                  <Text className="mb-4" fontSize="md" fontWeight="medium">
                    Try asking:
                  </Text>
                  <Flex direction="column" gap={3}>
                    {SUGGESTED_PROMPTS.map((prompt, index) => (
                      <Button
                        className="h-auto justify-start p-4 text-left"
                        color="tertiary"
                        key={index}
                        onClick={() => {
                          handleSuggestedPrompt(prompt);
                        }}
                        variant="outline"
                      >
                        <Text fontSize="md">{prompt}</Text>
                      </Button>
                    ))}
                  </Flex>
                </Box>
              )}

              {/* Input Area */}
              <Box className="border-t border-gray-100 p-6 dark:border-dark-100">
                <Flex align="end" gap={3}>
                  <Box className="flex-1">
                    <Input
                      disabled={isLoading}
                      onChange={(e) => {
                        setInputValue(e.target.value);
                      }}
                      onKeyDown={handleKeyDown}
                      placeholder="Ask Maya anything about your projects..."
                      rounded="lg"
                      size="lg"
                      value={inputValue}
                    />
                  </Box>
                  <Button
                    className="h-[3.2rem] w-[3.2rem] p-0"
                    disabled={!inputValue.trim() || isLoading}
                    onClick={() => {
                      handleSendMessage(inputValue);
                    }}
                    rounded="full"
                  >
                    <ArrowRightIcon />
                  </Button>
                </Flex>
                <Text className="mt-3" color="muted" fontSize="sm">
                  Press Enter to send, Shift+Enter for new line
                </Text>
              </Box>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
