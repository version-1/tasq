import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  ContextMenu,
  ContextMenuGroupLabel,
  ContextMenuHelp,
  ContextMenuItem,
  ContextMenuSeparator,
} from "./index";
import { IconProxy } from "@/components/ui/icon-proxy";

function ContextMenuStory() {
  const [isOpen, setIsOpen] = useState(true);

  return (
    <ContextMenu
      id="story-context-menu"
      isOpen={isOpen}
      label="Story actions"
      onOpenChange={setIsOpen}
      placement="bottom-start"
      size="wide"
      trigger={(props) => (
        <button type="button" {...props}>
          Actions
        </button>
      )}
    >
      <ContextMenuItem
        icon={<IconProxy name="clipboard" size={18} />}
        onSelect={() => setIsOpen(false)}
      >
        Copy thread ID
      </ContextMenuItem>
      <ContextMenuSeparator />
      <ContextMenuGroupLabel>Change status</ContextMenuGroupLabel>
      <ContextMenuItem
        accessory={<IconProxy name="check" size={18} />}
        disabled
        icon={<IconProxy name="play" size={18} />}
        selected
      >
        Ready (current)
      </ContextMenuItem>
      <ContextMenuItem
        icon={<IconProxy name="eye" size={18} />}
        onSelect={() => setIsOpen(false)}
      >
        Move to review
      </ContextMenuItem>
      <ContextMenuItem variant="danger" onSelect={() => setIsOpen(false)}>
        Delete project
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
