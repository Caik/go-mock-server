import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { FilterChipGroup } from './FilterChipGroup';

const OPTIONS = ['GET', 'POST', 'PUT'] as const;
type Option = typeof OPTIONS[number];

describe('FilterChipGroup', () => {
  it('renders the label', () => {
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={[]} onChange={() => {}} />);
    expect(screen.getByText('Method:')).toBeInTheDocument();
  });

  it('renders all options as buttons', () => {
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={[]} onChange={() => {}} />);
    expect(screen.getByText('GET')).toBeInTheDocument();
    expect(screen.getByText('POST')).toBeInTheDocument();
    expect(screen.getByText('PUT')).toBeInTheDocument();
  });

  it('marks selected chips as active', () => {
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={['GET']} onChange={() => {}} />);
    expect(screen.getByText('GET')).toHaveClass('active');
    expect(screen.getByText('POST')).not.toHaveClass('active');
  });

  it('calls onChange with added option when an inactive chip is clicked', async () => {
    const onChange = vi.fn();
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={['GET']} onChange={onChange} />);
    await userEvent.click(screen.getByText('POST'));
    expect(onChange).toHaveBeenCalledWith(['GET', 'POST']);
  });

  it('calls onChange without option when an active chip is clicked (deselect)', async () => {
    const onChange = vi.fn();
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={['GET']} onChange={onChange} />);
    await userEvent.click(screen.getByText('GET'));
    expect(onChange).toHaveBeenCalledWith([]);
  });

  it('uses getLabel to render display text', () => {
    const getLabel = (o: Option) => o.toLowerCase();
    render(<FilterChipGroup label="Method" options={OPTIONS} selected={[]} onChange={() => {}} getLabel={getLabel} />);
    expect(screen.getByText('get')).toBeInTheDocument();
    expect(screen.getByText('post')).toBeInTheDocument();
  });
});
