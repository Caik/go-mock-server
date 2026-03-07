// LogDetail - Expanded content for traffic log entries
import React, { useState } from 'react';
import type { TrafficEntry } from '~/types';
import { getStatusClass, formatHeaders } from '~/lib/formatters';

type DetailTab = 'request' | 'response';

interface LogDetailProps {
  entry: TrafficEntry;
}

export function LogDetail({ entry }: LogDetailProps) {
  const [activeTab, setActiveTab] = useState<DetailTab>('request');

  const fullPath = `${entry.request.host}${entry.request.path}${entry.request.query ?? ''}`;

  return (
    <>
      {/* Tabs */}
      <div className="tabs" style={{ marginBottom: '16px' }}>
        <div
          className={`tab ${activeTab === 'request' ? 'active' : ''}`}
          onClick={(e) => { e.stopPropagation(); setActiveTab('request'); }}
        >
          Request
        </div>
        <div
          className={`tab ${activeTab === 'response' ? 'active' : ''}`}
          onClick={(e) => { e.stopPropagation(); setActiveTab('response'); }}
        >
          Response
        </div>
      </div>

      <div className="expanded-details">
        {activeTab === 'request' && (
          <>
            <div className="detail-section">
              <h4>Request Info</h4>
              <div className="detail-row">
                <span className="detail-label">Method</span>
                <span className="detail-value">
                  <span className={`method-badge ${entry.request.method}`}>
                    {entry.request.method}
                  </span>
                </span>
              </div>
              <div className="detail-row">
                <span className="detail-label">URL</span>
                <span className="detail-value cell-mono">{fullPath}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Host</span>
                <span className="detail-value">{entry.request.host}</span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Path</span>
                <span className="detail-value cell-mono">{entry.request.path}</span>
              </div>
              {entry.request.query && (
                <div className="detail-row">
                  <span className="detail-label">Query</span>
                  <span className="detail-value cell-mono">{entry.request.query}</span>
                </div>
              )}
            </div>

            {entry.request.headers && Object.keys(entry.request.headers).length > 0 && (
              <div className="detail-section">
                <h4>Request Headers</h4>
                <pre className="code-block">{formatHeaders(entry.request.headers)}</pre>
              </div>
            )}
          </>
        )}

        {activeTab === 'response' && (
          <>
            <div className="detail-section">
              <h4>Response Info</h4>
              <div className="detail-row">
                <span className="detail-label">Status</span>
                <span className="detail-value">
                  <span className={`status-badge ${getStatusClass(entry.response.status_code)}`}>
                    {entry.response.status_code}
                  </span>
                </span>
              </div>
              <div className="detail-row">
                <span className="detail-label">Duration</span>
                <span className="detail-value">{entry.response.latency_ms}ms</span>
              </div>
              {entry.response.content_type && (
                <div className="detail-row">
                  <span className="detail-label">Content-Type</span>
                  <span className="detail-value cell-mono">{entry.response.content_type}</span>
                </div>
              )}
              <div className="detail-row">
                <span className="detail-label">Body Size</span>
                <span className="detail-value">{entry.response.body_size} bytes</span>
              </div>
            </div>

            {entry.metadata && Object.keys(entry.metadata).length > 0 && (
              <div className="detail-section">
                <h4>Mock Info</h4>
                {Object.entries(entry.metadata).map(([key, value]) => (
                  <div className="detail-row" key={key}>
                    <span className="detail-label">{key}</span>
                    <span className="detail-value">
                      {key === 'Matched' ? (
                        value === 'true'
                          ? <span className="match yes">✓ Yes</span>
                          : <span className="match no">✗ No</span>
                      ) : (
                        <span className="cell-mono">{value}</span>
                      )}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </>
  );
}
