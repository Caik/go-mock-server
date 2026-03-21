// Mocks page - Mock endpoint management
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { PageLayout } from '~/components/layout';
import { FilterChipGroup, ExpandableTable, MockEditModal, ToastContainer, ConfirmModal, type Column, type MockFormData } from '~/components/ui';
import { MockDetail } from '~/components/details';
import { getMocks, createMock, updateMock, deleteMock } from '~/services';
import { useUrlHash, useToast } from '~/hooks';
import type { MockDefinition, HttpMethod } from '~/types';
import { ALL_HTTP_METHODS } from '~/types';

interface MockFilters {
  methods: HttpMethod[];
  host: string;
  endpoint: string;
}

export default function MocksPage() {
  const [mocks, setMocks] = useState<MockDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedIdRaw] = useUrlHash('');
  const [filters, setFilters] = useState<MockFilters>({
    methods: [],
    host: '',
    endpoint: '',
  });

  // Modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMock, setEditingMock] = useState<MockDefinition | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // Delete confirm state
  const [confirmDeleteMock, setConfirmDeleteMock] = useState<MockDefinition | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { toasts, showToast } = useToast();

  const setSelectedId = useCallback((id: string | null) => {
    setSelectedIdRaw(id || '');
  }, [setSelectedIdRaw]);

  const handleRowClick = useCallback((mock: MockDefinition) => {
    setSelectedId(selectedId === mock.id ? null : mock.id);
  }, [selectedId, setSelectedId]);

  const loadMocks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const loadedMocks = await getMocks();
      setMocks(loadedMocks);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load mocks');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadMocks();
  }, [loadMocks]);

  const filteredMocks = useMemo(() => {
    return mocks.filter((mock) => {
      if (filters.methods.length > 0 && !filters.methods.includes(mock.method as HttpMethod)) {
        return false;
      }

      if (filters.host && !mock.host.toLowerCase().includes(filters.host.toLowerCase())) {
        return false;
      }

      if (filters.endpoint && !mock.endpoint.toLowerCase().includes(filters.endpoint.toLowerCase())) {
        return false;
      }

      return true;
    });
  }, [mocks, filters]);

  const handleNewMock = () => {
    setEditingMock(null);
    setIsModalOpen(true);
  };

  const handleEditMock = (mock: MockDefinition) => {
    setEditingMock(mock);
    setIsModalOpen(true);
  };

  const handleDeleteMock = (mock: MockDefinition) => {
    setConfirmDeleteMock(mock);
  };

  const handleConfirmDelete = async () => {
    if (!confirmDeleteMock) return;
    setIsDeleting(true);
    try {
      await deleteMock(confirmDeleteMock);
      showToast('Mock deleted', 'success');
      setConfirmDeleteMock(null);
      setSelectedId(null);
      await loadMocks();
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to delete mock', 'error');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setEditingMock(null);
  };

  const handleModalSave = async (data: MockFormData) => {
    setIsSaving(true);
    try {
      const mockData = {
        host: data.host,
        uri: data.endpoint,
        method: data.method,
        statusCode: data.statusCode,
        body: data.responseBody,
      };

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

  const columns: Column<MockDefinition>[] = [
    {
      header: 'Method',
      accessor: (mock) => (
        <span className={`method-badge ${mock.method}`}>{mock.method}</span>
      ),
    },
    {
      header: 'Status',
      accessor: (mock) => (
        <span className="badge">{mock.statusCode}</span>
      ),
    },
    {
      header: 'Endpoint',
      accessor: (mock) => mock.endpoint,
      className: 'cell-mono',
    },
    { header: 'Host', accessor: (mock) => mock.host },
  ];

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
          <button className="btn btn-primary" onClick={handleNewMock}>
            + New Mock
          </button>
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
            <ExpandableTable
              columns={columns}
              data={filteredMocks}
              getRowKey={(mock) => mock.id}
              selectedKey={selectedId || null}
              onRowClick={handleRowClick}
              renderExpandedContent={(mock) => (
                <MockDetail
                  mock={mock}
                  onEdit={() => handleEditMock(mock)}
                  onDelete={() => handleDeleteMock(mock)}
                />
              )}
              emptyMessage="No mocks found"
            />
          )}
        </div>
      </div>

      {/* Edit/Create Modal */}
      <MockEditModal
        isOpen={isModalOpen}
        onClose={handleModalClose}
        onSave={handleModalSave}
        mock={editingMock}
        isLoading={isSaving}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!confirmDeleteMock}
        title="Delete Mock"
        message={
          confirmDeleteMock
            ? `Delete ${confirmDeleteMock.method} ${confirmDeleteMock.endpoint} on ${confirmDeleteMock.host}?`
            : ''
        }
        confirmLabel={isDeleting ? 'Deleting...' : 'Delete'}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDeleteMock(null)}
        isDanger
      />

      {/* Toast notifications */}
      <ToastContainer toasts={toasts} />
    </PageLayout>
  );
}
