// Logs page - Traffic log viewer with SSE streaming
import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Pause, Play, Trash2 } from 'lucide-react';
import { PageLayout } from '~/components/layout';
import { FilterChipGroup, ExpandableTable, type Column } from '~/components/ui';
import { LogDetail } from '~/components/details';
import { subscribeToTraffic } from '~/services';
import { formatTime, getStatusClass } from '~/lib/formatters';
import type { TrafficEntry, HttpMethod, StatusCategory } from '~/types';
import { ALL_HTTP_METHODS, ALL_STATUS_CATEGORIES } from '~/types';

interface Filters {
  methods: HttpMethod[];
  statuses: StatusCategory[];
  host: string;
  path: string;
}

function matchesFilters(entry: TrafficEntry, filters: Filters): boolean {
  // Method filter (empty array = all methods)
  if (filters.methods.length > 0 && !filters.methods.includes(entry.request.method as HttpMethod)) {
    return false;
  }
  // Status filter (empty array = all statuses)
  if (filters.statuses.length > 0) {
    const entryStatusCategory = `${Math.floor(entry.response.status_code / 100)}xx` as StatusCategory;
    if (!filters.statuses.includes(entryStatusCategory)) {
      return false;
    }
  }
  // Host filter
  if (filters.host && !entry.request.host.toLowerCase().includes(filters.host.toLowerCase())) {
    return false;
  }
  // Path filter
  if (filters.path && !entry.request.path.toLowerCase().includes(filters.path.toLowerCase())) {
    return false;
  }
  return true;
}

export default function LogsPage() {
  const [entries, setEntries] = useState<TrafficEntry[]>([]);
  const [bufferedEntries, setBufferedEntries] = useState<TrafficEntry[]>([]);
  const [selectedEntryId, setSelectedEntryId] = useState<string | null>(null);
  const [isPaused, setIsPaused] = useState(false);
  const [filters, setFilters] = useState<Filters>({
    methods: [],
    statuses: [],
    host: '',
    path: '',
  });

  const selectedEntryRef = useRef(selectedEntryId);
  useEffect(() => {
    selectedEntryRef.current = selectedEntryId;
  }, [selectedEntryId]);

  const handleRowClick = useCallback((entry: TrafficEntry) => {
    setSelectedEntryId((prev) => (prev === entry.uuid ? null : entry.uuid));
  }, []);

  const handleLoadBuffered = useCallback(() => {
    setEntries((prev) => [...bufferedEntries, ...prev]);
    setBufferedEntries([]);
    setSelectedEntryId(null);
  }, [bufferedEntries]);

  useEffect(() => {
    const unsubscribe = subscribeToTraffic((newEntry: TrafficEntry) => {
      if (!isPaused) {
        if (selectedEntryRef.current) {
          setBufferedEntries((prev) => [newEntry, ...prev]);
        } else {
          setEntries((prev) => [newEntry, ...prev]);
        }
      }
    });
    return () => unsubscribe();
  }, [isPaused]);

  const handleClear = useCallback(() => {
    setEntries([]);
    setBufferedEntries([]);
    setSelectedEntryId(null);
  }, []);

  const filteredEntries = entries.filter((entry) => matchesFilters(entry, filters));

  const columns: Column<TrafficEntry>[] = [
    {
      header: 'Time',
      accessor: (entry) => formatTime(entry.timestamp),
      className: 'cell-secondary',
    },
    {
      header: 'Method',
      accessor: (entry) => (
        <span className={`method-badge ${entry.request.method}`}>{entry.request.method}</span>
      ),
    },
    { header: 'Host', accessor: (entry) => entry.request.host },
    {
      header: 'Path',
      accessor: (entry) => entry.request.path,
      className: 'cell-secondary cell-mono',
    },
    {
      header: 'Status',
      accessor: (entry) => (
        <span className={`status-badge ${getStatusClass(entry.response.status_code)}`}>
          {entry.response.status_code}
        </span>
      ),
    },
    { header: 'Duration', accessor: (entry) => `${entry.response.latency_ms}ms` },
  ];

  return (
    <PageLayout
      title="Logs"
      subtitle={
        <>
          <span className={`status-dot ${isPaused ? 'paused' : 'live'}`} />
          <span className="cell-secondary" style={{ fontSize: '14px' }}>
            {isPaused ? 'Paused' : 'Live'}
          </span>
          {bufferedEntries.length > 0 && (
            <button onClick={handleLoadBuffered} className="buffered-entries-badge">
              <span className="status-dot buffered" />
              <span>↑ {bufferedEntries.length} new</span>
            </button>
          )}
        </>
      }
      actions={
        <div className="btn-group">
          <button className="btn btn-secondary" onClick={() => setIsPaused(!isPaused)}>
            {isPaused ? <><Play size={14} /> Resume</> : <><Pause size={14} /> Pause</>}
          </button>
          <button className="btn btn-secondary" onClick={handleClear}>
            <Trash2 size={14} /> Clear
          </button>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        <div className="filter-bar">
          <FilterChipGroup
            label="Method"
            options={ALL_HTTP_METHODS}
            selected={filters.methods}
            onChange={(methods) => setFilters((prev) => ({ ...prev, methods }))}
          />
          <FilterChipGroup
            label="Status"
            options={ALL_STATUS_CATEGORIES}
            selected={filters.statuses}
            onChange={(statuses) => setFilters((prev) => ({ ...prev, statuses }))}
          />
          <input
            type="text"
            className="filter-input"
            placeholder="Filter by host..."
            style={{ width: '160px' }}
            value={filters.host}
            onChange={(e) => setFilters((prev) => ({ ...prev, host: e.target.value }))}
          />
          <input
            type="text"
            className="filter-input"
            placeholder="Filter by path..."
            style={{ width: '160px' }}
            value={filters.path}
            onChange={(e) => setFilters((prev) => ({ ...prev, path: e.target.value }))}
          />
        </div>

        <div className="table-container">
          <ExpandableTable
            columns={columns}
            data={filteredEntries}
            getRowKey={(entry) => entry.uuid}
            selectedKey={selectedEntryId}
            onRowClick={handleRowClick}
            renderExpandedContent={(entry) => <LogDetail entry={entry} />}
            emptyMessage="No log entries found"
          />
        </div>
      </div>
    </PageLayout>
  );
}

