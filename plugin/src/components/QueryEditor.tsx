import React, { useState, useEffect, useCallback } from 'react';
import {
  InlineField,
  Input,
  Select,
  InlineSwitch,
  Button,
  useStyles2,
  RadioButtonGroup,
  Collapse,
  IconButton,
} from '@grafana/ui';
import { QueryEditorProps, SelectableValue, GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { DataSource } from '../datasource/datasource';
import { SSHPrometheusDataSourceOptions, SSHPrometheusQuery, defaultQuery } from '../types';

type Props = QueryEditorProps<DataSource, SSHPrometheusQuery, SSHPrometheusDataSourceOptions>;

interface LabelFilter {
  label: string;
  operator: string;
  value: string;
}

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    display: flex;
    flex-direction: column;
    gap: 8px;
  `,
  topRow: css`
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  `,
  leftSection: css`
    display: flex;
    align-items: center;
    gap: 8px;
  `,
  builderRow: css`
    display: flex;
    align-items: flex-start;
    gap: 16px;
    flex-wrap: wrap;
  `,
  metricSection: css`
    display: flex;
    flex-direction: column;
    gap: 4px;
  `,
  labelSection: css`
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  `,
  sectionLabel: css`
    font-size: 12px;
    color: ${theme.colors.text.secondary};
    font-weight: 500;
  `,
  labelFilters: css`
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
  `,
  labelFilter: css`
    display: flex;
    align-items: center;
    gap: 4px;
    background: ${theme.colors.background.secondary};
    padding: 4px 8px;
    border-radius: 4px;
  `,
  operationsButton: css`
    margin-top: 8px;
  `,
  optionsRow: css`
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
    padding: 8px 0;
  `,
  optionItem: css`
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: ${theme.colors.text.secondary};
  `,
  optionLabel: css`
    font-weight: 500;
  `,
  codeEditor: css`
    width: 100%;
    font-family: monospace;
    font-size: 14px;
    padding: 8px 12px;
    border: 1px solid ${theme.colors.border.weak};
    border-radius: 4px;
    background: ${theme.colors.background.primary};
    color: ${theme.colors.text.primary};
    min-height: 60px;
    resize: vertical;
    &:focus {
      outline: none;
      border-color: ${theme.colors.primary.main};
    }
  `,
  kickstartButton: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    &:hover {
      background: ${theme.colors.action.hover};
    }
  `,
});

const operatorOptions: Array<SelectableValue<string>> = [
  { label: '=', value: '=' },
  { label: '!=', value: '!=' },
  { label: '=~', value: '=~' },
  { label: '!~', value: '!~' },
];

const formatOptions: Array<SelectableValue<string>> = [
  { label: 'Time series', value: 'time_series' },
  { label: 'Table', value: 'table' },
  { label: 'Heatmap', value: 'heatmap' },
];

const typeOptions: Array<SelectableValue<string>> = [
  { label: 'Both', value: 'both' },
  { label: 'Range', value: 'range' },
  { label: 'Instant', value: 'instant' },
];

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const styles = useStyles2(getStyles);
  const currentQuery = { ...defaultQuery, ...query };

  const [editorMode, setEditorMode] = useState<'builder' | 'code'>('builder');
  const [metrics, setMetrics] = useState<Array<SelectableValue<string>>>([]);
  const [labels, setLabels] = useState<Array<SelectableValue<string>>>([]);
  const [labelValues, setLabelValues] = useState<Record<string, Array<SelectableValue<string>>>>({});
  const [labelFilters, setLabelFilters] = useState<LabelFilter[]>([]);
  const [selectedMetric, setSelectedMetric] = useState<string | null>(currentQuery.expr || null);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [labelsLoading, setLabelsLoading] = useState(false);
  const [optionsOpen, setOptionsOpen] = useState(false);

  // Parse existing expression into metric and label filters
  useEffect(() => {
    if (currentQuery.expr) {
      const match = currentQuery.expr.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)?(\{.*\})?$/);
      if (match) {
        const metric = match[1] || '';
        setSelectedMetric(metric || null);

        if (match[2]) {
          // Parse label filters from {label="value", ...}
          const labelsStr = match[2].slice(1, -1);
          const filters: LabelFilter[] = [];
          const labelRegex = /([a-zA-Z_][a-zA-Z0-9_]*)\s*(=~|!=|!~|=)\s*"([^"]*)"/g;
          let labelMatch;
          while ((labelMatch = labelRegex.exec(labelsStr)) !== null) {
            filters.push({
              label: labelMatch[1],
              operator: labelMatch[2],
              value: labelMatch[3],
            });
          }
          setLabelFilters(filters);
        }
      }
    }
  }, []);

  const loadMetrics = useCallback(async () => {
    if (metrics.length > 0) return;
    setMetricsLoading(true);
    try {
      const result = await datasource.metricFindQuery('metrics()');
      const metricOptions = result.map((r) => ({ label: r.text, value: r.text })).sort((a, b) => a.label.localeCompare(b.label));
      setMetrics([{ label: 'Metrics explorer', value: '' }, ...metricOptions]);
    } catch (error) {
      setMetrics([{ label: 'Metrics explorer', value: '' }]);
    } finally {
      setMetricsLoading(false);
    }
  }, [datasource, metrics.length]);

  const loadLabels = useCallback(async () => {
    if (labels.length > 0) return;
    setLabelsLoading(true);
    try {
      const result = await datasource.metricFindQuery('label_names()');
      const labelOptions = result.map((r) => ({ label: r.text, value: r.text })).sort((a, b) => a.label.localeCompare(b.label));
      setLabels(labelOptions);
    } catch (error) {
      setLabels([]);
    } finally {
      setLabelsLoading(false);
    }
  }, [datasource, labels.length]);

  const loadLabelValues = useCallback(async (label: string) => {
    if (labelValues[label]) return;
    try {
      const result = await datasource.metricFindQuery(`label_values(${label})`);
      const valueOptions = result.map((r) => ({ label: r.text, value: r.text })).sort((a, b) => a.label.localeCompare(b.label));
      setLabelValues((prev) => ({ ...prev, [label]: valueOptions }));
    } catch (error) {
      setLabelValues((prev) => ({ ...prev, [label]: [] }));
    }
  }, [datasource, labelValues]);

  useEffect(() => {
    loadMetrics();
    loadLabels();
  }, [loadMetrics, loadLabels]);

  const buildExpression = useCallback(() => {
    let expr = selectedMetric || '';
    if (labelFilters.length > 0) {
      const filterStr = labelFilters
        .filter((f) => f.label && f.value)
        .map((f) => `${f.label}${f.operator}"${f.value}"`)
        .join(', ');
      if (filterStr) {
        expr += `{${filterStr}}`;
      }
    }
    return expr;
  }, [selectedMetric, labelFilters]);

  useEffect(() => {
    if (editorMode === 'builder') {
      const expr = buildExpression();
      if (expr !== currentQuery.expr) {
        onChange({ ...currentQuery, expr });
      }
    }
  }, [selectedMetric, labelFilters, editorMode, buildExpression]);

  const onMetricChange = (value: SelectableValue<string>) => {
    setSelectedMetric(value.value || null);
  };

  const onLabelFilterChange = (index: number, field: keyof LabelFilter, value: string) => {
    const newFilters = [...labelFilters];
    newFilters[index] = { ...newFilters[index], [field]: value };
    setLabelFilters(newFilters);

    if (field === 'label' && value) {
      loadLabelValues(value);
    }
  };

  const addLabelFilter = () => {
    setLabelFilters([...labelFilters, { label: '', operator: '=', value: '' }]);
  };

  const removeLabelFilter = (index: number) => {
    setLabelFilters(labelFilters.filter((_, i) => i !== index));
  };

  const onCodeChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange({ ...currentQuery, expr: event.target.value });
  };

  const onLegendChange = (value: string) => {
    onChange({ ...currentQuery, legendFormat: value });
  };

  const onFormatChange = (value: SelectableValue<string>) => {
    onChange({ ...currentQuery, format: value.value as 'time_series' | 'table' | 'heatmap' });
  };

  const onTypeChange = (value: string) => {
    const isInstant = value === 'instant';
    const isRange = value === 'range';
    onChange({
      ...currentQuery,
      instant: isInstant || value === 'both',
      range: isRange || value === 'both',
    });
  };

  const onStepChange = (value: string) => {
    onChange({ ...currentQuery, interval: value });
  };

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
      onRunQuery();
    }
  };

  const getTypeValue = () => {
    if (currentQuery.instant && currentQuery.range) return 'both';
    if (currentQuery.instant) return 'instant';
    return 'range';
  };

  return (
    <div className={styles.container}>
      {/* Top Row */}
      <div className={styles.topRow}>
        <div className={styles.leftSection}>
          <Button
            variant="secondary"
            size="sm"
            className={styles.kickstartButton}
            onClick={() => {
              loadMetrics();
              loadLabels();
            }}
          >
            Kick start your query
          </Button>
          <InlineField label="Explain" transparent>
            <InlineSwitch value={false} onChange={() => {}} disabled />
          </InlineField>
        </div>
        <RadioButtonGroup
          options={[
            { label: 'Builder', value: 'builder' as const },
            { label: 'Code', value: 'code' as const },
          ]}
          value={editorMode}
          onChange={(v) => setEditorMode(v as 'builder' | 'code')}
          size="sm"
        />
      </div>

      {editorMode === 'builder' ? (
        <>
          {/* Builder Mode */}
          <div className={styles.builderRow}>
            {/* Metric Section */}
            <div className={styles.metricSection}>
              <span className={styles.sectionLabel}>Metric</span>
              <Select
                options={metrics}
                value={selectedMetric ? { label: selectedMetric, value: selectedMetric } : { label: 'Metrics explorer', value: '' }}
                onChange={onMetricChange}
                isLoading={metricsLoading}
                placeholder="Metrics explorer"
                width={25}
                allowCustomValue
                onCreateOption={(v) => setSelectedMetric(v)}
                isClearable
              />
            </div>

            {/* Label Filters Section */}
            <div className={styles.labelSection}>
              <span className={styles.sectionLabel}>Label filters</span>
              <div className={styles.labelFilters}>
                {labelFilters.map((filter, index) => (
                  <div key={index} className={styles.labelFilter}>
                    <Select
                      options={labels}
                      value={filter.label ? { label: filter.label, value: filter.label } : null}
                      onChange={(v) => onLabelFilterChange(index, 'label', v.value || '')}
                      isLoading={labelsLoading}
                      placeholder="Select label"
                      width={20}
                      allowCustomValue
                      onCreateOption={(v) => onLabelFilterChange(index, 'label', v)}
                    />
                    <Select
                      options={operatorOptions}
                      value={operatorOptions.find((o) => o.value === filter.operator)}
                      onChange={(v) => onLabelFilterChange(index, 'operator', v.value || '=')}
                      width={8}
                    />
                    <Select
                      options={labelValues[filter.label] || []}
                      value={filter.value ? { label: filter.value, value: filter.value } : null}
                      onChange={(v) => onLabelFilterChange(index, 'value', v.value || '')}
                      placeholder="Select value"
                      width={20}
                      allowCustomValue
                      onCreateOption={(v) => onLabelFilterChange(index, 'value', v)}
                      onOpenMenu={() => filter.label && loadLabelValues(filter.label)}
                    />
                    <IconButton
                      name="times"
                      size="sm"
                      onClick={() => removeLabelFilter(index)}
                      tooltip="Remove filter"
                    />
                  </div>
                ))}
                <IconButton
                  name="plus"
                  size="md"
                  onClick={addLabelFilter}
                  tooltip="Add label filter"
                  variant="secondary"
                />
              </div>
            </div>
          </div>

          {/* Operations Button */}
          <div className={styles.operationsButton}>
            <Button variant="secondary" size="sm" icon="plus" disabled>
              Operations
            </Button>
          </div>
        </>
      ) : (
        /* Code Mode */
        <textarea
          className={styles.codeEditor}
          value={currentQuery.expr || ''}
          onChange={onCodeChange}
          onBlur={onRunQuery}
          onKeyDown={handleKeyDown}
          placeholder="Enter a PromQL expression (e.g., up, rate(http_requests_total[5m]))"
          rows={3}
        />
      )}

      {/* Options Section */}
      <Collapse
        label={
          <div className={styles.optionsRow}>
            <span style={{ fontWeight: 500 }}>Options</span>
            <span className={styles.optionItem}>
              <span className={styles.optionLabel}>Legend:</span> {currentQuery.legendFormat || 'Auto'}
            </span>
            <span className={styles.optionItem}>
              <span className={styles.optionLabel}>Format:</span> Time series
            </span>
            <span className={styles.optionItem}>
              <span className={styles.optionLabel}>Step:</span> {currentQuery.interval || 'auto'}
            </span>
            <span className={styles.optionItem}>
              <span className={styles.optionLabel}>Type:</span> {getTypeValue() === 'both' ? 'Both' : getTypeValue() === 'instant' ? 'Instant' : 'Range'}
            </span>
          </div>
        }
        isOpen={optionsOpen}
        onToggle={() => setOptionsOpen(!optionsOpen)}
        collapsible
      >
        <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', padding: '8px 0' }}>
          <InlineField label="Legend" labelWidth={12}>
            <Input
              width={25}
              value={currentQuery.legendFormat || ''}
              onChange={(e) => onLegendChange(e.currentTarget.value)}
              placeholder="Auto"
            />
          </InlineField>
          <InlineField label="Format" labelWidth={12}>
            <Select
              options={formatOptions}
              value={formatOptions.find((o) => o.value === (currentQuery.format || 'time_series'))}
              onChange={onFormatChange}
              width={20}
            />
          </InlineField>
          <InlineField label="Step" labelWidth={12}>
            <Input
              width={15}
              value={currentQuery.interval || ''}
              onChange={(e) => onStepChange(e.currentTarget.value)}
              placeholder="auto"
            />
          </InlineField>
          <InlineField label="Type" labelWidth={12}>
            <RadioButtonGroup
              options={typeOptions.map((o) => ({ label: o.label!, value: o.value! }))}
              value={getTypeValue()}
              onChange={onTypeChange}
              size="sm"
            />
          </InlineField>
          <InlineField label="Exemplars" labelWidth={12}>
            <InlineSwitch value={false} onChange={() => {}} disabled />
          </InlineField>
        </div>
      </Collapse>
    </div>
  );
}
