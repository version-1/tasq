import type { Meta, StoryObj } from "@storybook/react-vite";
import { storyRuns } from "@/stories/fixtures";
import { RunsSection } from "./index";

const meta = {
  title: "Features/Issues/IssueDetail/RunsSection",
  component: RunsSection,
  args: {
    issueID: 1,
    error: "",
    isLoading: false,
    runs: storyRuns,
  },
} satisfies Meta<typeof RunsSection>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Populated: Story = {};

export const Empty: Story = {
  args: {
    runs: [],
  },
};
