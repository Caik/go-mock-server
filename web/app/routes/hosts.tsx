// Hosts page - Host configuration management
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { PageLayout } from '~/components/layout';
import { FilterChipGroup, ExpandableTable, ConfirmModal, ToastContainer, HostEditModal, type Column } from '~/components/ui';
import { HostDetail } from '~/components/details';
import { getHosts, saveHost, deleteHost, type HostSaveData } from '~/services';
import { useUrlHash, useToast } from '~/hooks';
import type { HostConfig } from '~/types';

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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedIdRaw] = useUrlHash('');
  const [searchQuery, setSearchQuery] = useState('');
  const [configFilters, setConfigFilters] = useState<ConfigFilter[]>([]);

  // Modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingHost, setEditingHost] = useState<HostConfig | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Delete confirm state
  const [confirmDeleteHost, setConfirmDeleteHost] = useState<HostConfig | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { toasts, showToast } = useToast();

  const setSelectedId = useCallback((id: string | null) => {
    setSelectedIdRaw(id || '');
  }, [setSelectedIdRaw]);

  const handleRowClick = useCallback((host: HostConfig) => {
    setSelectedId(selectedId === host.hostname ? null : host.hostname);
  }, [selectedId, setSelectedId]);

  const loadHosts = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const loaded = await getHosts();
      setHosts(loaded);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load hosts');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadHosts();
  }, [loadHosts]);

  const filteredHosts = useMemo(() => {
    return hosts.filter((host) => {
      if (searchQuery && !host.hostname.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }
      if (configFilters.length > 0) {
        for (const filter of configFilters) {
          if (filter === 'latency' && !host.latency) return false;
          if (filter === 'errors' && (!host.errors || Object.keys(host.errors).length === 0)) return false;
          if (filter === 'uris' && (!host.uris || Object.keys(host.uris).length === 0)) return false;
        }
      }
      return true;
    });
  }, [hosts, searchQuery, configFilters]);

  const handleNewHost = () => {
    setEditingHost(null);
    setIsModalOpen(true);
  };

  const handleEditHost = (host: HostConfig) => {
    setEditingHost(host);
    setIsModalOpen(true);
  };

  const handleDeleteHost = (host: HostConfig) => {
    setConfirmDeleteHost(host);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setEditingHost(null);
  };

  const handleModalSave = async (data: HostSaveData) => {
    setIsSaving(true);
    try {
      await saveHost(data);
      showToast('Host saved', 'success');
      handleModalClose();
      await loadHosts();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save host', 'error');
    } finally {
      setIsSaving(false);
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
      await loadHosts();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to delete host', 'error');
    } finally {
      setIsDeleting(false);
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
        host.errors && Object.keys(host.errors).length > 0 ? (
          <span className="tag">{totalErrorPct(host.errors)}%</span>
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
  ];

  return (
    <PageLayout title="Hosts Configuration">
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
                  onEdit={() => handleEditHost(host)}
                  onDelete={() => handleDeleteHost(host)}
                />
              )}
              emptyMessage="No hosts found"
            />
          )}
        </div>
      </div>

      <HostEditModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        onSave={handleModalSave}
        host={editingHost}
        isLoading={isSaving}
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

      <ToastContainer toasts={toasts} />
    </PageLayout>
  );
}
