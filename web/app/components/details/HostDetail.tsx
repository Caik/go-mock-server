// HostDetail - Expanded content for host configurations
import React from 'react';
import type { HostConfig, MockDefinition } from '~/types';
import { getStatusClass } from '~/lib/formatters';

interface HostDetailProps {
  host: HostConfig;
  defaultMocks: MockDefinition[];
  onEdit?: () => void;
  onDelete?: () => void;
  onAddDefaultMock?: () => void;
  onEditDefaultMock?: (mock: MockDefinition) => void;
  onDeleteDefaultMock?: (mock: MockDefinition) => void;
}

export function HostDetail({ host, defaultMocks, onEdit, onDelete, onAddDefaultMock, onEditDefaultMock, onDeleteDefaultMock }: HostDetailProps) {
  const statusEntries = host.statuses ? Object.entries(host.statuses) : [];
  const uriEntries = host.uris ? Object.entries(host.uris) : [];

  return (
    <div onClick={(e) => e.stopPropagation()}>
      {/* Action buttons */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '16px' }}>
        <div className="btn-group">
          <button className="btn btn-secondary btn-sm" onClick={onEdit}>Edit</button>
          <button className="btn btn-danger btn-sm" onClick={onDelete}>Delete</button>
        </div>
      </div>

      {/* 2-column responsive layout */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '24px' }}>

        {/* Left column: host config */}
        <div>
          {!host.latency && statusEntries.length === 0 && uriEntries.length === 0 && (
            <p className="cell-secondary">No configuration set for this host.</p>
          )}

          {host.latency && (
            <div className="detail-section">
              <h4>Global Latency</h4>
              <div className="detail-row">
                <span className="detail-label">Min</span>
                <span className="detail-value">{host.latency.min}ms</span>
              </div>
              {host.latency.p95 != null && (
                <div className="detail-row">
                  <span className="detail-label">P95</span>
                  <span className="detail-value">{host.latency.p95}ms</span>
                </div>
              )}
              {host.latency.p99 != null && (
                <div className="detail-row">
                  <span className="detail-label">P99</span>
                  <span className="detail-value">{host.latency.p99}ms</span>
                </div>
              )}
              <div className="detail-row">
                <span className="detail-label">Max</span>
                <span className="detail-value">{host.latency.max}ms</span>
              </div>
            </div>
          )}

          {statusEntries.length > 0 && (
            <div className="detail-section">
              <h4>Status Simulation</h4>
              {statusEntries.map(([code, cfg]) => (
                <div key={code} className="detail-row">
                  <span className="detail-label">
                    <span className={`status-badge ${getStatusClass(parseInt(code, 10))}`}>{code}</span>
                  </span>
                  <span className="detail-value">{cfg.percentage}%</span>
                </div>
              ))}
            </div>
          )}

          {uriEntries.length > 0 && (
            <div className="detail-section">
              <h4>URI Overrides ({uriEntries.length})</h4>
              <div className="uri-list">
                {uriEntries.map(([pattern, uri]) => (
                  <div key={pattern} className="uri-item">
                    <div className="uri-path">{pattern}</div>
                    <div className="uri-config">
                      {uri.latency && (
                        <span>Latency: {uri.latency.min}–{uri.latency.max}ms</span>
                      )}
                      {uri.latency && uri.statuses && Object.keys(uri.statuses).length > 0 && ' · '}
                      {uri.statuses && Object.entries(uri.statuses).map(([code, cfg]) => (
                        <span key={code}>
                          <span className={`status-badge ${getStatusClass(parseInt(code, 10))}`}>{code}</span>
                          {' '}@ {cfg.percentage}%
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right column: default mocks */}
        <div>
          <div className="detail-section">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
              <h4 style={{ marginBottom: 0 }}>Default Mocks</h4>
              <button className="btn btn-secondary btn-sm" onClick={onAddDefaultMock}>+ Add</button>
            </div>

            {defaultMocks.length === 0 ? (
              <p className="cell-secondary" style={{ fontSize: '13px', marginTop: '8px' }}>
                No default mocks configured.
              </p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', marginTop: '4px' }}>
                {defaultMocks.map((mock) => (
                  <div
                    key={mock.id}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '6px 10px',
                      background: 'var(--bg-tertiary)',
                      border: '1px solid var(--border)',
                      borderRadius: '6px',
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <span className={`method-badge ${mock.method}`}>{mock.method}</span>
                      <span className={`status-badge ${getStatusClass(mock.statusCode)}`}>{mock.statusCode}</span>
                    </div>
                    <div className="btn-group">
                      <button className="btn btn-secondary btn-sm" onClick={() => onEditDefaultMock?.(mock)}>Edit</button>
                      <button className="btn btn-danger btn-sm" onClick={() => onDeleteDefaultMock?.(mock)}>Delete</button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}
