import type { Meta, StoryObj } from "@storybook/react-vite";
import { ConversationPage } from "./index";

const meta = {
  title: "Features/Issues/ConversationPage",
  component: ConversationPage,
} satisfies Meta<typeof ConversationPage>;

export default meta;

type Story = StoryObj<typeof meta>;

export const RouteContainer: Story = {};
