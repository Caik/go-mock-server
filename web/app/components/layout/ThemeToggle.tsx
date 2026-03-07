// Pill-shaped theme toggle with Sun and Moon icons
import { Sun, Moon } from 'lucide-react';
import { useTheme } from '~/hooks';

export function ThemeToggle() {
  const { theme, toggleTheme, hasSynced } = useTheme();

  return (
    <div className="theme-switch" onClick={toggleTheme}>
      <span
        className={`theme-switch-option ${hasSynced && theme === 'light' ? 'active' : ''}`}
        aria-label="Light theme"
      >
        <Sun size={16} />
      </span>
      <span
        className={`theme-switch-option ${hasSynced && theme === 'dark' ? 'active' : ''}`}
        aria-label="Dark theme"
      >
        <Moon size={16} />
      </span>
    </div>
  );
}

