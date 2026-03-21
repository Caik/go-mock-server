// HostEditModal - Modal for creating/editing host configurations
import React, { useState, useEffect } from 'react';
import { Modal } from './Modal';
import type { HostConfig } from '~/types';
import type { HostSaveData } from '~/services/host.service';

// Backend URI regex: ^/?(?:[\w-]+/)*[\w-]+/?(?:\?...)?$
const URI_REGEX = /^\/?(?:[\w-]+\/)*[\w-]+\/?(?:\?(?:[\w-]+=[\w-]+)(?:&[\w-]+=[\w-]+)*)?$/;

// =============================================================================
// State types
// =============================================================================
interface LatencyState { min: string; max: string; p95: string; p99: string; }
interface ErrorRow { code: string; percentage: string; }
interface UriState { pattern: string; latency: LatencyState; errorRows: ErrorRow[]; }

// =============================================================================
// Error types
// =============================================================================
interface LatencyErrors { min?: string; max?: string; p95?: string; p99?: string; range?: string; }
interface ErrorRowError { code?: string; percentage?: string; }
interface UriFormError {
  pattern?: string;
  latency?: LatencyErrors;
  errorRows?: ErrorRowError[];
  atLeastOne?: string;
  errorSum?: string;
}
interface FormErrors {
  hostname?: string;
  latency?: LatencyErrors;
  globalErrorRows?: ErrorRowError[];
  globalErrorSum?: string;
  atLeastOne?: string;
  uris?: UriFormError[];
}

// =============================================================================
// Helpers
// =============================================================================
function emptyLatency(): LatencyState { return { min: '', max: '', p95: '', p99: '' }; }
function emptyUri(): UriState { return { pattern: '', latency: emptyLatency(), errorRows: [] }; }

function hasLatencyAnyField(l: LatencyState): boolean {
  return l.min.trim() !== '' || l.max.trim() !== '' || l.p95.trim() !== '' || l.p99.trim() !== '';
}

function validateLatency(l: LatencyState): LatencyErrors {
  if (!hasLatencyAnyField(l)) return {};

  const hasMin = l.min.trim() !== '';
  const hasMax = l.max.trim() !== '';
  const hasP95 = l.p95.trim() !== '';
  const hasP99 = l.p99.trim() !== '';
  const errs: LatencyErrors = {};

  // Apply "both or neither" rule specifically to min/max
  if (hasMin && !hasMax) errs.max = 'Max is required';
  if (!hasMin && hasMax) errs.min = 'Min is required';

  // Only validate ranges and percentile fields when both min and max are present
  if (hasMin && hasMax) {
    const min = parseInt(l.min, 10);
    const max = parseInt(l.max, 10);

    if (isNaN(min)) errs.min = 'Must be a whole number';
    if (isNaN(max)) errs.max = 'Must be a whole number';

    if (!errs.min && !errs.max && min > max) {
      errs.range = 'Min must be ≤ Max';
    }

    if (hasP95) {
      const p95 = parseInt(l.p95, 10);
      if (isNaN(p95)) {
        errs.p95 = 'Must be a whole number';
      } else if (!errs.range) {
        if (p95 < min || p95 > max) errs.p95 = `P95 must be between ${min} and ${max}`;
      }
    }

    if (hasP99) {
      const p99 = parseInt(l.p99, 10);
      if (isNaN(p99)) {
        errs.p99 = 'Must be a whole number';
      } else if (!hasP95) {
        errs.p99 = 'P95 must be set before P99';
      } else if (!errs.p95 && !errs.range) {
        const p95 = parseInt(l.p95, 10);
        if (p99 < min || p99 < p95 || p99 > max) {
          errs.p99 = `P99 must be between P95 (${p95}) and Max (${max})`;
        }
      }
    }
  }

  return errs;
}

function validateErrorRows(rows: ErrorRow[]): { rowErrors: ErrorRowError[]; sum?: string } {
  const rowErrors = rows.map((row): ErrorRowError => {
    const codeBlank = !row.code.trim();
    const pctBlank = !row.percentage.trim();
    if (codeBlank && pctBlank) return {};
    if (codeBlank) return { code: 'Status code is required' };
    if (pctBlank) return { percentage: 'Percentage is required' };
    const code = parseInt(row.code, 10);
    if (isNaN(code) || code < 400 || code > 599) return { code: 'Must be 400–599' };
    const pct = parseInt(row.percentage, 10);
    if (isNaN(pct) || pct <= 0 || pct > 100) return { percentage: 'Must be 1–100' };
    return {};
  });

  if (rowErrors.some((e) => e.code || e.percentage)) return { rowErrors };

  const total = rows
    .filter((r) => r.code.trim() && r.percentage.trim())
    .reduce((sum, r) => sum + parseInt(r.percentage, 10), 0);

  return { rowErrors, sum: total > 100 ? `Total is ${total}% — must not exceed 100%` : undefined };
}

function buildLatencyPayload(l: LatencyState) {
  if (!l.min.trim() || !l.max.trim()) return undefined;
  return {
    min: parseInt(l.min, 10),
    max: parseInt(l.max, 10),
    ...(l.p95.trim() && { p95: parseInt(l.p95, 10) }),
    ...(l.p99.trim() && { p99: parseInt(l.p99, 10) }),
  };
}

function buildErrorsPayload(rows: ErrorRow[]): Record<string, { percentage: number }> | undefined {
  const filled = rows.filter((r) => r.code.trim() && r.percentage.trim());
  if (filled.length === 0) return undefined;
  const result: Record<string, { percentage: number }> = {};
  for (const row of filled) result[row.code.trim()] = { percentage: parseInt(row.percentage, 10) };
  return result;
}

function hasFormErrors(errors: FormErrors): boolean {
  if (errors.hostname || errors.globalErrorSum || errors.atLeastOne) return true;
  if (errors.latency && Object.values(errors.latency).some(Boolean)) return true;
  if (errors.globalErrorRows?.some((e) => e.code || e.percentage)) return true;
  if (errors.uris?.some((u) => {
    if (u.pattern || u.atLeastOne || u.errorSum) return true;
    if (u.latency && Object.values(u.latency).some(Boolean)) return true;
    if (u.errorRows?.some((e) => e.code || e.percentage)) return true;
    return false;
  })) return true;
  return false;
}

// =============================================================================
// Sub-components
// =============================================================================
function LatencyFields({ value, errors, onChange }: {
  value: LatencyState;
  errors?: LatencyErrors;
  onChange: (field: keyof LatencyState, value: string) => void;
}) {
  return (
    <>
      <div className="latency-grid">
        <div>
          <input type="number" className={`form-input${errors?.min || errors?.range ? ' form-input-error' : ''}`}
            placeholder="Min ms" value={value.min} min={0}
            onChange={(e) => onChange('min', e.target.value)} />
          {errors?.min && <p className="form-error">{errors.min}</p>}
        </div>
        <div>
          <input type="number" className={`form-input${errors?.max || errors?.range ? ' form-input-error' : ''}`}
            placeholder="Max ms" value={value.max} min={0}
            onChange={(e) => onChange('max', e.target.value)} />
          {errors?.max && <p className="form-error">{errors.max}</p>}
        </div>
        <div>
          <input type="number" className={`form-input${errors?.p95 ? ' form-input-error' : ''}`}
            placeholder="P95 ms (optional)" value={value.p95} min={0}
            onChange={(e) => onChange('p95', e.target.value)} />
          {errors?.p95 && <p className="form-error">{errors.p95}</p>}
        </div>
        <div>
          <input type="number" className={`form-input${errors?.p99 ? ' form-input-error' : ''}`}
            placeholder="P99 ms (optional)" value={value.p99} min={0}
            onChange={(e) => onChange('p99', e.target.value)} />
          {errors?.p99 && <p className="form-error">{errors.p99}</p>}
        </div>
      </div>
      {errors?.range && <p className="form-error">{errors.range}</p>}
    </>
  );
}

function ErrorRowsSection({ rows, rowErrors, onRowChange, onAdd, onRemove }: {
  rows: ErrorRow[];
  rowErrors?: ErrorRowError[];
  onRowChange: (index: number, field: 'code' | 'percentage', value: string) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
}) {
  return (
    <>
      {rows.length > 0 && (
        <div className="query-params-list">
          <div style={{ display: 'flex', gap: '8px', marginBottom: '4px', paddingLeft: '2px' }}>
            <span style={{ width: '80px', fontSize: '11px', color: 'var(--text-secondary)', flexShrink: 0 }}>HTTP Status</span>
            <span style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>Rate</span>
          </div>
          {rows.map((row, index) => {
            const rowErr = rowErrors?.[index] ?? {};
            return (
              <div key={index}>
                <div className="query-param-row">
                  <input type="text"
                    className={`form-input${rowErr.code ? ' form-input-error' : ''}`}
                    placeholder="503" value={row.code} style={{ maxWidth: '80px' }}
                    onChange={(e) => onRowChange(index, 'code', e.target.value)} />
                  <span className="query-param-equals">@</span>
                  <input type="text"
                    className={`form-input${rowErr.percentage ? ' form-input-error' : ''}`}
                    placeholder="5%" value={row.percentage} style={{ maxWidth: '70px' }}
                    onChange={(e) => onRowChange(index, 'percentage', e.target.value)} />
                  <span style={{ color: 'var(--text-secondary)', fontSize: '13px', flexShrink: 0 }}>of requests</span>
                  <button type="button" className="query-param-remove" onClick={() => onRemove(index)} aria-label="Remove">×</button>
                </div>
                {(rowErr.code || rowErr.percentage) && (
                  <p className="form-error">{rowErr.code ?? rowErr.percentage}</p>
                )}
              </div>
            );
          })}
        </div>
      )}
      <button type="button" className="btn btn-secondary btn-sm" onClick={onAdd}
        style={{ marginTop: rows.length > 0 ? '8px' : '0' }}>
        + Add Error
      </button>
    </>
  );
}

// =============================================================================
// Main component
// =============================================================================
interface HostEditModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: HostSaveData) => void;
  host?: HostConfig | null;
  isLoading?: boolean;
}

export function HostEditModal({ isOpen, onClose, onSave, host, isLoading }: HostEditModalProps) {
  const [hostname, setHostname] = useState('');
  const [latency, setLatency] = useState<LatencyState>(emptyLatency());
  const [errorRows, setErrorRows] = useState<ErrorRow[]>([]);
  const [uris, setUris] = useState<UriState[]>([]);
  const [errors, setErrors] = useState<FormErrors>({});

  const isEditMode = !!host;

  useEffect(() => {
    if (!isOpen) return;
    setErrors({});

    if (host) {
      setHostname(host.hostname);
      setLatency(host.latency ? {
        min: String(host.latency.min),
        max: String(host.latency.max),
        p95: host.latency.p95 != null ? String(host.latency.p95) : '',
        p99: host.latency.p99 != null ? String(host.latency.p99) : '',
      } : emptyLatency());
      setErrorRows(host.statuses
        ? Object.entries(host.statuses).map(([code, cfg]) => ({ code, percentage: String(cfg.percentage) }))
        : []);
      setUris(host.uris
        ? Object.entries(host.uris).map(([pattern, uri]) => ({
            pattern,
            latency: uri.latency ? {
              min: String(uri.latency.min),
              max: String(uri.latency.max),
              p95: uri.latency.p95 != null ? String(uri.latency.p95) : '',
              p99: uri.latency.p99 != null ? String(uri.latency.p99) : '',
            } : emptyLatency(),
            errorRows: uri.statuses
              ? Object.entries(uri.statuses).map(([code, cfg]) => ({ code, percentage: String(cfg.percentage) }))
              : [],
          }))
        : []);
    } else {
      setHostname('');
      setLatency(emptyLatency());
      setErrorRows([]);
      setUris([]);
    }
  }, [isOpen, host]);

  // ---------------------------------------------------------------------------
  // Submit
  // ---------------------------------------------------------------------------
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const newErrors: FormErrors = {};

    if (!hostname.trim()) newErrors.hostname = 'Hostname is required';

    // Global latency
    const latencyErrs = validateLatency(latency);
    if (Object.keys(latencyErrs).length > 0) newErrors.latency = latencyErrs;

    // Global errors
    const { rowErrors: globalRowErrs, sum: globalSum } = validateErrorRows(errorRows);
    if (globalRowErrs.some((e) => e.code || e.percentage)) newErrors.globalErrorRows = globalRowErrs;
    if (globalSum) newErrors.globalErrorSum = globalSum;

    // URIs
    const uriErrs: UriFormError[] = uris.map((uri) => {
      const uriErr: UriFormError = {};

      if (!uri.pattern.trim()) {
        uriErr.pattern = 'Pattern is required';
      } else if (!URI_REGEX.test(uri.pattern.trim())) {
        uriErr.pattern = 'Invalid pattern (e.g. /api/v1/users)';
      }

      const uriLatencyErrs = validateLatency(uri.latency);
      if (Object.keys(uriLatencyErrs).length > 0) uriErr.latency = uriLatencyErrs;

      const { rowErrors: uriRowErrs, sum: uriSum } = validateErrorRows(uri.errorRows);
      if (uriRowErrs.some((e) => e.code || e.percentage)) uriErr.errorRows = uriRowErrs;
      if (uriSum) uriErr.errorSum = uriSum;

      // Each URI must have at least latency (min+max) or errors
      // Only show if user hasn't started configuring anything (partial input → field errors guide them)
      const uriHasLatency = uri.latency.min.trim() !== '' && uri.latency.max.trim() !== '';
      const uriHasErrors = uri.errorRows.some((r) => r.code.trim() && r.percentage.trim());
      const uriAttemptingAnything = hasLatencyAnyField(uri.latency) || uri.errorRows.some((r) => r.code.trim() || r.percentage.trim());
      if (!uriHasLatency && !uriHasErrors && !uriAttemptingAnything) {
        uriErr.atLeastOne = 'Each URI must have at least a latency or an error configured';
      }

      return uriErr;
    });

    if (uriErrs.some((e) => Object.keys(e).length > 0)) newErrors.uris = uriErrs;

    // At least one global config (latency, errors, or URIs)
    // Only show if the user hasn't started configuring anything — partial/invalid input is guided by field errors
    const isConfiguringAnything =
      hasLatencyAnyField(latency) ||
      errorRows.some((r) => r.code.trim() || r.percentage.trim()) ||
      uris.length > 0;
    if (!isConfiguringAnything) {
      newErrors.atLeastOne = 'At least one configuration (latency, errors, or URIs) is required';
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    // Build payload
    const payload: HostSaveData = { host: hostname.trim() };
    const latencyPayload = buildLatencyPayload(latency);
    if (latencyPayload) payload.latency = latencyPayload;
    const errorsPayload = buildErrorsPayload(errorRows);
    if (errorsPayload) payload.statuses = errorsPayload;

    if (uris.length > 0) {
      payload.uris = {};
      for (const uri of uris) {
        payload.uris[uri.pattern.trim()] = {
          ...(buildLatencyPayload(uri.latency) && { latency: buildLatencyPayload(uri.latency)! }),
          ...(buildErrorsPayload(uri.errorRows) && { statuses: buildErrorsPayload(uri.errorRows)! }),
        };
      }
    }

    onSave(payload);
  };

  // ---------------------------------------------------------------------------
  // Global latency handlers
  // ---------------------------------------------------------------------------
  const handleLatencyChange = (field: keyof LatencyState, value: string) => {
    setLatency((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => {
      const updated = { ...latency, [field]: value };
      const userHasInput = hasLatencyAnyField(updated);
      const hasValidLatency = updated.min.trim() !== '' && updated.max.trim() !== '';
      if (!prev.latency) {
        if (prev.atLeastOne && userHasInput) return { ...prev, atLeastOne: undefined };
        return prev;
      }
      if (!userHasInput) return { ...prev, latency: undefined };
      if (!updated.min.trim() && !updated.max.trim()) {
        return { ...prev, latency: { ...prev.latency, min: undefined, max: undefined, range: undefined } };
      }
      // Re-run validation and keep only errors still present (clears resolved errors without adding new ones)
      const freshErrors = validateLatency(updated);
      const mergedLatency: LatencyErrors = {};
      for (const key of Object.keys(prev.latency) as (keyof LatencyErrors)[]) {
        if (prev.latency[key] && freshErrors[key]) mergedLatency[key] = freshErrors[key];
      }
      return {
        ...prev,
        latency: Object.keys(mergedLatency).length > 0 ? mergedLatency : undefined,
        ...(hasValidLatency && { atLeastOne: undefined }),
      };
    });
  };

  // ---------------------------------------------------------------------------
  // Global error row handlers
  // ---------------------------------------------------------------------------
  const handleErrorRowChange = (index: number, field: 'code' | 'percentage', value: string) => {
    let updatedRows: ErrorRow[] = [];
    setErrorRows((prev) => {
      updatedRows = prev.map((r, i) => (i === index ? { ...r, [field]: value } : r));
      return updatedRows;
    });
    setErrors((prev) => {
      const newRowErrors = prev.globalErrorRows
        ? prev.globalErrorRows.map((e, i) => {
            if (i !== index) return e;
            const row = updatedRows[i];
            if (row && !row.code.trim() && !row.percentage.trim()) return {};
            return { ...e, [field]: undefined };
          })
        : undefined;
      const total = updatedRows
        .filter((r) => r.code.trim() && r.percentage.trim())
        .reduce((sum, r) => sum + parseInt(r.percentage, 10), 0);
      return {
        ...prev,
        ...(newRowErrors && { globalErrorRows: newRowErrors }),
        ...(prev.globalErrorSum && total <= 100 && { globalErrorSum: undefined }),
      };
    });
  };

  const addErrorRow = () => {
    setErrorRows((prev) => [...prev, { code: '', percentage: '' }]);
    setErrors((prev) => ({
      ...prev,
      atLeastOne: undefined,
      globalErrorRows: [...(prev.globalErrorRows ?? []), {}],
    }));
  };

  const removeErrorRow = (index: number) => {
    let remainingRows: ErrorRow[] = [];
    setErrorRows((prev) => {
      remainingRows = prev.filter((_, i) => i !== index);
      return remainingRows;
    });
    setErrors((prev) => {
      const total = remainingRows
        .filter((r) => r.code.trim() && r.percentage.trim())
        .reduce((sum, r) => sum + parseInt(r.percentage, 10), 0);
      return {
        ...prev,
        globalErrorRows: prev.globalErrorRows?.filter((_, i) => i !== index),
        ...(prev.globalErrorSum && total <= 100 && { globalErrorSum: undefined }),
      };
    });
  };

  // ---------------------------------------------------------------------------
  // URI handlers
  // ---------------------------------------------------------------------------
  const addUri = () => setUris((prev) => [...prev, emptyUri()]);

  const removeUri = (uriIndex: number) => {
    setUris((prev) => prev.filter((_, i) => i !== uriIndex));
    setErrors((prev) => ({
      ...prev,
      uris: prev.uris?.filter((_, i) => i !== uriIndex),
    }));
  };

  const handleUriPatternChange = (uriIndex: number, value: string) => {
    setUris((prev) => prev.map((u, i) => (i === uriIndex ? { ...u, pattern: value } : u)));
    setErrors((prev) => {
      if (!prev.uris?.[uriIndex]) return prev;
      const newUriErrors = prev.uris.map((u, i) => (i === uriIndex ? { ...u, pattern: undefined } : u));
      return { ...prev, uris: newUriErrors };
    });
  };

  const handleUriLatencyChange = (uriIndex: number, field: keyof LatencyState, value: string) => {
    setUris((prev) => prev.map((u, i) =>
      i === uriIndex ? { ...u, latency: { ...u.latency, [field]: value } } : u
    ));
    setErrors((prev) => {
      if (!prev.uris?.[uriIndex]) return prev;
      const updatedLatency = { ...uris[uriIndex].latency, [field]: value };
      const hasValidLatency = updatedLatency.min.trim() !== '' && updatedLatency.max.trim() !== '';
      const newUriErrors = prev.uris.map((u, i) => {
        if (i !== uriIndex) return u;
        if (!hasLatencyAnyField(updatedLatency)) return { ...u, latency: undefined };
        if (!updatedLatency.min.trim() && !updatedLatency.max.trim()) {
          return { ...u, latency: u.latency ? { ...u.latency, min: undefined, max: undefined, range: undefined } : undefined };
        }
        if (!u.latency) {
          // No latency errors yet — just clear atLeastOne if user is filling a field
          return { ...u, ...(u.atLeastOne ? { atLeastOne: undefined } : {}) };
        }
        const freshErrors = validateLatency(updatedLatency);
        const mergedLatency: LatencyErrors = {};
        for (const key of Object.keys(u.latency) as (keyof LatencyErrors)[]) {
          if (u.latency[key] && freshErrors[key]) mergedLatency[key] = freshErrors[key];
        }
        return {
          ...u,
          latency: Object.keys(mergedLatency).length > 0 ? mergedLatency : undefined,
          ...(u.atLeastOne && { atLeastOne: undefined }),
        };
      });
      return { ...prev, uris: newUriErrors, ...(hasValidLatency && { atLeastOne: undefined }) };
    });
  };

  const handleUriErrorRowChange = (uriIndex: number, rowIndex: number, field: 'code' | 'percentage', value: string) => {
    let updatedRow: ErrorRow | null = null;
    setUris((prev) => prev.map((u, i) => {
      if (i !== uriIndex) return u;
      const newRows = u.errorRows.map((r, j) => {
        if (j !== rowIndex) return r;
        const updated = { ...r, [field]: value };
        updatedRow = updated;
        return updated;
      });
      return { ...u, errorRows: newRows };
    }));
    setErrors((prev) => {
      if (!prev.uris?.[uriIndex]) return prev;
      const newUriErrors = prev.uris.map((u, i) => {
        if (i !== uriIndex) return u;
        const newRowErrors = (u.errorRows ?? []).map((e, j) => {
          if (j !== rowIndex) return e;
          if (updatedRow && !updatedRow.code.trim() && !updatedRow.percentage.trim()) return {};
          return { ...e, [field]: undefined };
        });
        const rowNowValid = updatedRow && updatedRow.code.trim() !== '' && updatedRow.percentage.trim() !== '';
        return { ...u, errorRows: newRowErrors, ...(rowNowValid && { atLeastOne: undefined }) };
      });
      const anyUriNowValid = newUriErrors.some((u) => !u.atLeastOne);
      return { ...prev, uris: newUriErrors, ...(anyUriNowValid && { atLeastOne: undefined }) };
    });
  };

  const addUriErrorRow = (uriIndex: number) => {
    setUris((prev) => prev.map((u, i) =>
      i === uriIndex ? { ...u, errorRows: [...u.errorRows, { code: '', percentage: '' }] } : u
    ));
  };

  const removeUriErrorRow = (uriIndex: number, rowIndex: number) => {
    let remainingRows: ErrorRow[] = [];
    setUris((prev) => prev.map((u, i) => {
      if (i !== uriIndex) return u;
      remainingRows = u.errorRows.filter((_, j) => j !== rowIndex);
      return { ...u, errorRows: remainingRows };
    }));
    setErrors((prev) => {
      if (!prev.uris?.[uriIndex]?.errorRows) return prev;
      const total = remainingRows
        .filter((r) => r.code.trim() && r.percentage.trim())
        .reduce((sum, r) => sum + parseInt(r.percentage, 10), 0);
      const newUriErrors = prev.uris!.map((u, i) => {
        if (i !== uriIndex) return u;
        return {
          ...u,
          errorRows: u.errorRows?.filter((_, j) => j !== rowIndex),
          ...(u.errorSum && total <= 100 && { errorSum: undefined }),
        };
      });
      return { ...prev, uris: newUriErrors };
    });
  };

  const isSaveDisabled = isLoading || hasFormErrors(errors);

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditMode ? `Edit ${host!.hostname}` : 'New Host'}
      width="660px"
    >
      <form onSubmit={handleSubmit}>

        {/* Hostname */}
        <div className="form-group">
          <label htmlFor="hostname">Hostname</label>
          <input
            id="hostname" type="text"
            className={`form-input${errors.hostname ? ' form-input-error' : ''}`}
            placeholder="api.example.com"
            value={hostname}
            onChange={(e) => { setHostname(e.target.value); setErrors((p) => ({ ...p, hostname: undefined })); }}
            disabled={isEditMode}
          />
          {errors.hostname && <p className="form-error">{errors.hostname}</p>}
        </div>

        {/* Global Latency */}
        <div className="form-group">
          <label>Global Latency <span className="label-optional">(optional — fill Min + Max to enable)</span></label>
          <LatencyFields value={latency} errors={errors.latency} onChange={handleLatencyChange} />
        </div>

        {/* Global Errors */}
        <div className="form-group">
          <label>Global Errors <span className="label-optional">(optional)</span></label>
          <ErrorRowsSection
            rows={errorRows}
            rowErrors={errors.globalErrorRows}
            onRowChange={handleErrorRowChange}
            onAdd={addErrorRow}
            onRemove={removeErrorRow}
          />
          {errors.globalErrorSum && <p className="form-error" style={{ marginTop: '6px' }}>{errors.globalErrorSum}</p>}
        </div>

        {/* URI Overrides */}
        <div className="form-group">
          <label>URI Overrides <span className="label-optional">(optional — each URI needs latency or errors)</span></label>
          {uris.map((uri, uriIndex) => {
            const uriErr = errors.uris?.[uriIndex] ?? {};
            return (
              <div key={uriIndex} className="host-uri-card">
                <div className="host-uri-card-header">
                  <span>URI {uriIndex + 1}</span>
                  <button type="button" className="query-param-remove" onClick={() => removeUri(uriIndex)} aria-label="Remove URI">×</button>
                </div>

                {/* Pattern */}
                <div className="form-group">
                  <label style={{ fontSize: '13px' }}>Pattern</label>
                  <input type="text"
                    className={`form-input${uriErr.pattern ? ' form-input-error' : ''}`}
                    placeholder="/api/v1/users"
                    value={uri.pattern}
                    onChange={(e) => handleUriPatternChange(uriIndex, e.target.value)}
                  />
                  {uriErr.pattern && <p className="form-error">{uriErr.pattern}</p>}
                </div>

                {/* URI Latency */}
                <div className="form-group">
                  <label style={{ fontSize: '13px' }}>Latency <span className="label-optional">(optional)</span></label>
                  <LatencyFields
                    value={uri.latency}
                    errors={uriErr.latency}
                    onChange={(f, v) => handleUriLatencyChange(uriIndex, f, v)}
                  />
                </div>

                {/* URI Errors */}
                <div className="form-group" style={{ marginBottom: 0 }}>
                  <label style={{ fontSize: '13px' }}>Errors <span className="label-optional">(optional)</span></label>
                  <ErrorRowsSection
                    rows={uri.errorRows}
                    rowErrors={uriErr.errorRows}
                    onRowChange={(rowIndex, field, value) => handleUriErrorRowChange(uriIndex, rowIndex, field, value)}
                    onAdd={() => addUriErrorRow(uriIndex)}
                    onRemove={(rowIndex) => removeUriErrorRow(uriIndex, rowIndex)}
                  />
                  {uriErr.errorSum && <p className="form-error" style={{ marginTop: '6px' }}>{uriErr.errorSum}</p>}
                </div>

                {uriErr.atLeastOne && <p className="form-error" style={{ marginTop: '8px' }}>{uriErr.atLeastOne}</p>}
              </div>
            );
          })}
          <button type="button" className="btn btn-secondary btn-sm" onClick={addUri}
            style={{ marginTop: uris.length > 0 ? '8px' : '0' }}>
            + Add URI
          </button>
        </div>

        {errors.atLeastOne && <p className="form-error" style={{ marginBottom: '12px' }}>{errors.atLeastOne}</p>}

        <div className="form-actions">
          <button type="button" className="btn btn-secondary" onClick={onClose}>Cancel</button>
          <button type="submit" className="btn btn-primary" disabled={isSaveDisabled}>
            {isLoading ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
