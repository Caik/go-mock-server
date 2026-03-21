import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ExpandableTable, type Column } from './ExpandableTable';

interface Item {
  id: string;
  name: string;
  value: number;
}

const COLUMNS: Column<Item>[] = [
  { header: 'Name', accessor: 'name' },
  { header: 'Value', accessor: (item) => <span>{item.value}</span> },
];

const DATA: Item[] = [
  { id: '1', name: 'Alpha', value: 10 },
  { id: '2', name: 'Beta', value: 20 },
];

describe('ExpandableTable', () => {
  it('renders column headers', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />} />
    );
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('Value')).toBeInTheDocument();
  });

  it('renders rows from data using key accessor', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />} />
    );
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('renders rows using function accessor', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />} />
    );
    expect(screen.getByText('10')).toBeInTheDocument();
    expect(screen.getByText('20')).toBeInTheDocument();
  });

  it('shows the default emptyMessage when data is empty', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={[]} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />}
        emptyMessage="No results" />
    );
    expect(screen.getByText('No results')).toBeInTheDocument();
  });

  it('shows emptyContent node when provided and data is empty', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={[]} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />}
        emptyContent={<div>Custom empty state</div>} />
    );
    expect(screen.getByText('Custom empty state')).toBeInTheDocument();
  });

  it('calls onRowClick when a data row is clicked', async () => {
    const onRowClick = vi.fn();
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={onRowClick} renderExpandedContent={() => <div />} />
    );
    await userEvent.click(screen.getByText('Alpha'));
    expect(onRowClick).toHaveBeenCalledWith(DATA[0]);
  });

  it('renders expanded content when selectedKey matches a row', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey="1" onRowClick={() => {}} renderExpandedContent={(item) => <div>Details: {item.name}</div>} />
    );
    expect(screen.getByText('Details: Alpha')).toBeInTheDocument();
  });

  it('does not render expanded content for non-selected rows', () => {
    render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey="1" onRowClick={() => {}} renderExpandedContent={(item) => <div>Details: {item.name}</div>} />
    );
    expect(screen.queryByText('Details: Beta')).not.toBeInTheDocument();
  });

  it('applies rowClassName to rows', () => {
    const { container } = render(
      <ExpandableTable columns={COLUMNS} data={DATA} getRowKey={(i) => i.id}
        selectedKey={null} onRowClick={() => {}} renderExpandedContent={() => <div />}
        rowClassName={(item) => item.id === '1' ? 'highlight' : undefined} />
    );
    const rows = container.querySelectorAll('tbody tr.expandable-row');
    expect(rows[0]).toHaveClass('highlight');
    expect(rows[1]).not.toHaveClass('highlight');
  });
});
