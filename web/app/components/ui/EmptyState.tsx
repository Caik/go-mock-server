// EmptyState - Reusable empty state component for pages and tables
import type { LucideIcon } from 'lucide-react';

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: React.ReactNode;
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <div className="empty-state-icon">
        <Icon size={24} strokeWidth={1.5} />
      </div>
      <p className="empty-state-title">{title}</p>
      <p className="empty-state-description">{description}</p>
      {action && <div style={{ marginTop: '8px' }}>{action}</div>}
    </div>
  );
}
