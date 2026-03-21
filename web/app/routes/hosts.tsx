// Hosts page - Host configuration management
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Server } from 'lucide-react';
import { PageLayout } from '~/components/layout';
import { FilterChipGroup, ExpandableTable, ConfirmModal, ToastContainer, HostEditModal, MockEditModal, EmptyState, type Column, type MockFormData } from '~/components/ui';
import { HostDetail } from '~/components/details';
import { getHosts, saveHost, deleteHost, getDefaultMocks, createMock, updateMock, deleteMock, type HostSaveData } from '~/services';
import { useUrlHash, useToast } from '~/hooks';
import { getStatusClass } from '~/lib/formatters';
import type { HostConfig, MockDefinition } from '~/types';

type ConfigFilter = 'latency' | 'errors' | 'uris';
const ALL_CONFIG_FILTERS: ConfigFilter[] = ['latency', 'errors', 'uris'];
const CONFIG_LABELS: Record<ConfigFilter, string> = {
  latency: 'Latency',
  errors: 'Errors',
  uris: 'URIs',
};

function totalErrorPct(errors: Record<string, { percentage: number }>): number {
  return Object.values(errors).reduce((sum, e) => sum + e.percentage, 0);
}

export default function HostsPage() {
  const [hosts, setHosts] = useState<HostConfig[]>([]);
  const [defaultMocks, setDefaultMocks] = useState<MockDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedIdRaw] = useUrlHash('');
  const [searchQuery, setSearchQuery] = useState('');
  const [configFilters, setConfigFilters] = useState<ConfigFilter[]>([]);

  // Host modal state
  const [isHostModalOpen, setIsHostModalOpen] = useState(false);
  const [editingHost, setEditingHost] = useState<HostConfig | null>(null);
  const [isSavingHost, setIsSavingHost] = useState(false);

  // Default mock modal state
  const [isMockModalOpen, setIsMockModalOpen] = useState(false);
  const [editingMock, setEditingMock] = useState<MockDefinition | null>(null);
  const [mockInitialValues, setMockInitialValues] = useState<Partial<MockFormData>>({});
  const [isSavingMock, setIsSavingMock] = useState(false);

  // Delete confirm state
  const [confirmDeleteHost, setConfirmDeleteHost] = useState<HostConfig | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [confirmDeleteMock, setConfirmDeleteMock] = useState<MockDefinition | null>(null);
  const [isDeletingMock, setIsDeletingMock] = useState(false);

  const { toasts, showToast } = useToast();

  const setSelectedId = useCallback((id: string | null) => {
    setSelectedIdRaw(id || '');
  }, [setSelectedIdRaw]);

  const handleRowClick = useCallback((host: HostConfig) => {
    setSelectedId(selectedId === host.hostname ? null : host.hostname);
  }, [selectedId, setSelectedId]);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [loadedHosts, loadedDefaultMocks] = await Promise.all([getHosts(), getDefaultMocks()]);
      setHosts(loadedHosts);
      setDefaultMocks(loadedDefaultMocks);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const filteredHosts = useMemo(() => {
    return hosts.filter((host) => {
      if (searchQuery && !host.hostname.toLowerCase().includes(searchQuery.toLowerCase())) return false;
      if (configFilters.length > 0) {
        for (const filter of configFilters) {
          if (filter === 'latency' && !host.latency) return false;
          if (filter === 'errors' && (!host.statuses || Object.keys(host.statuses).length === 0)) return false;
          if (filter === 'uris' && (!host.uris || Object.keys(host.uris).length === 0)) return false;
        }
      }
      return true;
    });
  }, [hosts, searchQuery, configFilters]);

  const handleNewHost = () => { setEditingHost(null); setIsHostModalOpen(true); };
  const handleEditHost = (host: HostConfig) => { setEditingHost(host); setIsHostModalOpen(true); };
  const handleDeleteHost = (host: HostConfig) => { setConfirmDeleteHost(host); };

  const handleHostModalClose = () => { setIsHostModalOpen(false); setEditingHost(null); };

  const handleHostModalSave = async (data: HostSaveData) => {
    setIsSavingHost(true);
    try {
      await saveHost(data);
      showToast('Host saved', 'success');
      handleHostModalClose();
      await loadData();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save host', 'error');
    } finally {
      setIsSavingHost(false);
    }
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteHost) return;
    setIsDeleting(true);
    try {
      await deleteHost(confirmDeleteHost.hostname);
      showToast('Host deleted', 'success');
      setConfirmDeleteHost(null);
      setSelectedId(null);
      await loadData();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to delete host', 'error');
    } finally {
      setIsDeleting(false);
    }
  };

  // Default mock handlers
  const handleAddDefaultMock = (hostname: string) => {
    setEditingMock(null);
    setMockInitialValues({ host: hostname, endpoint: '/_default' });
    setIsMockModalOpen(true);
  };

  const handleEditDefaultMock = (mock: MockDefinition) => {
    setEditingMock(mock);
    setMockInitialValues({});
    setIsMockModalOpen(true);
  };

  const handleDeleteDefaultMock = (mock: MockDefinition) => {
    setConfirmDeleteMock(mock);
  };

  const handleConfirmDeleteMock = async () => {
    if (!confirmDeleteMock) return;
    setIsDeletingMock(true);
    try {
      await deleteMock(confirmDeleteMock);
      showToast('Default mock deleted', 'success');
      setConfirmDeleteMock(null);
      await loadData();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to delete default mock', 'error');
    } finally {
      setIsDeletingMock(false);
    }
  };

  const handleMockModalClose = () => { setIsMockModalOpen(false); setEditingMock(null); setMockInitialValues({}); };

  const handleMockModalSave = async (data: MockFormData) => {
    setIsSavingMock(true);
    try {
      const mockData = { host: data.host, uri: data.endpoint, method: data.method, statusCode: data.statusCode, body: data.responseBody };
      if (editingMock) {
        await updateMock(editingMock.id, mockData);
      } else {
        await createMock(mockData);
      }
      showToast('Default mock saved', 'success');
      handleMockModalClose();
      await loadData();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save default mock', 'error');
    } finally {
      setIsSavingMock(false);
    }
  };

  const columns: Column<HostConfig>[] = [
    {
      header: 'Hostname',
      accessor: (host) => host.hostname,
      className: 'cell-mono',
    },
    {
      header: 'Latency',
      accessor: (host) =>
        host.latency ? (
          <span className="tag">{host.latency.min}–{host.latency.max}ms</span>
        ) : (
          <span className="cell-secondary">—</span>
        ),
    },
    {
      header: 'Error Rate',
      accessor: (host) =>
        host.statuses && Object.keys(host.statuses).length > 0 ? (
          <span className="tag">{totalErrorPct(host.statuses)}%</span>
        ) : (
          <span className="cell-secondary">—</span>
        ),
    },
    {
      header: 'URI Overrides',
      accessor: (host) =>
        host.uris && Object.keys(host.uris).length > 0 ? (
          <span className="tag">{Object.keys(host.uris).length} URIs</span>
        ) : (
          <span className="cell-secondary">—</span>
        ),
    },
    {
      header: 'Default Mocks',
      accessor: (host) => {
        const hostDefaults = defaultMocks.filter((m) => m.host === host.hostname);
        if (hostDefaults.length === 0) return <span className="cell-secondary">—</span>;
        return (
          <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
            {hostDefaults.map((m) => (
              <span key={m.id} className={`status-badge ${getStatusClass(m.statusCode)}`}>
                {m.method} {m.statusCode}
              </span>
            ))}
          </div>
        );
      },
    },
  ];

  return (
    <PageLayout title="Hosts Configuration" pageAccent="var(--page-accent-hosts)">
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        <div className="filter-bar">
          <input
            type="text"
            className="filter-input"
            placeholder="Search hosts..."
            style={{ width: '200px' }}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <FilterChipGroup
            label="Has"
            options={ALL_CONFIG_FILTERS}
            selected={configFilters}
            onChange={setConfigFilters}
            getLabel={(filter) => CONFIG_LABELS[filter]}
          />
          <button className="btn btn-primary" onClick={handleNewHost}>+ New Host</button>
        </div>

        <div className="table-container">
          {loading && (
            <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--text-secondary)' }}>
              Loading hosts...
            </div>
          )}
          {error && (
            <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--error)' }}>
              Error: {error}
            </div>
          )}
          {!loading && !error && (
            <ExpandableTable
              columns={columns}
              data={filteredHosts}
              getRowKey={(host) => host.hostname}
              selectedKey={selectedId || null}
              onRowClick={handleRowClick}
              renderExpandedContent={(host) => (
                <HostDetail
                  host={host}
                  defaultMocks={defaultMocks.filter((m) => m.host === host.hostname)}
                  onEdit={() => handleEditHost(host)}
                  onDelete={() => handleDeleteHost(host)}
                  onAddDefaultMock={() => handleAddDefaultMock(host.hostname)}
                  onEditDefaultMock={handleEditDefaultMock}
                  onDeleteDefaultMock={handleDeleteDefaultMock}
                />
              )}
              emptyContent={
                <EmptyState
                  icon={Server}
                  title={hosts.length === 0 ? 'No hosts configured' : 'No hosts match your search'}
                  description={
                    hosts.length === 0
                      ? 'Add a host to configure latency simulation, error injection, and URI-level overrides.'
                      : 'Try adjusting your search or config filters.'
                  }
                  action={hosts.length === 0 ? (
                    <button className="btn btn-primary btn-sm" onClick={handleNewHost}>+ New Host</button>
                  ) : undefined}
                />
              }
            />
          )}
        </div>
      </div>

      <HostEditModal
        isOpen={isHostModalOpen}
        onClose={handleHostModalClose}
        onSave={handleHostModalSave}
        host={editingHost}
        isLoading={isSavingHost}
      />

      <MockEditModal
        isOpen={isMockModalOpen}
        onClose={handleMockModalClose}
        onSave={handleMockModalSave}
        mock={editingMock}
        initialValues={mockInitialValues}
        readonlyEndpoint
        isLoading={isSavingMock}
      />

      <ConfirmModal
        isOpen={!!confirmDeleteHost}
        title="Delete Host"
        message={confirmDeleteHost ? `Delete configuration for ${confirmDeleteHost.hostname}?` : ''}
        confirmLabel={isDeleting ? 'Deleting...' : 'Delete'}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDeleteHost(null)}
        isDanger
      />

      <ConfirmModal
        isOpen={!!confirmDeleteMock}
        title="Delete Default Mock"
        message={confirmDeleteMock ? `Delete ${confirmDeleteMock.method} default mock (${confirmDeleteMock.statusCode}) for ${confirmDeleteMock.host}?` : ''}
        confirmLabel={isDeletingMock ? 'Deleting...' : 'Delete'}
        onConfirm={handleConfirmDeleteMock}
        onCancel={() => setConfirmDeleteMock(null)}
        isDanger
      />

      <ToastContainer toasts={toasts} />
    </PageLayout>
  );
}
