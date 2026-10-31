import type { Meta, StoryObj } from "@storybook/react-vite";
import { storyIssue } from "@/stories/fixtures";
import { ContextMenu } from "@/components/ui/context-menu";
import { AutoOKControl } from "./index";

const meta = {
  title: "Features/Issues/AutoOKControl",
  component: AutoOKControl,
  decorators: [(Story) => (
    <div style={{ minHeight: 180 }}>
      <ContextMenu
        id="auto-ok-story-menu"
        isOpen
        label="Issue actions"
        onOpenChange={() => {}}
        placement="bottom-start"
        trigger={(props) => <button type="button" {...props}>Actions</button>}
      >
        <Story />
      </ContextMenu>
    </div>
  )],
  args: { issue: storyIssue(), onChange: async () => {} },
} satisfies Meta<typeof AutoOKControl>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Off: Story = {};
export const On: Story = { args: { issue: { ...storyIssue(), autoOk: true, autoOkCount: 2 } } };
export const LimitReached: Story = { args: { issue: { ...storyIssue(), autoOk: false, autoOkCount: 5, autoOkStopReason: "limit" } } };
export const Failed: Story = { args: { issue: { ...storyIssue(), autoOk: false, autoOkStopReason: "error" } } };
export const Readonly: Story = { args: { disabled: true } };
