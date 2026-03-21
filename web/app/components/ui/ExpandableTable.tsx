// ExpandableTable - Reusable table with expandable rows
import React from 'react';

export interface Column<T> {
  /** Column header text */
  header: string;
  /** Key to access data or render function */
  accessor: keyof T | ((item: T) => React.ReactNode);
  /** Optional className for the cell */
  className?: string;
  /** Optional inline styles */
  style?: React.CSSProperties;
}

interface ExpandableTableProps<T> {
  /** Array of column definitions */
  columns: Column<T>[];
  /** Data to display */
  data: T[];
  /** Function to get unique key for each row */
  getRowKey: (item: T) => string;
  /** Currently selected/expanded item key (null = none) */
  selectedKey: string | null;
  /** Callback when a row is clicked */
  onRowClick: (item: T) => void;
  /** Render function for expanded content */
  renderExpandedContent: (item: T) => React.ReactNode;
  /** Message to show when no data (simple string) */
  emptyMessage?: string;
  /** Rich empty state content to show when no data (overrides emptyMessage) */
  emptyContent?: React.ReactNode;
  /** Optional extra className for data rows */
  rowClassName?: (item: T) => string | undefined;
}

export function ExpandableTable<T>({
  columns,
  data,
  getRowKey,
  selectedKey,
  onRowClick,
  renderExpandedContent,
  emptyMessage = 'No data found',
  emptyContent,
  rowClassName,
}: ExpandableTableProps<T>) {
  const getCellContent = (item: T, column: Column<T>): React.ReactNode => {
    if (typeof column.accessor === 'function') {
      return column.accessor(item);
    }
    return item[column.accessor] as React.ReactNode;
  };

  return (
    <table className="table">
      <thead>
        <tr>
          {columns.map((col, idx) => (
            <th key={idx}>{col.header}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((item) => {
          const key = getRowKey(item);
          const isExpanded = selectedKey === key;

          const extraClass = rowClassName ? (rowClassName(item) ?? '') : '';
          return (
            <React.Fragment key={key}>
              <tr
                className={`expandable-row ${isExpanded ? 'expanded' : ''} ${extraClass}`}
                onClick={() => onRowClick(item)}
              >
                {columns.map((col, idx) => (
                  <td key={idx} className={col.className} style={col.style}>
                    {getCellContent(item, col)}
                  </td>
                ))}
              </tr>
              {isExpanded && (
                <tr className="expanded-content">
                  <td colSpan={columns.length}>
                    <div className="expanded-content-inner">
                      {renderExpandedContent(item)}
                    </div>
                  </td>
                </tr>
              )}
            </React.Fragment>
          );
        })}
        {data.length === 0 && (
          <tr>
            <td colSpan={columns.length} style={{ padding: 0, border: 'none' }}>
              {emptyContent ?? (
                <div className="empty-state">
                  <p className="empty-state-title">{emptyMessage}</p>
                </div>
              )}
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

