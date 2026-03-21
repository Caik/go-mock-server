import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Modal } from './Modal';

describe('Modal', () => {
  afterEach(() => {
    document.body.style.overflow = '';
  });

  it('renders nothing when isOpen is false', () => {
    render(<Modal isOpen={false} onClose={() => {}} title="Test Modal">content</Modal>);
    expect(screen.queryByText('Test Modal')).not.toBeInTheDocument();
  });

  it('renders title and children when isOpen is true', () => {
    render(<Modal isOpen={true} onClose={() => {}} title="My Modal"><p>Modal body</p></Modal>);
    expect(screen.getByText('My Modal')).toBeInTheDocument();
    expect(screen.getByText('Modal body')).toBeInTheDocument();
  });

  it('calls onClose when the overlay is clicked', async () => {
    const onClose = vi.fn();
    render(<Modal isOpen={true} onClose={onClose} title="Modal">content</Modal>);
    await userEvent.click(screen.getByRole('dialog').parentElement!);
    expect(onClose).toHaveBeenCalled();
  });

  it('does not call onClose when modal content is clicked', async () => {
    const onClose = vi.fn();
    render(<Modal isOpen={true} onClose={onClose} title="Modal"><p>Inner</p></Modal>);
    await userEvent.click(screen.getByText('Inner'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('calls onClose when the close button is clicked', async () => {
    const onClose = vi.fn();
    render(<Modal isOpen={true} onClose={onClose} title="Modal">content</Modal>);
    await userEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose when Escape key is pressed', async () => {
    const onClose = vi.fn();
    render(<Modal isOpen={true} onClose={onClose} title="Modal">content</Modal>);
    await userEvent.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalled();
  });

  it('does not add modal-danger class by default', () => {
    render(<Modal isOpen={true} onClose={() => {}} title="Modal">content</Modal>);
    const modal = screen.getByRole('dialog');
    expect(modal).not.toHaveClass('modal-danger');
  });

  it('adds modal-danger class when isDanger is true', () => {
    render(<Modal isOpen={true} onClose={() => {}} title="Delete?" isDanger>content</Modal>);
    const modal = screen.getByRole('dialog');
    expect(modal).toHaveClass('modal-danger');
  });

  it('sets body overflow to hidden when open', () => {
    render(<Modal isOpen={true} onClose={() => {}} title="Modal">content</Modal>);
    expect(document.body.style.overflow).toBe('hidden');
  });
});
