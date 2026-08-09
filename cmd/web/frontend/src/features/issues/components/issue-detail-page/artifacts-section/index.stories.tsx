import type { Meta, StoryObj } from "@storybook/react-vite";
import { ArtifactsSection } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/ArtifactsSection",
  component: ArtifactsSection,
  args: {
    artifacts: [
      {
        type: "pull_request",
        data_type: "url",
        data_value: "https://github.com/version-1/tasq/pull/14",
      },
    ],
  },
} satisfies Meta<typeof ArtifactsSection>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  args: {
    artifacts: [],
  },
};
