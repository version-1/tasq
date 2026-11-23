import type { Meta, StoryObj } from "@storybook/react-vite";
import { DocumentDetailView } from "./index";

const meta = {
  title: "Features/Projects/Document Detail",
  component: DocumentDetailView,
  args: { projectKey: "tasq", documentID: "doc-1" },
} satisfies Meta<typeof DocumentDetailView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Loading: Story = {};
export const Unavailable: Story = {
  args: { projectError: "Document was not found." },
};
