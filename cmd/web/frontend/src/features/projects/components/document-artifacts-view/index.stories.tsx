import type { Meta, StoryObj } from "@storybook/react-vite";
import { DocumentArtifactsView } from "./index";

const meta = {
  title: "Features/Projects/DocumentArtifactsView",
  component: DocumentArtifactsView,
  args: { projectID: 1, projectKey: "tasq" },
} satisfies Meta<typeof DocumentArtifactsView>;

export default meta;
type Story = StoryObj<typeof meta>;
export const Default: Story = {};
