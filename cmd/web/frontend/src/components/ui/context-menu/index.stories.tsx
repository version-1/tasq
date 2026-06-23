import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
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
      trigger={(props) => (
        <button type="button" {...props}>
          Actions
        </button>
      )}
    >
      <ContextMenuGroupLabel>Status</ContextMenuGroupLabel>
      <ContextMenuItem onSelect={() => setIsOpen(false)}>Move to review</ContextMenuItem>
      <ContextMenuItem disabled title="Current state">
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
