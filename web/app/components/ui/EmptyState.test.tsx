import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { EmptyState } from './EmptyState';

const MockIcon = () => <svg data-testid="mock-icon" />;

describe('EmptyState', () => {
  it('renders the title', () => {
    render(<EmptyState icon={MockIcon as any} title="Nothing here" description="Add something to get started." />);
    expect(screen.getByText('Nothing here')).toBeInTheDocument();
  });

  it('renders the description', () => {
    render(<EmptyState icon={MockIcon as any} title="Title" description="Helpful description text." />);
    expect(screen.getByText('Helpful description text.')).toBeInTheDocument();
  });

  it('renders the icon', () => {
    render(<EmptyState icon={MockIcon as any} title="Title" description="Desc" />);
    expect(screen.getByTestId('mock-icon')).toBeInTheDocument();
  });

  it('does not render action slot when not provided', () => {
    render(<EmptyState icon={MockIcon as any} title="Title" description="Desc" />);
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('renders action slot when provided', () => {
    render(
      <EmptyState
        icon={MockIcon as any}
        title="Title"
        description="Desc"
        action={<button>Add Item</button>}
      />
    );
    expect(screen.getByRole('button', { name: 'Add Item' })).toBeInTheDocument();
  });
});
