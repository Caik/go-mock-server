// MockDetail - Expanded content for mock definitions
import React, { useState, useEffect } from 'react';
import { getMockContent } from '~/services';
import type { MockDefinition } from '~/types';

interface MockDetailProps {
  mock: MockDefinition;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function MockDetail({ mock, onEdit, onDelete }: MockDetailProps) {
  const [body, setBody] = useState<string | null>(null);
  const [bodyLoading, setBodyLoading] = useState(true);
  const [bodyError, setBodyError] = useState(false);

  useEffect(() => {
    setBody(null);
    setBodyLoading(true);
    setBodyError(false);

    getMockContent(mock.id)
      .then((content) => setBody(content))
      .catch(() => setBodyError(true))
      .finally(() => setBodyLoading(false));
  }, [mock.id]);

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

      <div className="expanded-details">
        {/* Mock details */}
        <div className="detail-section">
          <h4>Details</h4>
          <div className="detail-row">
            <span className="detail-label">Host</span>
            <span className="detail-value">{mock.host}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Endpoint</span>
            <span className="detail-value cell-mono">{mock.endpoint}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Method</span>
            <span className="detail-value">
              <span className={`method-badge ${mock.method}`}>{mock.method}</span>
            </span>
          </div>
        </div>

        {/* Response Body */}
        <div className="detail-section">
          <h4>Response Body</h4>
          {bodyLoading && (
            <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>(loading...)</span>
          )}
          {!bodyLoading && bodyError && (
            <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>(unavailable)</span>
          )}
          {!bodyLoading && !bodyError && (
            body
              ? <pre className="code-block">{body}</pre>
              : <span style={{ color: 'var(--text-secondary)', fontSize: '13px', fontStyle: 'italic' }}>No response body</span>
          )}
        </div>
      </div>
    </div>
  );
}
