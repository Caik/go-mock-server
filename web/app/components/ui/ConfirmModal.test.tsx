import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfirmModal } from './ConfirmModal';

describe('ConfirmModal', () => {
  it('renders nothing when isOpen is false', () => {
    render(
      <ConfirmModal isOpen={false} title="Delete?" message="Are you sure?" onConfirm={() => {}} onCancel={() => {}} />
    );
    expect(screen.queryByText('Delete?')).not.toBeInTheDocument();
  });

  it('renders the title and message when open', () => {
    render(
      <ConfirmModal isOpen={true} title="Delete?" message="This cannot be undone." onConfirm={() => {}} onCancel={() => {}} />
    );
    expect(screen.getByText('Delete?')).toBeInTheDocument();
    expect(screen.getByText('This cannot be undone.')).toBeInTheDocument();
  });

  it('shows "Confirm" as the default confirm label', () => {
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" onConfirm={() => {}} onCancel={() => {}} />
    );
    expect(screen.getByRole('button', { name: 'Confirm' })).toBeInTheDocument();
  });

  it('shows a custom confirmLabel', () => {
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" confirmLabel="Delete" onConfirm={() => {}} onCancel={() => {}} />
    );
    expect(screen.getByRole('button', { name: 'Delete' })).toBeInTheDocument();
  });

  it('calls onConfirm when confirm button is clicked', async () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" onConfirm={onConfirm} onCancel={() => {}} />
    );
    await userEvent.click(screen.getByRole('button', { name: 'Confirm' }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it('calls onCancel when Cancel button is clicked', async () => {
    const onCancel = vi.fn();
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" onConfirm={() => {}} onCancel={onCancel} />
    );
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('gives the confirm button btn-danger class when isDanger is true', () => {
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" onConfirm={() => {}} onCancel={() => {}} isDanger />
    );
    expect(screen.getByRole('button', { name: 'Confirm' })).toHaveClass('btn-danger');
  });

  it('gives the confirm button btn-primary class when isDanger is false', () => {
    render(
      <ConfirmModal isOpen={true} title="Title" message="msg" onConfirm={() => {}} onCancel={() => {}} />
    );
    expect(screen.getByRole('button', { name: 'Confirm' })).toHaveClass('btn-primary');
  });
});
