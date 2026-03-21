// MockEditModal - Modal for creating/editing mocks
import React, { useState, useEffect } from 'react';
import { Modal } from './Modal';
import { getMockContent } from '~/services';
import type { MockDefinition, HttpMethod } from '~/types';
import { ALL_HTTP_METHODS } from '~/types';

interface MockEditModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: MockFormData) => void;
  mock?: MockDefinition | null;
  isLoading?: boolean;
}

export interface MockFormData {
  host: string;
  endpoint: string;
  method: HttpMethod;
  statusCode: number;
  responseBody: string;
}

interface FormErrors {
  host?: string;
  endpoint?: string;
  responseBody?: string;
}

interface QueryParam {
  key: string;
  value: string;
}

interface QueryParamError {
  key?: string;
  value?: string;
}

function parseEndpoint(raw: string): { path: string; params: QueryParam[] } {
  const qIndex = raw.indexOf('?');
  if (qIndex === -1) return { path: raw, params: [] };
  const path = raw.slice(0, qIndex);
  const params: QueryParam[] = [];
  new URLSearchParams(raw.slice(qIndex + 1)).forEach((value, key) => {
    params.push({ key, value });
  });
  return { path, params };
}

function buildEndpoint(path: string, params: QueryParam[]): string {
  const filled = params.filter((p) => p.key.trim() && p.value.trim());
  if (filled.length === 0) return path;
  const qs = filled
    .map((p) => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`)
    .join('&');
  return `${path}?${qs}`;
}

export function MockEditModal({ isOpen, onClose, onSave, mock, isLoading }: MockEditModalProps) {
  const [formData, setFormData] = useState<MockFormData>({
    host: '',
    endpoint: '',
    method: 'GET',
    statusCode: 200,
    responseBody: '',
  });
  const [queryParams, setQueryParams] = useState<QueryParam[]>([]);
  const [errors, setErrors] = useState<FormErrors>({});
  const [queryParamErrors, setQueryParamErrors] = useState<QueryParamError[]>([]);
  const [loadingContent, setLoadingContent] = useState(false);

  const isEditMode = !!mock;

  // Reset form when modal opens/closes or mock changes
  useEffect(() => {
    if (isOpen && mock) {
      setErrors({});
      setQueryParamErrors([]);
      const { path, params } = parseEndpoint(mock.endpoint);
      setFormData({
        host: mock.host,
        endpoint: path,
        method: mock.method as HttpMethod,
        statusCode: mock.statusCode,
        responseBody: '',
      });
      setQueryParams(params);

      setLoadingContent(true);
      getMockContent(mock.id)
        .then((body) => {
          setFormData((prev) => ({ ...prev, responseBody: body }));
        })
        .catch((err) => {
          console.error('Failed to load mock content:', err);
        })
        .finally(() => {
          setLoadingContent(false);
        });
    } else if (isOpen) {
      setErrors({});
      setQueryParamErrors([]);
      setFormData({ host: '', endpoint: '', method: 'GET', statusCode: 200, responseBody: '' });
      setQueryParams([]);
    }
  }, [isOpen, mock]);

  const validateQueryParams = (): QueryParamError[] => {
    return queryParams.map((p) => {
      const keyBlank = !p.key.trim();
      const valueBlank = !p.value.trim();
      if (keyBlank && valueBlank) return {};
      if (keyBlank) return { key: 'Key is required' };
      if (valueBlank) return { value: 'Value is required' };
      return {};
    });
  };

  const validate = (): FormErrors => {
    const newErrors: FormErrors = {};
    if (!formData.host.trim()) {
      newErrors.host = 'Host is required';
    }
    if (!formData.endpoint.trim()) {
      newErrors.endpoint = 'Endpoint is required';
    } else if (!formData.endpoint.startsWith('/')) {
      newErrors.endpoint = 'Endpoint must start with /';
    }
    if (!formData.responseBody.trim()) {
      newErrors.responseBody = 'Response body is required';
    }
    return newErrors;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newErrors = validate();
    const newParamErrors = validateQueryParams();
    const hasParamErrors = newParamErrors.some((e) => e.key || e.value);
    if (Object.keys(newErrors).length > 0 || hasParamErrors) {
      setErrors(newErrors);
      setQueryParamErrors(newParamErrors);
      return;
    }
    onSave({
      ...formData,
      endpoint: buildEndpoint(formData.endpoint, queryParams),
    });
  };

  const handleChange = (field: keyof MockFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (field in errors) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  };

  const handleParamChange = (index: number, field: 'key' | 'value', value: string) => {
    setQueryParams((prev) => prev.map((p, i) => (i === index ? { ...p, [field]: value } : p)));
    setQueryParamErrors((prev) =>
      prev.map((e, i) => {
        if (i !== index) return e;
        const updated = { ...queryParams[index], [field]: value };
        if (!updated.key.trim() && !updated.value.trim()) return {};
        return { ...e, [field]: undefined };
      })
    );
  };

  const addParam = () => {
    setQueryParams((prev) => [...prev, { key: '', value: '' }]);
    setQueryParamErrors((prev) => [...prev, {}]);
  };

  const removeParam = (index: number) => {
    setQueryParams((prev) => prev.filter((_, i) => i !== index));
    setQueryParamErrors((prev) => prev.filter((_, i) => i !== index));
  };

  const hasErrors =
    Object.values(errors).some(Boolean) ||
    queryParamErrors.some((e) => e.key || e.value);
  const isSaveDisabled = isLoading || loadingContent || hasErrors;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditMode ? 'Edit Mock' : 'New Mock'}
      width="600px"
    >
      <form onSubmit={handleSubmit}>
        {/* Host */}
        <div className="form-group">
          <label htmlFor="host">Host</label>
          <input
            id="host"
            type="text"
            className={`form-input${errors.host ? ' form-input-error' : ''}`}
            placeholder="example.com"
            value={formData.host}
            onChange={(e) => handleChange('host', e.target.value)}
          />
          {errors.host && <p className="form-error">{errors.host}</p>}
        </div>

        {/* Method + Endpoint */}
        <div className="form-group">
          <label>Request</label>
          <div className="input-group">
            <select
              className="form-select"
              value={formData.method}
              onChange={(e) => handleChange('method', e.target.value)}
            >
              {ALL_HTTP_METHODS.map((method) => (
                <option key={method} value={method}>{method}</option>
              ))}
            </select>
            <input
              type="text"
              className={`form-input${errors.endpoint ? ' form-input-error' : ''}`}
              placeholder="/api/users"
              value={formData.endpoint}
              onChange={(e) => handleChange('endpoint', e.target.value)}
              style={{ flex: 1 }}
            />
          </div>
          {errors.endpoint && <p className="form-error">{errors.endpoint}</p>}
        </div>

        {/* Query Parameters */}
        <div className="form-group">
          <label>Query Parameters</label>
          {queryParams.length > 0 && (
            <div className="query-params-list">
              {queryParams.map((param, index) => {
                const paramError = queryParamErrors[index] ?? {};
                return (
                  <div key={index}>
                    <div className="query-param-row">
                      <input
                        type="text"
                        className={`form-input${paramError.key ? ' form-input-error' : ''}`}
                        placeholder="parameter"
                        value={param.key}
                        onChange={(e) => handleParamChange(index, 'key', e.target.value)}
                      />
                      <span className="query-param-equals">=</span>
                      <input
                        type="text"
                        className={`form-input${paramError.value ? ' form-input-error' : ''}`}
                        placeholder="value"
                        value={param.value}
                        onChange={(e) => handleParamChange(index, 'value', e.target.value)}
                      />
                      <button
                        type="button"
                        className="query-param-remove"
                        onClick={() => removeParam(index)}
                        aria-label="Remove parameter"
                      >
                        ×
                      </button>
                    </div>
                    {(paramError.key || paramError.value) && (
                      <p className="form-error">{paramError.key ?? paramError.value}</p>
                    )}
                  </div>
                );
              })}
            </div>
          )}
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={addParam}
            style={{ marginTop: queryParams.length > 0 ? '8px' : '0' }}
          >
            + Add Parameter
          </button>
        </div>

        {/* Response Body */}
        <div className="form-group">
          <label htmlFor="responseBody">
            Response Body
            {loadingContent && <span style={{ marginLeft: '8px', color: 'var(--text-secondary)' }}>(loading...)</span>}
          </label>
          <textarea
            id="responseBody"
            className={`form-textarea${errors.responseBody ? ' form-input-error' : ''}`}
            placeholder={loadingContent ? 'Loading...' : '{"message": "Hello, World!"}'}
            value={formData.responseBody}
            onChange={(e) => handleChange('responseBody', e.target.value)}
            disabled={loadingContent}
            rows={10}
          />
          {errors.responseBody && <p className="form-error">{errors.responseBody}</p>}
        </div>

        {/* Actions */}
        <div className="form-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={isSaveDisabled}>
            {isLoading ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
