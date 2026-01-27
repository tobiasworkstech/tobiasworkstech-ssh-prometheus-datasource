import { DataQuery, DataSourceJsonData } from '@grafana/data';

export type AuthMethod = 'password' | 'key';
export type PrometheusAuthMethod = 'none' | 'basic' | 'bearer';

export interface SSHPrometheusQuery extends DataQuery {
  expr: string;
  legendFormat?: string;
  instant?: boolean;
  range?: boolean;
  interval?: string;
  format?: 'time_series' | 'table' | 'heatmap';
}

export const defaultQuery: Partial<SSHPrometheusQuery> = {
  expr: '',
  legendFormat: '',
  instant: false,
  range: true,
};

export interface SSHPrometheusDataSourceOptions extends DataSourceJsonData {
  // SSH Connection
  sshHost: string;
  sshPort: number;
  sshUsername: string;
  authMethod: AuthMethod;

  // Prometheus Connection
  prometheusUrl: string;

  // Prometheus Authentication
  prometheusAuthMethod: PrometheusAuthMethod;
  prometheusUsername?: string;

  // TLS Settings
  tlsSkipVerify?: boolean;
  tlsWithCACert?: boolean;
  tlsWithClientCert?: boolean;

  // HTTP Settings
  httpMethod?: 'GET' | 'POST';
  customQueryParameters?: string;

  // Timeouts
  timeout?: number;
}

export interface SSHPrometheusSecureJsonData {
  // SSH secrets
  sshPassword?: string;
  sshPrivateKey?: string;
  sshKeyPassphrase?: string;

  // Prometheus secrets
  prometheusPassword?: string;
  prometheusBearerToken?: string;

  // TLS secrets
  tlsCACert?: string;
  tlsClientCert?: string;
  tlsClientKey?: string;
}
