import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ArrowUpDown,
  Ban,
  ChevronDown,
  Check,
  Circle,
  Clipboard,
  Copy,
  Ellipsis,
  Eye,
  Filter,
  Folder,
  History,
  LayoutGrid,
  LayoutDashboard,
  MessageSquare,
  Pause,
  Pencil,
  Play,
  Plus,
  Settings,
  SquareKanban,
  X,
  type LucideIcon,
} from "lucide-react";

const icons = {
  "arrow-down": ArrowDown,
  "arrow-right": ArrowRight,
  "arrow-up": ArrowUp,
  "arrow-up-down": ArrowUpDown,
  ban: Ban,
  "chevron-down": ChevronDown,
  check: Check,
  circle: Circle,
  clipboard: Clipboard,
  copy: Copy,
  ellipsis: Ellipsis,
  eye: Eye,
  filter: Filter,
  folder: Folder,
  history: History,
  "layout-grid": LayoutGrid,
  "layout-dashboard": LayoutDashboard,
  "message-square": MessageSquare,
  pause: Pause,
  pencil: Pencil,
  play: Play,
  plus: Plus,
  settings: Settings,
  "square-kanban": SquareKanban,
  x: X,
} satisfies Record<string, LucideIcon>;

export type IconProxyName = keyof typeof icons;

type IconProxyProps = {
  name: IconProxyName;
  className?: string;
  size?: number;
  strokeWidth?: number;
};

export function IconProxy({
  name,
  className,
  size = 16,
  strokeWidth = 2,
}: IconProxyProps) {
  const Icon = icons[name];

  return (
    <Icon
      aria-hidden="true"
      className={className}
      focusable="false"
      size={size}
      strokeWidth={strokeWidth}
    />
  );
}
