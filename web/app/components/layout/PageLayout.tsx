// PageLayout - Shared layout wrapper for all pages
import type { ReactNode } from 'react';
import { Header } from './Header';

interface PageLayoutProps {
  title: string;
  subtitle?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  pageAccent?: string;
}

export function PageLayout({ title, subtitle, actions, children, pageAccent }: PageLayoutProps) {
  return (
    <>
      <Header title={title} subtitle={subtitle} actions={actions} pageAccent={pageAccent} />
      <div className="content" style={{ flex: 1, overflow: 'hidden', padding: '24px' }}>
        {children}
      </div>
    </>
  );
}

