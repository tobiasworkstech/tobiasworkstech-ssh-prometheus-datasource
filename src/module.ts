import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource/datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { SSHPrometheusQuery, SSHPrometheusDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, SSHPrometheusQuery, SSHPrometheusDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
