// Mocks page - Mock endpoint management
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { PageLayout } from '~/components/layout';
import { FilterChipGroup, MockEditModal, ToastContainer, ConfirmModal, type MockFormData } from '~/components/ui';
import { MockDetail } from '~/components/details';
import { getMocks, createMock, updateMock, deleteMock } from '~/services';
import { useToast } from '~/hooks';
import { getStatusClass } from '~/lib/formatters';

const HASH_SEP = '|';

function encodeHash(groupKey: string, mockId: string | null): string {
  return mockId ? `${groupKey}${HASH_SEP}${mockId}` : groupKey;
}

function decodeHash(hash: string): { groupKey: string; mockId: string | null } {
  const idx = hash.indexOf(HASH_SEP);
  if (idx === -1) return { groupKey: hash, mockId: null };
  return { groupKey: hash.slice(0, idx), mockId: hash.slice(idx + 1) };
}
import type { MockDefinition, HttpMethod } from '~/types';
import { ALL_HTTP_METHODS } from '~/types';

interface MockFilters {
  methods: HttpMethod[];
  host: string;
  endpoint: string;
}

interface MockGroup {
  key: string;
  host: string;
  endpoint: string;
  method: string;
  mocks: MockDefinition[];
}

function buildGroups(mocks: MockDefinition[]): MockGroup[] {
  const map = new Map<string, MockGroup>();
  for (const mock of mocks) {
    const key = `${mock.host}::${mock.endpoint}::${mock.method}`;
    if (!map.has(key)) {
      map.set(key, { key, host: mock.host, endpoint: mock.endpoint, method: mock.method, mocks: [] });
    }
    map.get(key)!.mocks.push(mock);
  }
  return Array.from(map.values());
}

export default function MocksPage() {
  const [mocks, setMocks] = useState<MockDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<MockFilters>({ methods: [], host: '', endpoint: '' });

  // Expanded state persisted in URL hash: groupKey|mockId
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(() => {
    if (typeof window === 'undefined') return new Set();
    const hash = window.location.hash.slice(1);
    if (!hash) return new Set();
    return new Set([decodeHash(hash).groupKey]);
  });
  const [expandedMockId, setExpandedMockId] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    const hash = window.location.hash.slice(1);
    if (!hash) return null;
    return decodeHash(hash).mockId;
  });

  // Modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMock, setEditingMock] = useState<MockDefinition | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Delete confirm state
  const [confirmDeleteMock, setConfirmDeleteMock] = useState<MockDefinition | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { toasts, showToast } = useToast();

  const loadMocks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      setMocks(await getMocks());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load mocks');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadMocks(); }, [loadMocks]);

  const filteredGroups = useMemo(() => {
    const filtered = mocks.filter((mock) => {
      if (filters.methods.length > 0 && !filters.methods.includes(mock.method as HttpMethod)) return false;
      if (filters.host && !mock.host.toLowerCase().includes(filters.host.toLowerCase())) return false;
      if (filters.endpoint && !mock.endpoint.toLowerCase().includes(filters.endpoint.toLowerCase())) return false;
      return true;
    });
    return buildGroups(filtered);
  }, [mocks, filters]);

  const updateHash = useCallback((groupKey: string | null, mockId: string | null) => {
    const hash = groupKey ? encodeHash(groupKey, mockId) : '';
    window.history.replaceState(null, '', hash ? `#${hash}` : window.location.pathname);
  }, []);

  const toggleGroup = useCallback((key: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
        setExpandedMockId(null);
        updateHash(null, null);
      } else {
        next.add(key);
        updateHash(key, null);
      }
      return next;
    });
  }, [updateHash]);

  const toggleMock = useCallback((groupKey: string, id: string) => {
    setExpandedMockId((prev) => {
      const next = prev === id ? null : id;
      updateHash(groupKey, next);
      return next;
    });
  }, [updateHash]);

  const handleNewMock = () => { setEditingMock(null); setIsModalOpen(true); };

  const handleEditMock = (mock: MockDefinition) => { setEditingMock(mock); setIsModalOpen(true); };

  const handleDeleteMock = (mock: MockDefinition) => { setConfirmDeleteMock(mock); };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteMock) return;
    setIsDeleting(true);
    try {
      await deleteMock(confirmDeleteMock);
      showToast('Mock deleted', 'success');
      setConfirmDeleteMock(null);
      setExpandedMockId(null);
      await loadMocks();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to delete mock', 'error');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleModalClose = () => { setIsModalOpen(false); setEditingMock(null); };

  const handleModalSave = async (data: MockFormData) => {
    setIsSaving(true);
    try {
      const mockData = { host: data.host, uri: data.endpoint, method: data.method, statusCode: data.statusCode, body: data.responseBody };
      if (editingMock) {
        await updateMock(editingMock.id, mockData);
      } else {
        await createMock(mockData);
      }
      showToast('Mock saved', 'success');
      handleModalClose();
      await loadMocks();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save mock', 'error');
    } finally {
      setIsSaving(false);
    }
  };

  const totalMocks = filteredGroups.reduce((sum, g) => sum + g.mocks.length, 0);

  return (
    <PageLayout title="Mocks">
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        <div className="filter-bar">
          <FilterChipGroup
            label="Method"
            options={ALL_HTTP_METHODS}
            selected={filters.methods}
            onChange={(methods) => setFilters((prev) => ({ ...prev, methods }))}
          />
          <input
            type="text"
            className="filter-input"
            placeholder="Filter by host..."
            style={{ width: '180px' }}
            value={filters.host}
            onChange={(e) => setFilters((prev) => ({ ...prev, host: e.target.value }))}
          />
          <input
            type="text"
            className="filter-input"
            placeholder="Filter by endpoint..."
            style={{ width: '200px' }}
            value={filters.endpoint}
            onChange={(e) => setFilters((prev) => ({ ...prev, endpoint: e.target.value }))}
          />
          <button className="btn btn-primary" onClick={handleNewMock}>+ New Mock</button>
        </div>

        <div className="table-container">
          {loading && (
            <div style={{ padding: '2rem', display: 'flex', justifyContent: 'center' }}>
              <div className="spinner" />
            </div>
          )}
          {error && (
            <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--danger)' }}>
              Error: {error}
            </div>
          )}
          {!loading && !error && (
            <table className="table">
              <thead>
                <tr>
                  <th>Method</th>
                  <th>Endpoint</th>
                  <th>Host</th>
                  <th>Statuses</th>
                </tr>
              </thead>
              <tbody>
                {filteredGroups.length === 0 && (
                  <tr>
                    <td colSpan={4} className="empty-state">No mocks found</td>
                  </tr>
                )}
                {filteredGroups.map((group) => {
                  const groupExpanded = expandedGroups.has(group.key);
                  return (
                    <React.Fragment key={group.key}>
                      {/* Group row */}
                      <tr
                        className={`expandable-row${groupExpanded ? ' expanded' : ''}`}
                        onClick={() => toggleGroup(group.key)}
                      >
                        <td><span className={`method-badge ${group.method}`}>{group.method}</span></td>
                        <td className="cell-mono">{group.endpoint}</td>
                        <td>{group.host}</td>
                        <td>
                          <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                            {group.mocks.map((m) => (
                              <span key={m.id} className={`status-badge ${getStatusClass(m.statusCode)}`}>
                                {m.statusCode}
                              </span>
                            ))}
                          </div>
                        </td>
                      </tr>

                      {/* Sub-rows (one per status code) */}
                      {groupExpanded && group.mocks.map((mock) => {
                        const mockExpanded = expandedMockId === mock.id;
                        return (
                          <React.Fragment key={mock.id}>
                            <tr
                              className={`expandable-row${mockExpanded ? ' expanded' : ''}`}
                              style={{ background: 'var(--bg-secondary)' }}
                              onClick={(e) => { e.stopPropagation(); toggleMock(group.key, mock.id); }}
                            >
                              <td style={{ paddingLeft: '40px' }}>
                                <span className={`status-badge ${getStatusClass(mock.statusCode)}`}>
                                  {mock.statusCode}
                                </span>
                              </td>
                              <td className="cell-mono cell-secondary">{mock.endpoint}</td>
                              <td className="cell-secondary">{mock.host}</td>
                              <td />
                            </tr>
                            {mockExpanded && (
                              <tr className="expanded-content" style={{ background: 'var(--bg-primary)' }}>
                                <td colSpan={4}>
                                  <div className="expanded-content-inner">
                                    <MockDetail
                                      mock={mock}
                                      onEdit={() => handleEditMock(mock)}
                                      onDelete={() => handleDeleteMock(mock)}
                                    />
                                  </div>
                                </td>
                              </tr>
                            )}
                          </React.Fragment>
                        );
                      })}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <MockEditModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        onSave={handleModalSave}
        mock={editingMock}
        isLoading={isSaving}
      />

      <ConfirmModal
        isOpen={!!confirmDeleteMock}
        title="Delete Mock"
        message={
          confirmDeleteMock
            ? `Delete ${confirmDeleteMock.method} ${confirmDeleteMock.endpoint} (${confirmDeleteMock.statusCode}) on ${confirmDeleteMock.host}?`
            : ''
        }
        confirmLabel={isDeleting ? 'Deleting...' : 'Delete'}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDeleteMock(null)}
        isDanger
      />

      <ToastContainer toasts={toasts} />
    </PageLayout>
  );
}
