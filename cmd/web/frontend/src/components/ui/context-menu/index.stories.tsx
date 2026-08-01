import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { IconProxy } from "@/components/ui/icon-proxy";
import {
  ContextMenu,
  ContextMenuGroupLabel,
  ContextMenuHelp,
  ContextMenuItem,
} from "./index";

function ContextMenuStory() {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <ContextMenu
      id="story-context-menu"
      isOpen={isOpen}
      label="Story actions"
      onOpenChange={setIsOpen}
      placement="bottom-start"
      trigger={(props) => (
        <button type="button" {...props}>
          Actions
        </button>
      )}
    >
      <ContextMenuGroupLabel>Status</ContextMenuGroupLabel>
      <ContextMenuItem
        icon={<IconProxy name="circle" />}
        onSelect={() => setIsOpen(false)}
      >
        Move to review
      </ContextMenuItem>
      <ContextMenuItem
        icon={<IconProxy name="ban" />}
        variant="danger"
        onSelect={() => setIsOpen(false)}
      >
        Delete project
      </ContextMenuItem>
      <ContextMenuItem disabled icon={<IconProxy name="check" />} title="Current state">
        Already in progress
      </ContextMenuItem>
      <ContextMenuHelp>Only valid transitions are enabled.</ContextMenuHelp>
    </ContextMenu>
  );
}

const meta = {
  title: "UI/ContextMenu",
  component: ContextMenuStory,
} satisfies Meta<typeof ContextMenuStory>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Open: Story = {};
