// HostDetail - Expanded content for host configurations
import React from 'react';
import type { HostConfig } from '~/types';
import { getStatusClass } from '~/lib/formatters';

interface HostDetailProps {
  host: HostConfig;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function HostDetail({ host, onEdit, onDelete }: HostDetailProps) {
  const statusEntries = host.statuses ? Object.entries(host.statuses) : [];
  const uriEntries = host.uris ? Object.entries(host.uris) : [];
  const hasConfig = host.latency || statusEntries.length > 0 || uriEntries.length > 0;

  return (
    <div onClick={(e) => e.stopPropagation()}>
      {/* Action buttons */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '12px' }}>
        <div className="btn-group">
          <button className="btn btn-secondary btn-sm" onClick={onEdit}>
            Edit
          </button>
          <button className="btn btn-danger btn-sm" onClick={onDelete}>
            Delete
          </button>
        </div>
      </div>

      {!hasConfig && (
        <p className="cell-secondary">No configuration set for this host.</p>
      )}

      <div className="expanded-details">
        {/* Global Latency */}
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

        {/* Status Simulation */}
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
      </div>

      {/* URI Overrides */}
      {uriEntries.length > 0 && (
        <div className="detail-section" style={{ marginTop: '16px' }}>
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
  );
}
