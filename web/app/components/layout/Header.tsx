// Header component with page title, optional status, action buttons, and theme toggle
import type { ReactNode, CSSProperties } from 'react';
import { ThemeToggle } from './ThemeToggle';

interface HeaderProps {
  /** Page title */
  title: string;
  /** Optional subtitle or status indicator (e.g., "Live" indicator on logs) */
  subtitle?: ReactNode;
  /** Optional action buttons slot */
  actions?: ReactNode;
  /** CSS color value for the page-level left-border accent */
  pageAccent?: string;
}

export function Header({ title, subtitle, actions, pageAccent }: HeaderProps) {
  return (
    <header className="header" style={pageAccent ? { '--page-accent': pageAccent } as CSSProperties : undefined}>
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-semibold" style={{ color: 'var(--color-text-primary)' }}>
          {title}
        </h1>
        {subtitle && (
          <div className="flex items-center gap-2">
            {subtitle}
          </div>
        )}
      </div>
      <div className="flex items-center gap-4">
        {actions && (
          <div className="flex items-center gap-2">
            {actions}
          </div>
        )}
        <ThemeToggle />
      </div>
    </header>
  );
}

