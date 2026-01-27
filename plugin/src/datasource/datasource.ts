import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  MutableDataFrame,
  FieldType,
  MetricFindValue,
} from '@grafana/data';
import { getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import { SSHPrometheusDataSourceOptions, SSHPrometheusQuery, defaultQuery } from '../types';

export class DataSource extends DataSourceApi<SSHPrometheusQuery, SSHPrometheusDataSourceOptions> {
  resourceUrl: string;
  dsUid: string;

  constructor(instanceSettings: DataSourceInstanceSettings<SSHPrometheusDataSourceOptions>) {
    super(instanceSettings);
    this.dsUid = instanceSettings.uid;
    // Use resources endpoint for backend plugin calls
    this.resourceUrl = `/api/datasources/uid/${instanceSettings.uid}/resources`;
  }

  async query(options: DataQueryRequest<SSHPrometheusQuery>): Promise<DataQueryResponse> {
    const { range } = options;
    const from = range!.from.valueOf();
    const to = range!.to.valueOf();

    const promises = options.targets
      .filter((target) => !target.hide && target.expr)
      .map(async (target) => {
        const query = { ...defaultQuery, ...target };
        const expr = getTemplateSrv().replace(query.expr, options.scopedVars);

        const params: Record<string, string | number> = {
          query: expr,
        };

        let endpoint = 'query';
        if (query.range && !query.instant) {
          endpoint = 'query_range';
          params.start = Math.floor(from / 1000);
          params.end = Math.floor(to / 1000);
          params.step = this.calculateStep(from, to, options.maxDataPoints || 1000, query.interval);
        } else {
          params.time = Math.floor(to / 1000);
        }

        const response = await getBackendSrv().post(`${this.resourceUrl}/api/v1/${endpoint}`, params);

        return this.transformResponse(response, query);
      });

    const data = await Promise.all(promises);
    return { data: data.flat() };
  }

  private calculateStep(from: number, to: number, maxDataPoints: number, interval?: string): number {
    if (interval) {
      const parsed = this.parseInterval(interval);
      if (parsed) {
        return parsed;
      }
    }
    const rangeSeconds = (to - from) / 1000;
    return Math.max(Math.floor(rangeSeconds / maxDataPoints), 1);
  }

  private parseInterval(interval: string): number | null {
    const match = interval.match(/^(\d+)(s|m|h|d)$/);
    if (!match) {
      return null;
    }
    const value = parseInt(match[1], 10);
    const unit = match[2];
    switch (unit) {
      case 's':
        return value;
      case 'm':
        return value * 60;
      case 'h':
        return value * 3600;
      case 'd':
        return value * 86400;
      default:
        return null;
    }
  }

  private transformResponse(response: any, query: SSHPrometheusQuery): MutableDataFrame[] {
    if (!response?.data?.result) {
      return [];
    }

    const resultType = response.data.resultType;

    return response.data.result.map((result: any) => {
      const labels = result.metric || {};
      const legendFormat = query.legendFormat || '';
      const name = this.formatLegend(labels, legendFormat);

      const frame = new MutableDataFrame({
        refId: query.refId,
        name,
        fields: [
          { name: 'Time', type: FieldType.time },
          { name: 'Value', type: FieldType.number, labels },
        ],
      });

      if (resultType === 'matrix') {
        for (const [timestamp, value] of result.values || []) {
          frame.appendRow([timestamp * 1000, parseFloat(value)]);
        }
      } else if (resultType === 'vector') {
        const [timestamp, value] = result.value || [];
        if (timestamp !== undefined) {
          frame.appendRow([timestamp * 1000, parseFloat(value)]);
        }
      }

      return frame;
    });
  }

  private formatLegend(labels: Record<string, string>, format: string): string {
    if (!format) {
      return Object.entries(labels)
        .map(([k, v]) => `${k}="${v}"`)
        .join(', ') || 'value';
    }

    let result = format;
    for (const [key, value] of Object.entries(labels)) {
      result = result.replace(new RegExp(`\\{\\{\\s*${key}\\s*\\}\\}`, 'g'), value);
    }
    return result;
  }

  async metricFindQuery(query: string, options?: any): Promise<MetricFindValue[]> {
    const interpolated = getTemplateSrv().replace(query, options?.scopedVars);

    // Handle label_values function
    const labelValuesMatch = interpolated.match(/^label_values\((?:(.+),\s*)?(.+)\)$/);
    if (labelValuesMatch) {
      const metric = labelValuesMatch[1];
      const label = labelValuesMatch[2].trim();
      return this.getLabelValues(label, metric);
    }

    // Handle label_names function
    if (interpolated === 'label_names()') {
      return this.getLabels();
    }

    // Handle metrics function
    const metricsMatch = interpolated.match(/^metrics\((.+)?\)$/);
    if (metricsMatch) {
      const filter = metricsMatch[1];
      return this.getMetrics(filter);
    }

    // Default: treat as PromQL query
    const response = await getBackendSrv().post(`${this.resourceUrl}/api/v1/query`, {
      query: interpolated,
      time: Math.floor(Date.now() / 1000),
    });

    if (!response?.data?.result) {
      return [];
    }

    return response.data.result.map((r: any) => ({
      text: Object.entries(r.metric || {}).map(([k, v]) => `${k}="${v}"`).join(', ') || r.value?.[1],
    }));
  }

  private async getLabels(): Promise<MetricFindValue[]> {
    const response = await getBackendSrv().get(`${this.resourceUrl}/api/v1/labels`);
    if (!response?.data) {
      return [];
    }
    return response.data.map((label: string) => ({ text: label }));
  }

  private async getLabelValues(label: string, metric?: string): Promise<MetricFindValue[]> {
    let url = `${this.resourceUrl}/api/v1/label/${encodeURIComponent(label)}/values`;
    if (metric) {
      url += `?match[]=${encodeURIComponent(metric)}`;
    }
    const response = await getBackendSrv().get(url);
    if (!response?.data) {
      return [];
    }
    return response.data.map((value: string) => ({ text: value }));
  }

  private async getMetrics(filter?: string): Promise<MetricFindValue[]> {
    const response = await getBackendSrv().get(`${this.resourceUrl}/api/v1/label/__name__/values`);
    if (!response?.data) {
      return [];
    }
    let metrics = response.data as string[];
    if (filter) {
      const regex = new RegExp(filter);
      metrics = metrics.filter((m: string) => regex.test(m));
    }
    return metrics.map((m: string) => ({ text: m }));
  }

  async testDatasource(): Promise<{ status: string; message: string }> {
    // For backend plugins, call Grafana's built-in health check endpoint
    // which invokes the backend's CheckHealth handler
    try {
      const response = await getBackendSrv().get(`/api/datasources/uid/${this.dsUid}/health`);
      // Status can be "ok" or "OK" depending on Grafana version
      if (response?.status?.toLowerCase() === 'ok') {
        return { status: 'success', message: response.message || 'Data source is working' };
      }
      return { status: 'error', message: response?.message || 'Unknown error' };
    } catch (error: any) {
      const message = error?.data?.message || error?.message || 'Failed to connect';
      return { status: 'error', message };
    }
  }

  getDefaultQuery(): Partial<SSHPrometheusQuery> {
    return defaultQuery;
  }
}
