// Header component with page title, optional status, action buttons, and theme toggle
import type { ReactNode } from 'react';
import { ThemeToggle } from './ThemeToggle';

interface HeaderProps {
  /** Page title */
  title: string;
  /** Optional subtitle or status indicator (e.g., "Live" indicator on logs) */
  subtitle?: ReactNode;
  /** Optional action buttons slot */
  actions?: ReactNode;
}

export function Header({ title, subtitle, actions }: HeaderProps) {
  return (
    <header className="header">
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

